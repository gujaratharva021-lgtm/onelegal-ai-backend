package services

import (
	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"

	"github.com/google/uuid"
)

type CaseTimelineService struct {
	repo     *repositories.CaseTimelineRepository
	caseRepo *repositories.CaseRepository
}

func NewCaseTimelineService() *CaseTimelineService {
	return &CaseTimelineService{
		repo:     repositories.NewCaseTimelineRepository(),
		caseRepo: repositories.NewCaseRepository(),
	}
}

func (s *CaseTimelineService) Create(caseID, userID uuid.UUID, req models.CaseTimelineEventRequest) (*models.CaseTimelineEvent, error) {
	if _, err := s.caseRepo.FindByID(caseID, userID); err != nil {
		return nil, err
	}
	event := models.CaseTimelineEvent{
		CaseID:    caseID,
		UserID:    userID,
		EventDate: req.EventDate,
		Title:     req.Title,
		Notes:     req.Notes,
	}
	if err := s.repo.Create(&event); err != nil {
		return nil, err
	}
	return &event, nil
}

func (s *CaseTimelineService) List(caseID, userID uuid.UUID) ([]models.CaseTimelineEvent, error) {
	if _, err := s.caseRepo.FindByID(caseID, userID); err != nil {
		return nil, err
	}
	return s.repo.ListByCase(caseID, userID)
}

func (s *CaseTimelineService) Update(id, userID uuid.UUID, req models.CaseTimelineEventRequest) (*models.CaseTimelineEvent, error) {
	event, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, err
	}
	event.EventDate = req.EventDate
	event.Title = req.Title
	event.Notes = req.Notes
	if err := s.repo.Update(event); err != nil {
		return nil, err
	}
	return event, nil
}
