package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VideoRoomStatus is the room lifecycle: Ringing (created, pushed to the
// client, not yet answered) -> Active (client joined, WebRTC connecting/
// connected) -> Ended. A room sitting in Ringing is NOT a "busy" call — the
// client hasn't picked up yet, so a second call attempt must still be
// allowed to go through; only a genuinely Active (answered) call makes the
// client busy. This distinction is the fix for the false "busy" bug: the
// previous two-state model (active/ended) marked a room "active" the
// instant it was created, so any call nobody ever answered stayed "active"
// forever and permanently blocked every future call attempt.
type VideoRoomStatus string

const (
	VideoRoomStatusRinging VideoRoomStatus = "ringing"
	VideoRoomStatusActive  VideoRoomStatus = "active"
	VideoRoomStatusEnded   VideoRoomStatus = "ended"
)

// VideoRoomOutcome is the final call state, set once the room ends. Empty
// while the room is still active.
type VideoRoomOutcome string

const (
	VideoRoomOutcomeCompleted VideoRoomOutcome = "completed" // client joined, then either side hung up
	VideoRoomOutcomeRejected  VideoRoomOutcome = "rejected"  // client declined before joining
	VideoRoomOutcomeMissed    VideoRoomOutcome = "missed"    // client never responded (ringing timed out)
	VideoRoomOutcomeCancelled VideoRoomOutcome = "cancelled" // lawyer ended it before the client answered
)

// VideoRoom is a strictly private, exactly-two-participant WebRTC call
// session between one lawyer and one client. It's a live-call concept,
// distinct from Meeting (the scheduling/calendar entity) — a room is
// created the moment the lawyer presses "Start Meeting"/"Video Call" and
// closed the moment either side ends the call. This is the app's single
// call-log table (no separate "video_calls" table — extending this one
// avoids a duplicate parallel model).
type VideoRoom struct {
	ID           uuid.UUID       `json:"room_id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	MeetingID    *uuid.UUID      `json:"meeting_id" gorm:"type:uuid;index"`
	CaseID       *uuid.UUID      `json:"case_id" gorm:"type:uuid;index"`
	LawyerID     uuid.UUID       `json:"lawyer_id" gorm:"type:uuid;not null;index"`
	ClientID     uuid.UUID       `json:"client_id" gorm:"type:uuid;not null;index"`
	ClientUserID uuid.UUID       `json:"client_user_id" gorm:"type:uuid;not null;index"`
	Status       VideoRoomStatus `json:"status" gorm:"default:ringing;index"`
	// Outcome/StartedAt/DurationSeconds are all populated at Join/End time —
	// see VideoRoomService.Join/End.
	Outcome         VideoRoomOutcome `json:"outcome"`
	StartedAt       *time.Time       `json:"started_at"`
	DurationSeconds int              `json:"duration_seconds"`
	CreatedAt       time.Time        `json:"created_at"`
	EndedAt         *time.Time       `json:"ended_at"`
	DeletedAt       gorm.DeletedAt   `json:"-" gorm:"index"`
}

type StartMeetingCallRequest struct {
	MeetingID uuid.UUID `json:"meeting_id" binding:"required"`
}

// EndCallRequest is optional — Reason lets the caller state WHY the call is
// ending before it was ever accepted (client declining = "rejected", an
// unanswered ring timing out client-side = "missed"). Omitted/ignored once
// the call has actually been accepted, since then it's always "completed".
type EndCallRequest struct {
	Reason string `json:"reason"`
}
