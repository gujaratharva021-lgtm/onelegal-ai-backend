package repositories

import (
	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"
)

// AdminRepository is the only place in the codebase that queries across
// every lawyer's data at once — every other repository scopes by owner
// (user_id/lawyer_id). Reachable only through admin-role-gated routes.
type AdminRepository struct{}

func NewAdminRepository() *AdminRepository {
	return &AdminRepository{}
}

func (r *AdminRepository) CountLawyers() (int64, error) {
	var count int64
	err := database.DB.Model(&models.User{}).Where("role = ?", models.RoleAdvocate).Count(&count).Error
	return count, err
}

func (r *AdminRepository) CountClients() (int64, error) {
	var count int64
	err := database.DB.Model(&models.Client{}).Count(&count).Error
	return count, err
}

func (r *AdminRepository) CountCases() (int64, error) {
	var count int64
	err := database.DB.Model(&models.Case{}).Count(&count).Error
	return count, err
}

func (r *AdminRepository) CountMeetings() (int64, error) {
	var count int64
	err := database.DB.Model(&models.Meeting{}).Count(&count).Error
	return count, err
}

// CountPendingPayments counts invoices that have been issued but not yet
// confirmed paid — "sent" (pending) and "submitted" (payment initiated,
// awaiting the lawyer's manual verification).
func (r *AdminRepository) CountPendingPayments() (int64, error) {
	var count int64
	err := database.DB.Model(&models.Invoice{}).
		Where("status IN ?", []string{string(models.InvoiceStatusSent), string(models.InvoiceStatusSubmitted)}).
		Count(&count).Error
	return count, err
}

// SumRevenuePaise sums only invoices the lawyer has manually confirmed Paid
// — never a client-reported UPI result alone.
func (r *AdminRepository) SumRevenuePaise() (int64, error) {
	var total int64
	err := database.DB.Model(&models.Invoice{}).
		Where("status = ?", models.InvoiceStatusPaid).
		Select("COALESCE(SUM(amount_paise), 0)").
		Scan(&total).Error
	return total, err
}

func (r *AdminRepository) ListLawyers() ([]models.AdminLawyerSummary, error) {
	var lawyers []models.User
	if err := database.DB.Where("role = ?", models.RoleAdvocate).Order("created_at DESC").Find(&lawyers).Error; err != nil {
		return nil, err
	}

	summaries := make([]models.AdminLawyerSummary, 0, len(lawyers))
	for _, l := range lawyers {
		var clientCount, caseCount int64
		database.DB.Model(&models.Client{}).Where("user_id = ?", l.ID).Count(&clientCount)
		database.DB.Model(&models.Case{}).Where("user_id = ?", l.ID).Count(&caseCount)
		summaries = append(summaries, models.AdminLawyerSummary{
			ID:          l.ID.String(),
			Name:        l.Name,
			Email:       l.Email,
			Phone:       l.Phone,
			LawFirm:     l.LawFirm,
			IsOnline:    l.IsOnline,
			ClientCount: clientCount,
			CaseCount:   caseCount,
			CreatedAt:   l.CreatedAt,
		})
	}
	return summaries, nil
}

func (r *AdminRepository) ListClients() ([]models.AdminClientSummary, error) {
	var clients []models.Client
	if err := database.DB.Order("created_at DESC").Find(&clients).Error; err != nil {
		return nil, err
	}

	summaries := make([]models.AdminClientSummary, 0, len(clients))
	for _, cl := range clients {
		var lawyer models.User
		lawyerName := ""
		if err := database.DB.First(&lawyer, "id = ?", cl.UserID).Error; err == nil {
			lawyerName = lawyer.Name
		}
		summaries = append(summaries, models.AdminClientSummary{
			ID:            cl.ID.String(),
			Name:          cl.Name,
			Email:         cl.Email,
			Phone:         cl.Phone,
			Status:        string(cl.Status),
			AccountStatus: string(cl.AccountStatus),
			LawyerName:    lawyerName,
			CreatedAt:     cl.CreatedAt,
		})
	}
	return summaries, nil
}

func (r *AdminRepository) ListCases() ([]models.AdminCaseSummary, error) {
	var cases []models.Case
	if err := database.DB.Order("created_at DESC").Find(&cases).Error; err != nil {
		return nil, err
	}

	summaries := make([]models.AdminCaseSummary, 0, len(cases))
	for _, c := range cases {
		var lawyer models.User
		lawyerName := ""
		if err := database.DB.First(&lawyer, "id = ?", c.UserID).Error; err == nil {
			lawyerName = lawyer.Name
		}
		clientName := ""
		if c.ClientID != nil {
			var client models.Client
			if err := database.DB.First(&client, "id = ?", *c.ClientID).Error; err == nil {
				clientName = client.Name
			}
		}
		summaries = append(summaries, models.AdminCaseSummary{
			ID:         c.ID.String(),
			Title:      c.Title,
			CaseNumber: c.CaseNumber,
			Status:     string(c.Status),
			Priority:   string(c.Priority),
			LawyerName: lawyerName,
			ClientName: clientName,
			CreatedAt:  c.CreatedAt,
		})
	}
	return summaries, nil
}

func (r *AdminRepository) ListPayments() ([]models.AdminPaymentSummary, error) {
	var invoices []models.Invoice
	if err := database.DB.Order("created_at DESC").Find(&invoices).Error; err != nil {
		return nil, err
	}

	summaries := make([]models.AdminPaymentSummary, 0, len(invoices))
	for _, inv := range invoices {
		var lawyer models.User
		lawyerName := ""
		if err := database.DB.First(&lawyer, "id = ?", inv.LawyerID).Error; err == nil {
			lawyerName = lawyer.Name
		}
		var client models.Client
		clientName := ""
		if err := database.DB.First(&client, "id = ?", inv.ClientID).Error; err == nil {
			clientName = client.Name
		}
		summaries = append(summaries, models.AdminPaymentSummary{
			ID:            inv.ID.String(),
			InvoiceNumber: inv.InvoiceNumber,
			AmountPaise:   inv.AmountPaise,
			Status:        string(inv.Status),
			LawyerName:    lawyerName,
			ClientName:    clientName,
			PaymentDate:   inv.PaymentDate,
			CreatedAt:     inv.CreatedAt,
		})
	}
	return summaries, nil
}
