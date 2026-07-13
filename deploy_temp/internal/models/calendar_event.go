package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CalendarEventType string

const (
	CalendarEventHearing  CalendarEventType = "hearing"
	CalendarEventMeeting  CalendarEventType = "meeting"
	CalendarEventReminder CalendarEventType = "reminder"
	CalendarEventTask     CalendarEventType = "task"
	CalendarEventPersonal CalendarEventType = "personal"
)

type CalendarEvent struct {
	ID               uuid.UUID         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID           uuid.UUID         `json:"user_id" gorm:"type:uuid;not null;index"`
	Title            string            `json:"title" gorm:"not null"`
	Description      string            `json:"description"`
	EventType        CalendarEventType `json:"event_type" gorm:"default:personal;index"`
	StartTime        time.Time         `json:"start_time" gorm:"index"`
	EndTime          *time.Time        `json:"end_time"`
	Location         string            `json:"location"`
	RelatedCaseID    *uuid.UUID        `json:"related_case_id" gorm:"type:uuid"`
	RelatedMeetingID *uuid.UUID        `json:"related_meeting_id" gorm:"type:uuid"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	DeletedAt        gorm.DeletedAt    `json:"-" gorm:"index"`
}

type CalendarEventRequest struct {
	Title            string            `json:"title" binding:"required"`
	Description      string            `json:"description"`
	EventType        CalendarEventType `json:"event_type"`
	StartTime        time.Time         `json:"start_time" binding:"required"`
	EndTime          *time.Time        `json:"end_time"`
	Location         string            `json:"location"`
	RelatedCaseID    *uuid.UUID        `json:"related_case_id"`
	RelatedMeetingID *uuid.UUID        `json:"related_meeting_id"`
}
