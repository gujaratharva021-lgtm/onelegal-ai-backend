package repositories

import (
	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"

	"github.com/google/uuid"
)

type PaymentRepository struct{}

func NewPaymentRepository() *PaymentRepository {
	return &PaymentRepository{}
}

func (r *PaymentRepository) Create(p *models.Payment) error {
	return database.DB.Create(p).Error
}

func (r *PaymentRepository) FindByInvoiceID(invoiceID uuid.UUID) (*models.Payment, error) {
	var p models.Payment
	if err := database.DB.Where("invoice_id = ?", invoiceID).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepository) Update(p *models.Payment) error {
	return database.DB.Save(p).Error
}

// ListByLawyer is the lawyer's full Payment History across every client.
func (r *PaymentRepository) ListByLawyer(lawyerID uuid.UUID) ([]models.Payment, error) {
	var payments []models.Payment
	if err := database.DB.Where("lawyer_id = ?", lawyerID).
		Order("created_at DESC").Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

// ListByClientID is the client's own Payment History.
func (r *PaymentRepository) ListByClientID(clientID uuid.UUID) ([]models.Payment, error) {
	var payments []models.Payment
	if err := database.DB.Where("client_id = ?", clientID).
		Order("created_at DESC").Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}
