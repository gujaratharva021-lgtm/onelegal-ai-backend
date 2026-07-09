package repositories

import (
	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"

	"github.com/google/uuid"
)

type NoteRepository struct{}

func NewNoteRepository() *NoteRepository {
	return &NoteRepository{}
}

func (r *NoteRepository) Create(n *models.Note) error {
	return database.DB.Create(n).Error
}

func (r *NoteRepository) FindByID(id, userID uuid.UUID) (*models.Note, error) {
	var n models.Note
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&n).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NoteRepository) ListByUser(userID uuid.UUID, category, search string) ([]models.Note, error) {
	var notes []models.Note
	q := database.DB.Where("user_id = ?", userID)
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if search != "" {
		q = q.Where("title ILIKE ? OR content ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if err := q.Order("updated_at DESC").Find(&notes).Error; err != nil {
		return nil, err
	}
	return notes, nil
}

func (r *NoteRepository) Update(n *models.Note) error {
	return database.DB.Save(n).Error
}

func (r *NoteRepository) Delete(id, userID uuid.UUID) error {
	return database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Note{}).Error
}
