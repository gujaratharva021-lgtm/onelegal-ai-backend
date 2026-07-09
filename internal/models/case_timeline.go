package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CaseTimelineEvent struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CaseID    uuid.UUID      `json:"case_id" gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	EventDate time.Time      `json:"event_date" gorm:"not null"`
	Title     string         `json:"title" gorm:"not null"`
	Notes     string         `json:"notes"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type CaseTimelineEventRequest struct {
	EventDate time.Time `json:"event_date" binding:"required"`
	Title     string    `json:"title" binding:"required"`
	Notes     string    `json:"notes"`
}
