package services

import (
	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"

	"github.com/google/uuid"
)

type HearingService struct {
	repo *repositories.HearingRepository
}

func NewHearingService() *HearingService {
	return &HearingService{repo: repositories.NewHearingRepository()}
}

func (s *HearingService) Create(userID uuid.UUID, req models.HearingRequest) (*models.Hearing, error) {
	status := req.Status
	if status == "" {
		status = models.HearingStatusUpcoming
	}
	hearing := models.Hearing{
		UserID:      userID,
		CaseID:      req.CaseID,
		CourtName:   req.CourtName,
		CourtNumber: req.CourtNumber,
		Judge:       req.Judge,
		HearingDate: req.HearingDate,
		HearingType: req.HearingType,
		Status:      status,
		Notes:       req.Notes,
	}
	if err := s.repo.Create(&hearing); err != nil {
		return nil, err
	}
	return &hearing, nil
}

func (s *HearingService) List(userID uuid.UUID, status string) ([]models.Hearing, error) {
	return s.repo.ListByUser(userID, status)
}

func (s *HearingService) ListToday(userID uuid.UUID) ([]models.Hearing, error) {
	return s.repo.ListToday(userID)
}

func (s *HearingService) Get(id, userID uuid.UUID) (*models.Hearing, error) {
	return s.repo.FindByID(id, userID)
}

func (s *HearingService) Update(id, userID uuid.UUID, req models.HearingRequest) (*models.Hearing, error) {
	hearing, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, err
	}
	hearing.CaseID = req.CaseID
	hearing.CourtName = req.CourtName
	hearing.CourtNumber = req.CourtNumber
	hearing.Judge = req.Judge
	hearing.HearingDate = req.HearingDate
	hearing.HearingType = req.HearingType
	if req.Status != "" {
		hearing.Status = req.Status
	}
	hearing.Notes = req.Notes
	if err := s.repo.Update(hearing); err != nil {
		return nil, err
	}
	return hearing, nil
}

func (s *HearingService) Delete(id, userID uuid.UUID) error {
	return s.repo.Delete(id, userID)
}
