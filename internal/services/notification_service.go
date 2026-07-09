package services

import (
	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"

	"github.com/google/uuid"
)

type NotificationService struct {
	repo     *repositories.NotificationRepository
	userRepo *repositories.UserRepository
}

func NewNotificationService() *NotificationService {
	return &NotificationService{
		repo:     repositories.NewNotificationRepository(),
		userRepo: repositories.NewUserRepository(),
	}
}

func (s *NotificationService) List(userID uuid.UUID, unreadOnly bool) (interface{}, error) {
	return s.repo.ListByUser(userID, unreadOnly)
}

func (s *NotificationService) MarkRead(id, userID uuid.UUID) error {
	return s.repo.MarkRead(id, userID)
}

func (s *NotificationService) MarkAllRead(userID uuid.UUID) error {
	return s.repo.MarkAllRead(userID)
}

func (s *NotificationService) CountUnread(userID uuid.UUID) (int64, error) {
	return s.repo.CountUnread(userID)
}

// Create registers a reminder-style notification (e.g. "30 min before this
// hearing"). Flutter also schedules a matching local device notification;
// this record is the source of truth so reminders survive reinstall/backup.
func (s *NotificationService) Create(userID uuid.UUID, req models.NotificationRequest) (*models.Notification, error) {
	notifType := req.Type
	if notifType == "" {
		notifType = models.NotificationTypeGeneral
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	n := &models.Notification{
		UserID:       userID,
		Title:        req.Title,
		Body:         req.Body,
		Type:         notifType,
		RelatedID:    req.RelatedID,
		ScheduledFor: req.ScheduledFor,
		Enabled:      enabled,
	}
	if err := s.repo.Create(n); err != nil {
		return nil, err
	}

	// Best-effort real push alongside the in-app bell notification — every
	// existing call site (client account created, password reset, case
	// closed, user-created reminders) gets push "for free" through this one
	// integration point. Silently does nothing if Firebase isn't configured
	// or the user has no registered device token.
	if globalPushService != nil {
		if user, err := s.userRepo.FindByID(userID); err == nil && user.DeviceToken != "" {
			globalPushService.SendToToken(user.DeviceToken, n.Title, n.Body, map[string]string{
				"type": string(n.Type),
			})
		}
	}

	return n, nil
}

func (s *NotificationService) Update(id, userID uuid.UUID, req models.NotificationRequest) (*models.Notification, error) {
	n, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, err
	}
	n.Title = req.Title
	n.Body = req.Body
	if req.Type != "" {
		n.Type = req.Type
	}
	n.ScheduledFor = req.ScheduledFor
	if req.Enabled != nil {
		n.Enabled = *req.Enabled
	}
	if err := s.repo.Update(n); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *NotificationService) Delete(id, userID uuid.UUID) error {
	return s.repo.Delete(id, userID)
}
