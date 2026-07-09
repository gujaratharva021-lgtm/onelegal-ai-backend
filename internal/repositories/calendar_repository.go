package repositories

import (
	"time"

	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"

	"github.com/google/uuid"
)

type CalendarRepository struct{}

func NewCalendarRepository() *CalendarRepository {
	return &CalendarRepository{}
}

func (r *CalendarRepository) Create(e *models.CalendarEvent) error {
	return database.DB.Create(e).Error
}

func (r *CalendarRepository) FindByID(id, userID uuid.UUID) (*models.CalendarEvent, error) {
	var e models.CalendarEvent
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *CalendarRepository) ListByRange(userID uuid.UUID, from, to time.Time) ([]models.CalendarEvent, error) {
	var events []models.CalendarEvent
	if err := database.DB.Where("user_id = ? AND start_time >= ? AND start_time <= ?", userID, from, to).
		Order("start_time ASC").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *CalendarRepository) Update(e *models.CalendarEvent) error {
	return database.DB.Save(e).Error
}

func (r *CalendarRepository) Delete(id, userID uuid.UUID) error {
	return database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.CalendarEvent{}).Error
}
