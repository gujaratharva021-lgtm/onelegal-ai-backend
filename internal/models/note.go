package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NoteCategory string

const (
	NoteCategoryPersonal NoteCategory = "personal"
	NoteCategoryCase     NoteCategory = "case"
	NoteCategoryMeeting  NoteCategory = "meeting"
)

type Note struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	CaseID    *uuid.UUID     `json:"case_id" gorm:"type:uuid;index"`
	Title     string         `json:"title" gorm:"not null"`
	Content   string         `json:"content" gorm:"type:text"`
	Category  NoteCategory   `json:"category" gorm:"default:personal;index"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type NoteRequest struct {
	CaseID   *uuid.UUID   `json:"case_id"`
	Title    string       `json:"title" binding:"required"`
	Content  string       `json:"content"`
	Category NoteCategory `json:"category"`
}
