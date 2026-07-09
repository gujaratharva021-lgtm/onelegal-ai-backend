package repositories

import (
	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"

	"github.com/google/uuid"
)

type ResearchRepository struct{}

func NewResearchRepository() *ResearchRepository {
	return &ResearchRepository{}
}

func (r *ResearchRepository) Create(rh *models.ResearchHistory) error {
	return database.DB.Create(rh).Error
}

func (r *ResearchRepository) FindByID(id, userID uuid.UUID) (*models.ResearchHistory, error) {
	var rh models.ResearchHistory
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&rh).Error; err != nil {
		return nil, err
	}
	return &rh, nil
}

func (r *ResearchRepository) ListByUser(userID uuid.UUID, bookmarkedOnly bool) ([]models.ResearchHistory, error) {
	var history []models.ResearchHistory
	q := database.DB.Where("user_id = ?", userID)
	if bookmarkedOnly {
		q = q.Where("bookmarked = true")
	}
	if err := q.Order("created_at DESC").Find(&history).Error; err != nil {
		return nil, err
	}
	return history, nil
}

func (r *ResearchRepository) ListRecent(userID uuid.UUID, limit int) ([]models.ResearchHistory, error) {
	var history []models.ResearchHistory
	if err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&history).Error; err != nil {
		return nil, err
	}
	return history, nil
}

func (r *ResearchRepository) UpdateBookmark(id, userID uuid.UUID, bookmarked bool) error {
	return database.DB.Model(&models.ResearchHistory{}).Where("id = ? AND user_id = ?", id, userID).Update("bookmarked", bookmarked).Error
}

func (r *ResearchRepository) Delete(id, userID uuid.UUID) error {
	return database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.ResearchHistory{}).Error
}
