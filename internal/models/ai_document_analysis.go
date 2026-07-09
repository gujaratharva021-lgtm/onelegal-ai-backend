package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AIAnalysisType string

const (
	AIAnalysisTypeSummary  AIAnalysisType = "summary"
	AIAnalysisTypeContract AIAnalysisType = "contract"
)

type AIDocumentAnalysis struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	Type      AIAnalysisType `json:"type" gorm:"not null;index"`
	FileName  string         `json:"file_name"`
	FilePath  string         `json:"-"`
	Result    string         `json:"result" gorm:"type:text"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
