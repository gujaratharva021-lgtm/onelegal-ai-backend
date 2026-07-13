package repositories

import (
	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"

	"github.com/google/uuid"
)

type DocumentRepository struct{}

func NewDocumentRepository() *DocumentRepository {
	return &DocumentRepository{}
}

func (r *DocumentRepository) Create(d *models.Document) error {
	return database.DB.Create(d).Error
}

func (r *DocumentRepository) FindByID(id, userID uuid.UUID) (*models.Document, error) {
	var d models.Document
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DocumentRepository) ListByCase(caseID, userID uuid.UUID) ([]models.Document, error) {
	var docs []models.Document
	if err := database.DB.Where("case_id = ? AND user_id = ?", caseID, userID).
		Order("created_at DESC").Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *DocumentRepository) ListByUser(userID uuid.UUID, caseID string) ([]models.Document, error) {
	var docs []models.Document
	q := database.DB.Where("user_id = ?", userID)
	if caseID != "" {
		q = q.Where("case_id = ?", caseID)
	}
	if err := q.Order("created_at DESC").Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

// ListByClient covers documents attached directly to the client as well as
// documents attached to any of the client's cases, so the client profile's
// Documents tab shows everything relevant regardless of how it was uploaded.
func (r *DocumentRepository) ListByClient(userID, clientID uuid.UUID) ([]models.Document, error) {
	var docs []models.Document
	if err := database.DB.Where(
		"user_id = ? AND (client_id = ? OR case_id IN (SELECT id FROM cases WHERE client_id = ? AND user_id = ?))",
		userID, clientID, clientID, userID,
	).Order("created_at DESC").Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

// ListByClientID is used by the client portal — the client viewing their own
// documents, resolved server-side from their authenticated account rather
// than any lawyer id.
func (r *DocumentRepository) ListByClientID(clientID uuid.UUID) ([]models.Document, error) {
	var docs []models.Document
	if err := database.DB.Where(
		"client_id = ? OR case_id IN (SELECT id FROM cases WHERE client_id = ?)",
		clientID, clientID,
	).Order("created_at DESC").Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *DocumentRepository) Update(d *models.Document) error {
	return database.DB.Save(d).Error
}

func (r *DocumentRepository) Delete(id, userID uuid.UUID) error {
	return database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Document{}).Error
}
