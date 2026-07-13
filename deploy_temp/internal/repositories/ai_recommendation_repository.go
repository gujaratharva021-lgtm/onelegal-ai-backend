package repositories

import (
	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"

	"github.com/google/uuid"
)

type AIRecommendationRepository struct{}

func NewAIRecommendationRepository() *AIRecommendationRepository {
	return &AIRecommendationRepository{}
}

func (r *AIRecommendationRepository) Create(rec *models.AIRecommendation) error {
	return database.DB.Create(rec).Error
}

func (r *AIRecommendationRepository) ListByCase(caseID, userID uuid.UUID) ([]models.AIRecommendation, error) {
	var results []models.AIRecommendation
	if err := database.DB.Where("case_id = ? AND user_id = ?", caseID, userID).
		Order("created_at DESC").Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}
