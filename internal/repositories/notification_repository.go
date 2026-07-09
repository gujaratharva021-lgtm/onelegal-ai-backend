package repositories

import (
	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"

	"github.com/google/uuid"
)

type NotificationRepository struct{}

func NewNotificationRepository() *NotificationRepository {
	return &NotificationRepository{}
}

func (r *NotificationRepository) Create(n *models.Notification) error {
	return database.DB.Create(n).Error
}

func (r *NotificationRepository) ListByUser(userID uuid.UUID, unreadOnly bool) ([]models.Notification, error) {
	var notifications []models.Notification
	q := database.DB.Where("user_id = ?", userID)
	if unreadOnly {
		q = q.Where("is_read = false")
	}
	if err := q.Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *NotificationRepository) MarkRead(id, userID uuid.UUID) error {
	return database.DB.Model(&models.Notification{}).Where("id = ? AND user_id = ?", id, userID).Update("is_read", true).Error
}

func (r *NotificationRepository) MarkAllRead(userID uuid.UUID) error {
	return database.DB.Model(&models.Notification{}).Where("user_id = ? AND is_read = false", userID).Update("is_read", true).Error
}

func (r *NotificationRepository) CountUnread(userID uuid.UUID) (int64, error) {
	var count int64
	err := database.DB.Model(&models.Notification{}).Where("user_id = ? AND is_read = false", userID).Count(&count).Error
	return count, err
}

func (r *NotificationRepository) FindByID(id, userID uuid.UUID) (*models.Notification, error) {
	var n models.Notification
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&n).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NotificationRepository) Update(n *models.Notification) error {
	return database.DB.Save(n).Error
}

func (r *NotificationRepository) Delete(id, userID uuid.UUID) error {
	return database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Notification{}).Error
}
