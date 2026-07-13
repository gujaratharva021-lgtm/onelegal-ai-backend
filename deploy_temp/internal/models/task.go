package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusCompleted TaskStatus = "completed"
)

type Task struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	CaseID      *uuid.UUID     `json:"case_id" gorm:"type:uuid;index"`
	Title       string         `json:"title" gorm:"not null"`
	Description string         `json:"description"`
	DueDate     *time.Time     `json:"due_date"`
	Priority    TaskPriority   `json:"priority" gorm:"default:medium"`
	Status      TaskStatus     `json:"status" gorm:"default:pending;index"`
	ReminderAt  *time.Time     `json:"reminder_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	// CaseTitle is a read-only convenience for the Home Dashboard's task
	// list — populated by DashboardService at read time from the linked
	// Case (when set), never persisted as its own column.
	CaseTitle string `json:"case_title" gorm:"-"`
}

type TaskRequest struct {
	CaseID      *uuid.UUID   `json:"case_id"`
	Title       string       `json:"title" binding:"required"`
	Description string       `json:"description"`
	DueDate     *time.Time   `json:"due_date"`
	Priority    TaskPriority `json:"priority"`
	Status      TaskStatus   `json:"status"`
	ReminderAt  *time.Time   `json:"reminder_at"`
}
