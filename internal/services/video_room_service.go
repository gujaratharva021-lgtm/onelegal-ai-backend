package services

import (
	"errors"
	"sync"
	"time"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"
	"legaltech-backend/internal/ws"

	"github.com/google/uuid"
)

// startCallMu serializes StartCall end-to-end (check-for-existing-room
// through Create+Register). Without it, two concurrent "start" requests for
// the same meeting (double-tap, double device) can both pass the
// no-active-room check before either creates one, producing two independent
// VideoRoom rows/RoomRegistry entries for a single meeting. A single global
// lock is deliberately simple here — call-start volume is inherently low
// (one lawyer manually starting one call at a time), so serializing all
// starts costs nothing observable.
var startCallMu sync.Mutex

// VideoRoomService implements a strictly private, exactly-two-participant
// (lawyer + client) WebRTC call — no email, no share link, no SMTP. A
// "client" here means an existing, already-registered app User whose email
// matches the lawyer's Client CRM record; if no such account exists yet,
// starting the call is rejected outright rather than falling back to any
// other invitation channel.
type VideoRoomService struct {
	repo        *repositories.VideoRoomRepository
	meetingRepo *repositories.MeetingRepository
	clientRepo  *repositories.ClientRepository
	userRepo    *repositories.UserRepository
	hub         *ws.Hub
	rooms       *ws.RoomRegistry
}

func NewVideoRoomService(hub *ws.Hub, rooms *ws.RoomRegistry) *VideoRoomService {
	return &VideoRoomService{
		repo:        repositories.NewVideoRoomRepository(),
		meetingRepo: repositories.NewMeetingRepository(),
		clientRepo:  repositories.NewClientRepository(),
		userRepo:    repositories.NewUserRepository(),
		hub:         hub,
		rooms:       rooms,
	}
}

// StartCall is lawyer-only. It resolves the meeting's assigned client to a
// real, already-registered user account, opens a two-person room, and
// pushes a real-time "video-call-incoming" signal if the client is online.
func (s *VideoRoomService) StartCall(lawyerID, meetingID uuid.UUID) (*models.VideoRoom, string, bool, error) {
	startCallMu.Lock()
	defer startCallMu.Unlock()

	meeting, err := s.meetingRepo.FindByID(meetingID, lawyerID)
	if err != nil {
		return nil, "", false, errors.New("meeting not found")
	}
	if meeting.ClientID == nil {
		return nil, "", false, errors.New("this meeting has no client assigned")
	}

	// Idempotent: a double-tap or a race between two requests for the same
	// meeting rejoins the already-open (ringing or active) room instead of
	// creating a second, orphaned one.
	if existing, err := s.repo.FindOpenByMeetingID(meetingID); err == nil {
		clientName := ""
		online := false
		if client, err := s.clientRepo.FindByID(*meeting.ClientID, lawyerID); err == nil {
			clientName = client.Name
			if client.AccountUserID != nil {
				if clientUser, err := s.userRepo.FindByID(*client.AccountUserID); err == nil {
					online = clientUser.IsOnline || s.hub.IsOnline(clientUser.ID)
				}
			}
		}
		return existing, clientName, online, nil
	}

	client, err := s.clientRepo.FindByID(*meeting.ClientID, lawyerID)
	if err != nil {
		return nil, "", false, errors.New("client not found")
	}
	// Blocks calling a client whose account has been explicitly deactivated
	// (all their cases closed). Legacy clients created before the account
	// system existed have an empty AccountStatus, not "inactive", so this
	// never affects them — they keep resolving by email below as before.
	if client.AccountStatus == models.ClientAccountInactive {
		return nil, "", false, errors.New("this client's account is inactive — their case has been closed")
	}

	// Prefer the durable linked account (set at client creation); fall back
	// to a live email lookup for clients created before that link existed.
	var clientUser *models.User
	if client.AccountUserID != nil {
		clientUser, err = s.userRepo.FindByID(*client.AccountUserID)
	}
	if clientUser == nil {
		if client.Email == "" {
			return nil, "", false, errors.New("this client has no email on file — add one to enable video calls")
		}
		clientUser, err = s.userRepo.FindByEmail(client.Email)
		if err != nil {
			return nil, "", false, errors.New("this client does not have a LegalTech AI account yet — they must sign up with " + client.Email + " before you can call them")
		}
	}
	if clientUser.ID == lawyerID {
		return nil, "", false, errors.New("client account cannot be the same as your own account")
	}

	// Busy ONLY means a call the client has actually ANSWERED is still in
	// progress (Status == Active) for a DIFFERENT meeting — an unanswered
	// ringing call must never block a new attempt (FindOpenByMeetingID
	// above already handled the same-meeting idempotency case).
	if busy, err := s.repo.FindOngoingForClientUser(clientUser.ID); err == nil && busy.MeetingID != nil && *busy.MeetingID != meetingID {
		return nil, "", false, errors.New("this client is busy on another call right now")
	}

	room := &models.VideoRoom{
		MeetingID:    &meetingID,
		CaseID:       meeting.CaseID,
		LawyerID:     lawyerID,
		ClientID:     client.ID,
		ClientUserID: clientUser.ID,
		Status:       models.VideoRoomStatusRinging,
	}
	if err := s.repo.Create(room); err != nil {
		return nil, "", false, err
	}

	// Registered BEFORE the push so the client's very first signaling
	// message (if they're already connected) is never rejected by the room
	// membership check in the signaling handler.
	s.rooms.Register(room.ID.String(), lawyerID, clientUser.ID)

	lawyer, _ := s.userRepo.FindByID(lawyerID)
	lawyerName := ""
	if lawyer != nil {
		lawyerName = lawyer.Name
	}

	delivered := s.hub.SendJSON(clientUser.ID, map[string]interface{}{
		"type":          "video-call-incoming",
		"call_id":       room.ID,
		"from":          lawyerID,
		"from_name":     lawyerName,
		"to":            clientUser.ID,
		"meeting_title": meeting.Title,
	})

	// "Online" means the client has an authenticated session (IsOnline, set
	// at login/logout — see AuthService.setOnline) OR the real-time push
	// above actually reached a live socket. A WebSocket merely not being
	// connected at this exact instant (e.g. app briefly backgrounded on
	// Android) must never by itself report the client as offline — that was
	// the bug: SendJSON's success/failure was previously the ONLY signal
	// used, so a normal background/reconnect blip looked identical to the
	// client never having logged in at all.
	online := clientUser.IsOnline || delivered

	return room, client.Name, online, nil
}

// ActiveForClient is the check performed when the client opens/logs into
// the app: "is there a call ringing or in progress for me right now".
func (s *VideoRoomService) ActiveForClient(clientUserID uuid.UUID) (*models.VideoRoom, string, error) {
	room, err := s.repo.FindOpenForClientUser(clientUserID)
	if err != nil {
		return nil, "", err
	}
	lawyer, err := s.userRepo.FindByID(room.LawyerID)
	if err != nil {
		return nil, "", err
	}
	return room, lawyer.Name, nil
}

// Join is client-only — only the room's exact assigned client account may
// ever join; every other authenticated user is rejected outright.
func (s *VideoRoomService) Join(roomID, clientUserID uuid.UUID) (*models.VideoRoom, string, error) {
	room, err := s.repo.FindByID(roomID)
	if err != nil {
		return nil, "", errors.New("meeting room not found")
	}
	if room.Status == models.VideoRoomStatusEnded {
		return nil, "", errors.New("this meeting has already ended")
	}
	if room.ClientUserID != clientUserID {
		return nil, "", errors.New("you are not authorized to join this meeting")
	}
	// Ringing -> Active: this is the moment the call actually becomes
	// "answered"/busy from the client's side.
	if room.Status == models.VideoRoomStatusRinging {
		room.Status = models.VideoRoomStatusActive
	}
	if room.StartedAt == nil {
		now := time.Now()
		room.StartedAt = &now
	}
	if err := s.repo.Update(room); err != nil {
		return nil, "", err
	}
	lawyer, err := s.userRepo.FindByID(room.LawyerID)
	if err != nil {
		return nil, "", err
	}
	return room, lawyer.Name, nil
}

// End is callable by either the lawyer or the assigned client — anyone else
// is rejected. Closing the room also revokes the signaling registry entry,
// so no further offer/answer/ICE messages for this call_id are relayed to
// anyone, ever. reason is only consulted for a call that was never
// accepted (StartedAt == nil) — a completed call is always "completed".
func (s *VideoRoomService) End(roomID, userID uuid.UUID, reason string) (*models.VideoRoom, error) {
	room, err := s.repo.FindByID(roomID)
	if err != nil {
		return nil, errors.New("meeting room not found")
	}
	if userID != room.LawyerID && userID != room.ClientUserID {
		return nil, errors.New("you are not authorized to end this meeting")
	}
	if room.Status != models.VideoRoomStatusEnded {
		now := time.Now()
		room.Status = models.VideoRoomStatusEnded
		room.EndedAt = &now

		if room.StartedAt != nil {
			room.Outcome = models.VideoRoomOutcomeCompleted
			room.DurationSeconds = int(now.Sub(*room.StartedAt).Seconds())
		} else {
			switch {
			case reason == string(models.VideoRoomOutcomeRejected) && userID == room.ClientUserID:
				room.Outcome = models.VideoRoomOutcomeRejected
			case reason == string(models.VideoRoomOutcomeMissed):
				room.Outcome = models.VideoRoomOutcomeMissed
			default:
				room.Outcome = models.VideoRoomOutcomeCancelled
			}
		}
		if err := s.repo.Update(room); err != nil {
			return nil, err
		}
		s.syncMeetingStatus(room)
	}
	s.rooms.Unregister(room.ID.String())
	return room, nil
}

// syncMeetingStatus reflects the call's outcome onto the Meeting it was
// started from (if any) — reusing the existing Meeting module as the call
// history surface (Client Profile → Meetings) rather than a separate log.
func (s *VideoRoomService) syncMeetingStatus(room *models.VideoRoom) {
	if room.MeetingID == nil {
		return
	}
	meeting, err := s.meetingRepo.FindByID(*room.MeetingID, room.LawyerID)
	if err != nil {
		return
	}
	if room.Outcome == models.VideoRoomOutcomeCompleted {
		meeting.Status = models.MeetingStatusCompleted
	} else {
		meeting.Status = models.MeetingStatusCancelled
	}
	_ = s.meetingRepo.Update(meeting)
}
