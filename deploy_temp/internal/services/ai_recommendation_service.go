package services

import (
	"fmt"
	"strings"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"

	"github.com/google/uuid"
)

type AIRecommendationService struct {
	repo      *repositories.AIRecommendationRepository
	caseRepo  *repositories.CaseRepository
	aiService *AIService
}

func NewAIRecommendationService(aiService *AIService) *AIRecommendationService {
	return &AIRecommendationService{
		repo:      repositories.NewAIRecommendationRepository(),
		caseRepo:  repositories.NewCaseRepository(),
		aiService: aiService,
	}
}

func buildCaseSummary(c *models.Case) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Case Title: %s\n", c.Title)
	if c.CaseNumber != "" {
		fmt.Fprintf(&b, "Case Number: %s\n", c.CaseNumber)
	}
	if c.CourtName != "" {
		fmt.Fprintf(&b, "Court: %s\n", c.CourtName)
	}
	if c.Judge != "" {
		fmt.Fprintf(&b, "Judge: %s\n", c.Judge)
	}
	if c.Opponent != "" {
		fmt.Fprintf(&b, "Opponent: %s\n", c.Opponent)
	}
	if c.CaseType != "" {
		fmt.Fprintf(&b, "Case Type: %s\n", c.CaseType)
	}
	fmt.Fprintf(&b, "Priority: %s\n", c.Priority)
	fmt.Fprintf(&b, "Status: %s\n", c.Status)
	if c.Description != "" {
		fmt.Fprintf(&b, "Description: %s\n", c.Description)
	}
	return b.String()
}

func (s *AIRecommendationService) Generate(userID, caseID uuid.UUID) (*models.AIRecommendation, error) {
	c, err := s.caseRepo.FindByID(caseID, userID)
	if err != nil {
		return nil, err
	}

	text, err := s.aiService.RecommendForCase(buildCaseSummary(c))
	if err != nil {
		return nil, err
	}

	rec := &models.AIRecommendation{
		UserID:         userID,
		CaseID:         caseID,
		Recommendation: text,
	}
	if err := s.repo.Create(rec); err != nil {
		return nil, err
	}
	return rec, nil
}

func (s *AIRecommendationService) List(caseID, userID uuid.UUID) ([]models.AIRecommendation, error) {
	if _, err := s.caseRepo.FindByID(caseID, userID); err != nil {
		return nil, err
	}
	return s.repo.ListByCase(caseID, userID)
}
