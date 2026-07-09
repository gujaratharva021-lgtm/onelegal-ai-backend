package services

import (
	"errors"
	"strings"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"
	"legaltech-backend/internal/utils"

	"github.com/google/uuid"
)

var (
	ErrDuplicateEmail            = errors.New("a client with this email already exists")
	ErrDuplicatePhone            = errors.New("a client with this phone number already exists")
	ErrLoginIDRequired           = errors.New("client login ID is required")
	ErrTemporaryPasswordRequired = errors.New("temporary password is required")
	ErrLoginIDTaken              = errors.New("this login ID is already in use by another account")
	ErrNoLinkedAccount           = errors.New("this client has no login account to reset")
)

type ClientService struct {
	repo                *repositories.ClientRepository
	userRepo            *repositories.UserRepository
	notificationService *NotificationService
}

func NewClientService() *ClientService {
	return &ClientService{
		repo:                repositories.NewClientRepository(),
		userRepo:            repositories.NewUserRepository(),
		notificationService: NewNotificationService(),
	}
}

func (s *ClientService) checkDuplicate(userID uuid.UUID, email, phone string, excludeID *uuid.UUID) error {
	existing, _ := s.repo.FindDuplicate(userID, email, phone, excludeID)
	if existing == nil {
		return nil
	}
	if email != "" && existing.Email == email {
		return ErrDuplicateEmail
	}
	if phone != "" && existing.Phone == phone {
		return ErrDuplicatePhone
	}
	return nil
}

func normalizeClientStatus(status string) models.ClientStatus {
	if strings.EqualFold(strings.TrimSpace(status), string(models.ClientStatusClosed)) {
		return models.ClientStatusClosed
	}
	return models.ClientStatusActive
}

// Create saves the CRM record and, in the same step, creates the client's
// login account (Role = RoleClient) using the lawyer-supplied Login ID and
// temporary password, then links the two. There is no public client signup
// — this is the only place a client-role User is ever created.
func (s *ClientService) Create(userID uuid.UUID, req models.ClientRequest) (*models.Client, error) {
	if err := s.checkDuplicate(userID, req.Email, req.Phone, nil); err != nil {
		return nil, err
	}

	loginID := strings.TrimSpace(req.LoginID)
	if loginID == "" {
		return nil, ErrLoginIDRequired
	}
	if strings.TrimSpace(req.TemporaryPassword) == "" {
		return nil, ErrTemporaryPasswordRequired
	}
	if _, err := s.userRepo.FindByEmail(loginID); err == nil {
		return nil, ErrLoginIDTaken
	}

	hashedPassword, err := utils.HashPassword(req.TemporaryPassword)
	if err != nil {
		return nil, errors.New("failed to hash temporary password")
	}

	accountUser := models.User{
		Name:               req.Name,
		Email:              loginID,
		Password:           hashedPassword,
		Phone:              req.Phone,
		Role:               models.RoleClient,
		MustChangePassword: true,
	}
	if err := s.userRepo.Create(&accountUser); err != nil {
		return nil, errors.New("failed to create client login account")
	}

	client := models.Client{
		UserID:        userID,
		Name:          req.Name,
		Email:         req.Email,
		Phone:         req.Phone,
		Address:       req.Address,
		City:          req.City,
		State:         req.State,
		DateOfBirth:   req.DateOfBirth,
		Gender:        req.Gender,
		CaseType:      req.CaseType,
		Notes:         req.Notes,
		Status:        normalizeClientStatus(req.Status),
		LoginID:       loginID,
		AccountUserID: &accountUser.ID,
		AccountStatus: models.ClientAccountActive,
	}
	if err := s.repo.Create(&client); err != nil {
		return nil, err
	}

	_, _ = s.notificationService.Create(userID, models.NotificationRequest{
		Title: "Client account created successfully",
		Body:  client.Name + " can now log in with the login ID you provided.",
		Type:  models.NotificationTypeGeneral,
	})

	return &client, nil
}

// ResetPassword lets the owning lawyer set a new temporary password for a
// client's existing login account — the client's Login ID never changes.
func (s *ClientService) ResetPassword(id, userID uuid.UUID, newPassword string) (*models.Client, error) {
	client, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, err
	}
	if client.AccountUserID == nil {
		return nil, ErrNoLinkedAccount
	}
	account, err := s.userRepo.FindByID(*client.AccountUserID)
	if err != nil {
		return nil, ErrNoLinkedAccount
	}
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return nil, errors.New("failed to hash new password")
	}
	account.Password = hashedPassword
	account.MustChangePassword = true
	if err := s.userRepo.Update(account); err != nil {
		return nil, errors.New("failed to update password")
	}

	_, _ = s.notificationService.Create(account.ID, models.NotificationRequest{
		Title: "Your password has been reset",
		Body:  "Your advocate has set a new temporary password for your account.",
		Type:  models.NotificationTypeGeneral,
	})

	return client, nil
}

func (s *ClientService) List(userID uuid.UUID, search string) ([]models.Client, error) {
	return s.repo.ListByUser(userID, search)
}

func (s *ClientService) Get(id, userID uuid.UUID) (*models.Client, error) {
	return s.repo.FindByID(id, userID)
}

func (s *ClientService) Update(id, userID uuid.UUID, req models.ClientRequest) (*models.Client, error) {
	client, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, err
	}
	if err := s.checkDuplicate(userID, req.Email, req.Phone, &id); err != nil {
		return nil, err
	}
	client.Name = req.Name
	client.Email = req.Email
	client.Phone = req.Phone
	client.Address = req.Address
	client.City = req.City
	client.State = req.State
	client.DateOfBirth = req.DateOfBirth
	client.Gender = req.Gender
	client.CaseType = req.CaseType
	client.Notes = req.Notes
	if req.Status != "" {
		client.Status = normalizeClientStatus(req.Status)
	}
	if err := s.repo.Update(client); err != nil {
		return nil, err
	}
	return client, nil
}

// Archive marks a client Closed without soft-deleting the row — the client
// and their linked cases/documents/meetings/invoices remain fully visible
// and accessible, just flagged as no longer an active engagement.
func (s *ClientService) Archive(id, userID uuid.UUID) (*models.Client, error) {
	client, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, err
	}
	client.Status = models.ClientStatusClosed
	if err := s.repo.Update(client); err != nil {
		return nil, err
	}
	return client, nil
}

func (s *ClientService) Delete(id, userID uuid.UUID) error {
	return s.repo.Delete(id, userID)
}
