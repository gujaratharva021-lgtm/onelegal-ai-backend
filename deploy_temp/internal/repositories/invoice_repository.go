package repositories

import (
	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"

	"github.com/google/uuid"
)

type InvoiceRepository struct{}

func NewInvoiceRepository() *InvoiceRepository {
	return &InvoiceRepository{}
}

func (r *InvoiceRepository) Create(inv *models.Invoice) error {
	return database.DB.Create(inv).Error
}

// FindByID is lawyer-scoped — used by the issuing lawyer's own endpoints.
func (r *InvoiceRepository) FindByID(id, lawyerID uuid.UUID) (*models.Invoice, error) {
	var inv models.Invoice
	if err := database.DB.Where("id = ? AND lawyer_id = ?", id, lawyerID).First(&inv).Error; err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *InvoiceRepository) ListByLawyer(lawyerID uuid.UUID) ([]models.Invoice, error) {
	var invoices []models.Invoice
	if err := database.DB.Where("lawyer_id = ?", lawyerID).Order("created_at DESC").Find(&invoices).Error; err != nil {
		return nil, err
	}
	return invoices, nil
}

func (r *InvoiceRepository) ListByLawyerAndClient(lawyerID, clientID uuid.UUID) ([]models.Invoice, error) {
	var invoices []models.Invoice
	if err := database.DB.Where("lawyer_id = ? AND client_id = ?", lawyerID, clientID).
		Order("created_at DESC").Find(&invoices).Error; err != nil {
		return nil, err
	}
	return invoices, nil
}

// ListByClientID is used by the client portal — the client viewing their own
// invoices, resolved server-side from their authenticated account rather
// than any lawyer id. A draft is the lawyer's own unfinished work-in-progress
// (see InvoiceService.Send) — it never appears to the client until sent.
func (r *InvoiceRepository) ListByClientID(clientID uuid.UUID) ([]models.Invoice, error) {
	var invoices []models.Invoice
	if err := database.DB.Where("client_id = ? AND status != ?", clientID, models.InvoiceStatusDraft).
		Order("created_at DESC").Find(&invoices).Error; err != nil {
		return nil, err
	}
	return invoices, nil
}

// FindByIDForClient is client-scoped — used by the client's own Pay Now /
// invoice-detail actions, where the caller is the client, not the lawyer.
// Excludes drafts for the same reason as ListByClientID above.
func (r *InvoiceRepository) FindByIDForClient(id, clientID uuid.UUID) (*models.Invoice, error) {
	var inv models.Invoice
	if err := database.DB.Where("id = ? AND client_id = ? AND status != ?", id, clientID, models.InvoiceStatusDraft).
		First(&inv).Error; err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *InvoiceRepository) Update(inv *models.Invoice) error {
	return database.DB.Save(inv).Error
}

func (r *InvoiceRepository) DeleteDraft(id, lawyerID uuid.UUID) error {
	return database.DB.Where("id = ? AND lawyer_id = ? AND status = ?", id, lawyerID, models.InvoiceStatusDraft).
		Delete(&models.Invoice{}).Error
}
