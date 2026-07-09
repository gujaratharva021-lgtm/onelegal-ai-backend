package services

import (
	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"

	"github.com/google/uuid"
)

type NoteService struct {
	repo *repositories.NoteRepository
}

func NewNoteService() *NoteService {
	return &NoteService{repo: repositories.NewNoteRepository()}
}

func (s *NoteService) Create(userID uuid.UUID, req models.NoteRequest) (*models.Note, error) {
	category := req.Category
	if category == "" {
		category = models.NoteCategoryPersonal
	}
	note := models.Note{
		UserID:   userID,
		CaseID:   req.CaseID,
		Title:    req.Title,
		Content:  req.Content,
		Category: category,
	}
	if err := s.repo.Create(&note); err != nil {
		return nil, err
	}
	return &note, nil
}

func (s *NoteService) List(userID uuid.UUID, category, search string) ([]models.Note, error) {
	return s.repo.ListByUser(userID, category, search)
}

func (s *NoteService) Get(id, userID uuid.UUID) (*models.Note, error) {
	return s.repo.FindByID(id, userID)
}

func (s *NoteService) Update(id, userID uuid.UUID, req models.NoteRequest) (*models.Note, error) {
	note, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, err
	}
	note.CaseID = req.CaseID
	note.Title = req.Title
	note.Content = req.Content
	if req.Category != "" {
		note.Category = req.Category
	}
	if err := s.repo.Update(note); err != nil {
		return nil, err
	}
	return note, nil
}

func (s *NoteService) Delete(id, userID uuid.UUID) error {
	return s.repo.Delete(id, userID)
}
