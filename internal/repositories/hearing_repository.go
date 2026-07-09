package repositories

import (
	"time"

	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"

	"github.com/google/uuid"
)

type HearingRepository struct{}

func NewHearingRepository() *HearingRepository {
	return &HearingRepository{}
}

func (r *HearingRepository) Create(h *models.Hearing) error {
	return database.DB.Create(h).Error
}

func (r *HearingRepository) FindByID(id, userID uuid.UUID) (*models.Hearing, error) {
	var h models.Hearing
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&h).Error; err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *HearingRepository) ListByUser(userID uuid.UUID, status string) ([]models.Hearing, error) {
	var hearings []models.Hearing
	q := database.DB.Where("user_id = ?", userID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Order("hearing_date ASC").Find(&hearings).Error; err != nil {
		return nil, err
	}
	return hearings, nil
}

func (r *HearingRepository) ListToday(userID uuid.UUID) ([]models.Hearing, error) {
	var hearings []models.Hearing
	// NOT time.Now().Truncate(24*time.Hour) — Truncate rounds to boundaries
	// aligned on Go's absolute zero time (UTC), not the local calendar day,
	// so on an IST server (UTC+5:30) that window is offset by 5:30 from
	// actual local midnight and silently drops/misses hearings depending on
	// the time of day. time.Date(...) below is the same pattern already used
	// correctly in dashboard_stats_service.go.
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)
	if err := database.DB.Where("user_id = ? AND hearing_date >= ? AND hearing_date < ?", userID, start, end).
		Order("hearing_date ASC").Find(&hearings).Error; err != nil {
		return nil, err
	}
	return hearings, nil
}

func (r *HearingRepository) Update(h *models.Hearing) error {
	return database.DB.Save(h).Error
}

func (r *HearingRepository) Delete(id, userID uuid.UUID) error {
	return database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Hearing{}).Error
}
