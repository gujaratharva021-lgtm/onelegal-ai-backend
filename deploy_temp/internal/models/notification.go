package models

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeTask    NotificationType = "task"
	NotificationTypeMeeting NotificationType = "meeting"
	NotificationTypeCourt   NotificationType = "court"
	NotificationTypeGeneral NotificationType = "general"
)

type Notification struct {
	ID           uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       uuid.UUID        `json:"user_id" gorm:"type:uuid;not null;index"`
	Title        string           `json:"title" gorm:"not null"`
	Body         string           `json:"body"`
	Type         NotificationType `json:"type" gorm:"default:general;index"`
	IsRead       bool             `json:"is_read" gorm:"default:false;index"`
	RelatedID    *uuid.UUID       `json:"related_id" gorm:"type:uuid"`
	ScheduledFor *time.Time       `json:"scheduled_for"`
	Enabled      bool             `json:"enabled" gorm:"default:true"`
	CreatedAt    time.Time        `json:"created_at"`
}

// NotificationRequest is used to create or update a reminder-style
// notification (e.g. "remind me 30 min before this hearing").
type NotificationRequest struct {
	Title        string           `json:"title" binding:"required"`
	Body         string           `json:"body"`
	Type         NotificationType `json:"type"`
	RelatedID    *uuid.UUID       `json:"related_id"`
	ScheduledFor *time.Time       `json:"scheduled_for"`
	Enabled      *bool            `json:"enabled"`
}
