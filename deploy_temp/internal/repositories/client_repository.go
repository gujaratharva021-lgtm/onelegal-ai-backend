package repositories

import (
	"errors"

	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClientRepository struct{}

func NewClientRepository() *ClientRepository {
	return &ClientRepository{}
}

func (r *ClientRepository) Create(c *models.Client) error {
	return database.DB.Create(c).Error
}

func (r *ClientRepository) FindByID(id, userID uuid.UUID) (*models.Client, error) {
	var c models.Client
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ClientRepository) ListByUser(userID uuid.UUID, search string) ([]models.Client, error) {
	var clients []models.Client
	q := database.DB.Where("user_id = ?", userID)
	if search != "" {
		q = q.Where("name ILIKE ? OR email ILIKE ? OR phone ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if err := q.Order("name ASC").Find(&clients).Error; err != nil {
		return nil, err
	}
	return clients, nil
}

func (r *ClientRepository) FindDuplicate(userID uuid.UUID, email, phone string, excludeID *uuid.UUID) (*models.Client, error) {
	if email == "" && phone == "" {
		return nil, nil
	}
	q := database.DB.Where("user_id = ?", userID)
	if email != "" && phone != "" {
		q = q.Where("(email = ? AND email <> '') OR (phone = ? AND phone <> '')", email, phone)
	} else if email != "" {
		q = q.Where("email = ? AND email <> ''", email)
	} else {
		q = q.Where("phone = ? AND phone <> ''", phone)
	}
	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}
	var c models.Client
	err := q.First(&c).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// FindByAccountUserID resolves the CRM Client record linked to a client-role
// login account — used on every client-portal request to scope data to
// "their own" records, and on login to check AccountStatus.
func (r *ClientRepository) FindByAccountUserID(userID uuid.UUID) (*models.Client, error) {
	var c models.Client
	if err := database.DB.Where("account_user_id = ?", userID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ClientRepository) Update(c *models.Client) error {
	return database.DB.Save(c).Error
}

func (r *ClientRepository) Delete(id, userID uuid.UUID) error {
	return database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Client{}).Error
}
