package repositories

import (
	"time"

	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"

	"github.com/google/uuid"
)

type MeetingRepository struct{}

func NewMeetingRepository() *MeetingRepository {
	return &MeetingRepository{}
}

func (r *MeetingRepository) Create(m *models.Meeting) error {
	return database.DB.Create(m).Error
}

func (r *MeetingRepository) FindByID(id, userID uuid.UUID) (*models.Meeting, error) {
	var m models.Meeting
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MeetingRepository) ListUpcoming(userID uuid.UUID) ([]models.Meeting, error) {
	var meetings []models.Meeting
	if err := database.DB.Where("user_id = ? AND scheduled_at >= ? AND status != ?", userID, time.Now(), models.MeetingStatusCancelled).
		Order("scheduled_at ASC").Find(&meetings).Error; err != nil {
		return nil, err
	}
	return meetings, nil
}

func (r *MeetingRepository) ListHistory(userID uuid.UUID) ([]models.Meeting, error) {
	var meetings []models.Meeting
	if err := database.DB.Where("user_id = ? AND (scheduled_at < ? OR status = ?)", userID, time.Now(), models.MeetingStatusCompleted).
		Order("scheduled_at DESC").Find(&meetings).Error; err != nil {
		return nil, err
	}
	return meetings, nil
}

func (r *MeetingRepository) ListByClient(userID, clientID uuid.UUID) ([]models.Meeting, error) {
	var meetings []models.Meeting
	if err := database.DB.Where("user_id = ? AND client_id = ?", userID, clientID).
		Order("scheduled_at DESC").Find(&meetings).Error; err != nil {
		return nil, err
	}
	return meetings, nil
}

// ListByClientID is used by the client portal — the client viewing their own
// meetings, resolved server-side from their authenticated account.
func (r *MeetingRepository) ListByClientID(clientID uuid.UUID) ([]models.Meeting, error) {
	var meetings []models.Meeting
	if err := database.DB.Where("client_id = ?", clientID).
		Order("scheduled_at DESC").Find(&meetings).Error; err != nil {
		return nil, err
	}
	return meetings, nil
}

func (r *MeetingRepository) Update(m *models.Meeting) error {
	return database.DB.Save(m).Error
}

func (r *MeetingRepository) Delete(id, userID uuid.UUID) error {
	return database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Meeting{}).Error
}
