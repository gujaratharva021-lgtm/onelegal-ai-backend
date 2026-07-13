package services

import (
	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"
)

// AdminService backs the hidden Admin Dashboard — every figure and list it
// returns is a live query against the same tables the lawyer/client apps
// use, never mock/hardcoded data. Reachable only via routes gated by
// middleware.RequireRole(models.RoleAdmin).
type AdminService struct {
	repo *repositories.AdminRepository
}

func NewAdminService() *AdminService {
	return &AdminService{repo: repositories.NewAdminRepository()}
}

func (s *AdminService) DashboardStats() (*models.AdminDashboardStats, error) {
	lawyers, err := s.repo.CountLawyers()
	if err != nil {
		return nil, err
	}
	clients, err := s.repo.CountClients()
	if err != nil {
		return nil, err
	}
	cases, err := s.repo.CountCases()
	if err != nil {
		return nil, err
	}
	meetings, err := s.repo.CountMeetings()
	if err != nil {
		return nil, err
	}
	pendingPayments, err := s.repo.CountPendingPayments()
	if err != nil {
		return nil, err
	}
	revenue, err := s.repo.SumRevenuePaise()
	if err != nil {
		return nil, err
	}

	return &models.AdminDashboardStats{
		TotalLawyers:    lawyers,
		TotalClients:    clients,
		TotalCases:      cases,
		TotalMeetings:   meetings,
		PendingPayments: pendingPayments,
		RevenuePaise:    revenue,
	}, nil
}

func (s *AdminService) ListLawyers() ([]models.AdminLawyerSummary, error) {
	return s.repo.ListLawyers()
}

func (s *AdminService) ListClients() ([]models.AdminClientSummary, error) {
	return s.repo.ListClients()
}

func (s *AdminService) ListCases() ([]models.AdminCaseSummary, error) {
	return s.repo.ListCases()
}

func (s *AdminService) ListPayments() ([]models.AdminPaymentSummary, error) {
	return s.repo.ListPayments()
}
