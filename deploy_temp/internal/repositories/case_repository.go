package repositories

import (
	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CaseRepository struct{}

func NewCaseRepository() *CaseRepository {
	return &CaseRepository{}
}

func (r *CaseRepository) Create(c *models.Case) error {
	return database.DB.Create(c).Error
}

func (r *CaseRepository) FindByID(id, userID uuid.UUID) (*models.Case, error) {
	var c models.Case
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

type CaseListFilter struct {
	Search   string
	Status   string
	ClientID string
	Priority string
	Court    string
	DateFrom string
	DateTo   string
	Page     int
	PageSize int
}

func (r *CaseRepository) buildListQuery(userID uuid.UUID, f CaseListFilter) *gorm.DB {
	q := database.DB.Model(&models.Case{}).Where("user_id = ?", userID)
	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where("title ILIKE ? OR case_number ILIKE ?", like, like)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.ClientID != "" {
		q = q.Where("client_id = ?", f.ClientID)
	}
	if f.Priority != "" {
		q = q.Where("priority = ?", f.Priority)
	}
	if f.Court != "" {
		q = q.Where("court_name ILIKE ?", "%"+f.Court+"%")
	}
	if f.DateFrom != "" {
		q = q.Where("next_hearing_date >= ?", f.DateFrom)
	}
	if f.DateTo != "" {
		q = q.Where("next_hearing_date <= ?", f.DateTo)
	}
	return q
}

func (r *CaseRepository) ListByUser(userID uuid.UUID, f CaseListFilter) ([]models.Case, int64, error) {
	var cases []models.Case
	var total int64
	if err := r.buildListQuery(userID, f).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	q := r.buildListQuery(userID, f).Order("created_at DESC")
	if f.PageSize > 0 {
		page := f.Page
		if page < 1 {
			page = 1
		}
		q = q.Limit(f.PageSize).Offset((page - 1) * f.PageSize)
	}
	if err := q.Find(&cases).Error; err != nil {
		return nil, 0, err
	}
	return cases, total, nil
}

func (r *CaseRepository) ListRecent(userID uuid.UUID, limit int) ([]models.Case, error) {
	var cases []models.Case
	if err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&cases).Error; err != nil {
		return nil, err
	}
	return cases, nil
}

// CountOpenByClient counts this client's non-completed cases — used to
// decide whether their linked login account should be active or inactive.
func (r *CaseRepository) CountOpenByClient(clientID, userID uuid.UUID) (int64, error) {
	var count int64
	err := database.DB.Model(&models.Case{}).
		Where("user_id = ? AND client_id = ? AND status <> ?", userID, clientID, models.CaseStatusCompleted).
		Count(&count).Error
	return count, err
}

// ListByClientID is used by the client portal (the client viewing their own
// cases) — intentionally not scoped by lawyer id since the caller here is
// the client themself, and clientID is always resolved server-side from
// their authenticated account, never taken from client input.
func (r *CaseRepository) ListByClientID(clientID uuid.UUID) ([]models.Case, error) {
	var cases []models.Case
	if err := database.DB.Where("client_id = ?", clientID).Order("created_at DESC").Find(&cases).Error; err != nil {
		return nil, err
	}
	return cases, nil
}

func (r *CaseRepository) Update(c *models.Case) error {
	return database.DB.Save(c).Error
}

func (r *CaseRepository) Delete(id, userID uuid.UUID) error {
	result := database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Case{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
