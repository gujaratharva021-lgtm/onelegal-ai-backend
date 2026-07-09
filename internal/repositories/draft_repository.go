package repositories

import (
	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"

	"github.com/google/uuid"
)

type DraftRepository struct{}

func NewDraftRepository() *DraftRepository {
	return &DraftRepository{}
}

func (r *DraftRepository) Create(d *models.Draft) error {
	return database.DB.Create(d).Error
}

func (r *DraftRepository) FindByID(id, userID uuid.UUID) (*models.Draft, error) {
	var d models.Draft
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DraftRepository) ListByUser(userID uuid.UUID, draftType string) ([]models.Draft, error) {
	var drafts []models.Draft
	q := database.DB.Where("user_id = ?", userID)
	if draftType != "" {
		q = q.Where("draft_type = ?", draftType)
	}
	if err := q.Order("updated_at DESC").Find(&drafts).Error; err != nil {
		return nil, err
	}
	return drafts, nil
}

func (r *DraftRepository) Update(d *models.Draft) error {
	return database.DB.Save(d).Error
}

func (r *DraftRepository) Delete(id, userID uuid.UUID) error {
	return database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Draft{}).Error
}
