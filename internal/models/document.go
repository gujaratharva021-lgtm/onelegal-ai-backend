package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Document struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	CaseID       *uuid.UUID     `json:"case_id" gorm:"type:uuid;index"`
	ClientID     *uuid.UUID     `json:"client_id" gorm:"type:uuid;index"`
	Title        string         `json:"title" gorm:"not null"`
	DocumentType string         `json:"document_type" gorm:"index"`
	Content      string         `json:"content" gorm:"type:text"`
	FileName     string         `json:"file_name"`
	FilePath     string         `json:"-"`
	FileURL      string         `json:"file_url"`
	FileType     string         `json:"file_type"`
	FileSize     int64          `json:"file_size"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type DocumentRequest struct {
	CaseID       *uuid.UUID `json:"case_id"`
	ClientID     *uuid.UUID `json:"client_id"`
	Title        string     `json:"title" binding:"required"`
	DocumentType string     `json:"document_type"`
	Content      string     `json:"content"`
	FileURL      string     `json:"file_url"`
	FileType     string     `json:"file_type"`
	FileSize     int64      `json:"file_size"`
}
