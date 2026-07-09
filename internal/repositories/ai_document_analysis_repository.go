package repositories

import (
	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"

	"github.com/google/uuid"
)

type AIDocumentAnalysisRepository struct{}

func NewAIDocumentAnalysisRepository() *AIDocumentAnalysisRepository {
	return &AIDocumentAnalysisRepository{}
}

func (r *AIDocumentAnalysisRepository) Create(a *models.AIDocumentAnalysis) error {
	return database.DB.Create(a).Error
}

func (r *AIDocumentAnalysisRepository) ListByUser(userID uuid.UUID, analysisType string) ([]models.AIDocumentAnalysis, error) {
	var results []models.AIDocumentAnalysis
	q := database.DB.Where("user_id = ?", userID)
	if analysisType != "" {
		q = q.Where("type = ?", analysisType)
	}
	if err := q.Order("created_at DESC").Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}
