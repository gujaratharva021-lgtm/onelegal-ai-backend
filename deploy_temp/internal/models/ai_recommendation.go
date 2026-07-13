package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AIRecommendation struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID         uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	CaseID         uuid.UUID      `json:"case_id" gorm:"type:uuid;not null;index"`
	Recommendation string         `json:"recommendation" gorm:"type:text"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}
