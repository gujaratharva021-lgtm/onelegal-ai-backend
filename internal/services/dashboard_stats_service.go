package services

import (
	"time"

	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"

	"github.com/google/uuid"
)

type CaseStatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type CaseMonthCount struct {
	Month string `json:"month"`
	Count int64  `json:"count"`
}

type DashboardStatistics struct {
	TotalCases        int64             `json:"total_cases"`
	ActiveCases       int64             `json:"active_cases"`
	ClosedCases       int64             `json:"closed_cases"`
	TodaysHearings    int64             `json:"todays_hearings"`
	UpcomingHearings  int64             `json:"upcoming_hearings"`
	PendingTasks      int64             `json:"pending_tasks"`
	TotalClients      int64             `json:"total_clients"`
	DocumentsUploaded int64             `json:"documents_uploaded"`
	CasesByStatus     []CaseStatusCount `json:"cases_by_status"`
	CasesByMonth      []CaseMonthCount  `json:"cases_by_month"`
}

type DashboardStatsService struct{}

func NewDashboardStatsService() *DashboardStatsService {
	return &DashboardStatsService{}
}

func (s *DashboardStatsService) Get(userID uuid.UUID) (*DashboardStatistics, error) {
	db := database.DB
	stats := &DashboardStatistics{}

	if err := db.Model(&models.Case{}).Where("user_id = ?", userID).Count(&stats.TotalCases).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&models.Case{}).
		Where("user_id = ? AND status IN ?", userID, []string{string(models.CaseStatusOngoing), string(models.CaseStatusUpcoming)}).
		Count(&stats.ActiveCases).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&models.Case{}).
		Where("user_id = ? AND status = ?", userID, models.CaseStatusCompleted).
		Count(&stats.ClosedCases).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour)

	if err := db.Model(&models.Hearing{}).
		Where("user_id = ? AND hearing_date >= ? AND hearing_date < ?", userID, todayStart, todayEnd).
		Count(&stats.TodaysHearings).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&models.Hearing{}).
		Where("user_id = ? AND hearing_date >= ?", userID, todayEnd).
		Count(&stats.UpcomingHearings).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&models.Task{}).
		Where("user_id = ? AND status = ?", userID, "pending").
		Count(&stats.PendingTasks).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&models.Client{}).Where("user_id = ?", userID).Count(&stats.TotalClients).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&models.Document{}).Where("user_id = ?", userID).Count(&stats.DocumentsUploaded).Error; err != nil {
		return nil, err
	}

	var statusCounts []CaseStatusCount
	if err := db.Model(&models.Case{}).
		Select("status, count(*) as count").
		Where("user_id = ?", userID).
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		return nil, err
	}
	stats.CasesByStatus = statusCounts

	var monthCounts []CaseMonthCount
	if err := db.Model(&models.Case{}).
		Select("to_char(created_at, 'YYYY-MM') as month, count(*) as count").
		Where("user_id = ? AND created_at >= ?", userID, now.AddDate(0, -11, 0)).
		Group("month").
		Order("month ASC").
		Scan(&monthCounts).Error; err != nil {
		return nil, err
	}
	stats.CasesByMonth = monthCounts

	return stats, nil
}
