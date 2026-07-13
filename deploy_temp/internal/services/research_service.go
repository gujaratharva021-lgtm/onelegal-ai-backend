package services

import (
	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"

	"github.com/google/uuid"
)

type ResearchService struct {
	repo      *repositories.ResearchRepository
	aiService *AIService
}

func NewResearchService(aiService *AIService) *ResearchService {
	return &ResearchService{repo: repositories.NewResearchRepository(), aiService: aiService}
}

func (s *ResearchService) Search(userID uuid.UUID, req models.ResearchRequest) (*models.ResearchHistory, error) {
	summary, err := s.aiService.LegalResearch(req.Query, string(req.Category))
	if err != nil {
		return nil, err
	}

	entry := models.ResearchHistory{
		UserID:        userID,
		Query:         req.Query,
		Category:      req.Category,
		ResultSummary: summary,
	}
	if err := s.repo.Create(&entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func (s *ResearchService) List(userID uuid.UUID, bookmarkedOnly bool) ([]models.ResearchHistory, error) {
	return s.repo.ListByUser(userID, bookmarkedOnly)
}

func (s *ResearchService) ListRecent(userID uuid.UUID, limit int) ([]models.ResearchHistory, error) {
	return s.repo.ListRecent(userID, limit)
}

func (s *ResearchService) SetBookmark(id, userID uuid.UUID, bookmarked bool) error {
	return s.repo.UpdateBookmark(id, userID, bookmarked)
}

func (s *ResearchService) Delete(id, userID uuid.UUID) error {
	return s.repo.Delete(id, userID)
}
