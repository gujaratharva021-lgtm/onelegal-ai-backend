package models

import (
	"time"

	"github.com/google/uuid"
)

type ResearchCategory string

const (
	ResearchCategoryCase      ResearchCategory = "case"
	ResearchCategoryLaw       ResearchCategory = "law"
	ResearchCategoryAct       ResearchCategory = "act"
	ResearchCategorySection   ResearchCategory = "section"
	ResearchCategoryJudgement ResearchCategory = "judgement"
)

type ResearchHistory struct {
	ID            uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        uuid.UUID        `json:"user_id" gorm:"type:uuid;not null;index"`
	Query         string           `json:"query" gorm:"not null"`
	Category      ResearchCategory `json:"category" gorm:"index"`
	ResultSummary string           `json:"result_summary" gorm:"type:text"`
	Bookmarked    bool             `json:"bookmarked" gorm:"default:false;index"`
	CreatedAt     time.Time        `json:"created_at"`
}

type ResearchRequest struct {
	Query    string           `json:"query" binding:"required"`
	Category ResearchCategory `json:"category" binding:"required"`
}

type ResearchBookmarkRequest struct {
	Bookmarked bool `json:"bookmarked"`
}
