package repositories

import (
	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"

	"github.com/google/uuid"
)

type CaseTimelineRepository struct{}

func NewCaseTimelineRepository() *CaseTimelineRepository {
	return &CaseTimelineRepository{}
}

func (r *CaseTimelineRepository) Create(e *models.CaseTimelineEvent) error {
	return database.DB.Create(e).Error
}

func (r *CaseTimelineRepository) ListByCase(caseID, userID uuid.UUID) ([]models.CaseTimelineEvent, error) {
	var events []models.CaseTimelineEvent
	if err := database.DB.Where("case_id = ? AND user_id = ?", caseID, userID).
		Order("event_date ASC").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *CaseTimelineRepository) FindByID(id, userID uuid.UUID) (*models.CaseTimelineEvent, error) {
	var e models.CaseTimelineEvent
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *CaseTimelineRepository) Update(e *models.CaseTimelineEvent) error {
	return database.DB.Save(e).Error
}
