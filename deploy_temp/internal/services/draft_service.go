package services

import (
	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"

	"github.com/google/uuid"
)

type DraftService struct {
	repo      *repositories.DraftRepository
	aiService *AIService
}

func NewDraftService(aiService *AIService) *DraftService {
	return &DraftService{repo: repositories.NewDraftRepository(), aiService: aiService}
}

func (s *DraftService) Create(userID uuid.UUID, req models.DraftRequest) (*models.Draft, error) {
	status := req.Status
	if status == "" {
		status = models.DraftStatusDraft
	}
	draft := models.Draft{
		UserID:    userID,
		CaseID:    req.CaseID,
		ClientID:  req.ClientID,
		Title:     req.Title,
		DraftType: req.DraftType,
		Content:   req.Content,
		Status:    status,
	}
	if err := s.repo.Create(&draft); err != nil {
		return nil, err
	}
	return &draft, nil
}

func (s *DraftService) GenerateWithAI(userID uuid.UUID, req models.AIDraftGenerateRequest) (*models.Draft, error) {
	content, err := s.aiService.GenerateDraft(req.DraftType, req.Title, req.Instructions)
	if err != nil {
		return nil, err
	}
	draft := models.Draft{
		UserID:    userID,
		Title:     req.Title,
		DraftType: req.DraftType,
		Content:   content,
		Status:    models.DraftStatusDraft,
	}
	if err := s.repo.Create(&draft); err != nil {
		return nil, err
	}
	return &draft, nil
}

func (s *DraftService) List(userID uuid.UUID, draftType string) ([]models.Draft, error) {
	return s.repo.ListByUser(userID, draftType)
}

func (s *DraftService) Get(id, userID uuid.UUID) (*models.Draft, error) {
	return s.repo.FindByID(id, userID)
}

func (s *DraftService) Update(id, userID uuid.UUID, req models.DraftRequest) (*models.Draft, error) {
	draft, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, err
	}
	draft.CaseID = req.CaseID
	draft.ClientID = req.ClientID
	draft.Title = req.Title
	if req.DraftType != "" {
		draft.DraftType = req.DraftType
	}
	draft.Content = req.Content
	if req.Status != "" {
		draft.Status = req.Status
	}
	if err := s.repo.Update(draft); err != nil {
		return nil, err
	}
	return draft, nil
}

func (s *DraftService) Delete(id, userID uuid.UUID) error {
	return s.repo.Delete(id, userID)
}
