package repositories

import (
	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"

	"github.com/google/uuid"
)

type VideoRoomRepository struct{}

func NewVideoRoomRepository() *VideoRoomRepository {
	return &VideoRoomRepository{}
}

func (r *VideoRoomRepository) Create(room *models.VideoRoom) error {
	return database.DB.Create(room).Error
}

func (r *VideoRoomRepository) FindByID(id uuid.UUID) (*models.VideoRoom, error) {
	var room models.VideoRoom
	if err := database.DB.Where("id = ?", id).First(&room).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

// FindOpenByMeetingID reports the current not-yet-ended room (ringing or
// active) for a meeting, if any — used to make StartCall idempotent against
// duplicate/concurrent "start" requests for the same meeting.
func (r *VideoRoomRepository) FindOpenByMeetingID(meetingID uuid.UUID) (*models.VideoRoom, error) {
	var room models.VideoRoom
	if err := database.DB.
		Where("meeting_id = ? AND status IN ?", meetingID, []models.VideoRoomStatus{models.VideoRoomStatusRinging, models.VideoRoomStatusActive}).
		Order("created_at DESC").
		First(&room).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

// FindOpenForClientUser is the "does this client have a call ringing or in
// progress right now" check performed when the client logs in / opens the
// app, so they can discover a call they might have missed the WS push for.
func (r *VideoRoomRepository) FindOpenForClientUser(clientUserID uuid.UUID) (*models.VideoRoom, error) {
	var room models.VideoRoom
	if err := database.DB.
		Where("client_user_id = ? AND status IN ?", clientUserID, []models.VideoRoomStatus{models.VideoRoomStatusRinging, models.VideoRoomStatusActive}).
		Order("created_at DESC").
		First(&room).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

// FindOngoingForClientUser reports a call the client is genuinely ANSWERED
// and connected on right now (Status == Active only — NOT merely ringing).
// This is the sole definition of "busy": an unanswered ringing call must
// never block a new call attempt.
func (r *VideoRoomRepository) FindOngoingForClientUser(clientUserID uuid.UUID) (*models.VideoRoom, error) {
	var room models.VideoRoom
	if err := database.DB.
		Where("client_user_id = ? AND status = ?", clientUserID, models.VideoRoomStatusActive).
		Order("created_at DESC").
		First(&room).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *VideoRoomRepository) Update(room *models.VideoRoom) error {
	return database.DB.Save(room).Error
}
