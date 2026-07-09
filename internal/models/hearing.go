package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HearingStatus string

const (
	HearingStatusUpcoming  HearingStatus = "upcoming"
	HearingStatusOngoing   HearingStatus = "ongoing"
	HearingStatusCompleted HearingStatus = "completed"
	HearingStatusAdjourned HearingStatus = "adjourned"
	HearingStatusCancelled HearingStatus = "cancelled"
)

type Hearing struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	CaseID      uuid.UUID      `json:"case_id" gorm:"type:uuid;not null;index"`
	CourtName   string         `json:"court_name"`
	CourtNumber string         `json:"court_number"`
	Judge       string         `json:"judge"`
	HearingDate time.Time      `json:"hearing_date" gorm:"index"`
	HearingType string         `json:"hearing_type"`
	Status      HearingStatus  `json:"status" gorm:"default:upcoming;index"`
	Notes       string         `json:"notes"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	// CaseTitle/CaseStatus are read-only conveniences for the Home Dashboard's
	// hearing cards — populated by DashboardService at read time from the
	// linked Case, never persisted as their own columns.
	CaseTitle  string `json:"case_title" gorm:"-"`
	CaseStatus string `json:"case_status" gorm:"-"`
	ClientName string `json:"client_name" gorm:"-"`
}

type HearingRequest struct {
	CaseID      uuid.UUID     `json:"case_id" binding:"required"`
	CourtName   string        `json:"court_name"`
	CourtNumber string        `json:"court_number"`
	Judge       string        `json:"judge"`
	HearingDate time.Time     `json:"hearing_date" binding:"required"`
	HearingType string        `json:"hearing_type"`
	Status      HearingStatus `json:"status"`
	Notes       string        `json:"notes"`
}
