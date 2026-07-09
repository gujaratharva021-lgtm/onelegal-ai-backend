package services

import (
	"fmt"
	"strings"
	"time"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"
	"legaltech-backend/internal/utils"

	"github.com/google/uuid"
)

type MeetingService struct {
	repo *repositories.MeetingRepository
}

func NewMeetingService() *MeetingService {
	return &MeetingService{repo: repositories.NewMeetingRepository()}
}

func (s *MeetingService) Create(userID uuid.UUID, req models.MeetingRequest) (*models.Meeting, error) {
	roomSuffix, err := utils.GenerateSecureToken(6)
	if err != nil {
		return nil, err
	}
	roomID := fmt.Sprintf("legaltechai-%s", strings.ToLower(strings.ReplaceAll(roomSuffix, "_", "")))

	duration := req.DurationMinutes
	if duration <= 0 {
		duration = 30
	}
	mode := req.Mode
	if mode == "" {
		mode = models.MeetingModeVideo
	}

	meeting := models.Meeting{
		UserID:          userID,
		CaseID:          req.CaseID,
		ClientID:        req.ClientID,
		Title:           req.Title,
		RoomID:          roomID,
		ScheduledAt:     req.ScheduledAt,
		ReminderAt:      req.ReminderAt,
		DurationMinutes: duration,
		Mode:            mode,
		Participants:    strings.Join(req.Participants, ","),
		Notes:           req.Notes,
		Status:          models.MeetingStatusScheduled,
	}
	if err := s.repo.Create(&meeting); err != nil {
		return nil, err
	}
	return &meeting, nil
}

func (s *MeetingService) Update(id, userID uuid.UUID, req models.MeetingRequest) (*models.Meeting, error) {
	meeting, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, err
	}
	meeting.CaseID = req.CaseID
	meeting.ClientID = req.ClientID
	meeting.Title = req.Title
	meeting.ScheduledAt = req.ScheduledAt
	meeting.ReminderAt = req.ReminderAt
	if req.DurationMinutes > 0 {
		meeting.DurationMinutes = req.DurationMinutes
	}
	if req.Participants != nil {
		meeting.Participants = strings.Join(req.Participants, ",")
	}
	if req.Mode != "" {
		meeting.Mode = req.Mode
	}
	meeting.Notes = req.Notes
	if err := s.repo.Update(meeting); err != nil {
		return nil, err
	}
	return meeting, nil
}

// Cancel is a distinct, non-destructive alternative to Delete — the meeting
// row (and its history) is preserved with Status = Cancelled rather than
// removed outright.
func (s *MeetingService) Cancel(id, userID uuid.UUID) (*models.Meeting, error) {
	return s.UpdateStatus(id, userID, models.MeetingStatusCancelled)
}

func (s *MeetingService) ListUpcoming(userID uuid.UUID) ([]models.Meeting, error) {
	return s.repo.ListUpcoming(userID)
}

func (s *MeetingService) ListHistory(userID uuid.UUID) ([]models.Meeting, error) {
	return s.repo.ListHistory(userID)
}

func (s *MeetingService) ListForClient(userID, clientID uuid.UUID) ([]models.Meeting, error) {
	return s.repo.ListByClient(userID, clientID)
}

func (s *MeetingService) Get(id, userID uuid.UUID) (*models.Meeting, error) {
	return s.repo.FindByID(id, userID)
}

func (s *MeetingService) UpdateStatus(id, userID uuid.UUID, status models.MeetingStatus) (*models.Meeting, error) {
	meeting, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, err
	}
	meeting.Status = status
	if err := s.repo.Update(meeting); err != nil {
		return nil, err
	}
	return meeting, nil
}

func (s *MeetingService) Delete(id, userID uuid.UUID) error {
	return s.repo.Delete(id, userID)
}

var _ = time.Now
