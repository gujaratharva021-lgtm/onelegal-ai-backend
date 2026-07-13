package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DraftType string

const (
	DraftTypeLegalNotice         DraftType = "legal_notice"
	DraftTypeAgreement           DraftType = "agreement"
	DraftTypeAffidavit           DraftType = "affidavit"
	DraftTypePetition            DraftType = "petition"
	DraftTypePowerOfAttorney     DraftType = "power_of_attorney"
	DraftTypeBailApplication     DraftType = "bail_application"
	DraftTypeDivorcePetition     DraftType = "divorce_petition"
	DraftTypeEmploymentAgreement DraftType = "employment_agreement"
	DraftTypeRentAgreement       DraftType = "rent_agreement"
	DraftTypeCustom              DraftType = "custom"
)

type DraftStatus string

const (
	DraftStatusDraft DraftStatus = "draft"
	DraftStatusFinal DraftStatus = "final"
)

type Draft struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	CaseID    *uuid.UUID     `json:"case_id" gorm:"type:uuid;index"`
	ClientID  *uuid.UUID     `json:"client_id" gorm:"type:uuid;index"`
	Title     string         `json:"title" gorm:"not null"`
	DraftType DraftType      `json:"draft_type" gorm:"index"`
	Content   string         `json:"content" gorm:"type:text"`
	Status    DraftStatus    `json:"status" gorm:"default:draft"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type DraftRequest struct {
	CaseID    *uuid.UUID  `json:"case_id"`
	ClientID  *uuid.UUID  `json:"client_id"`
	Title     string      `json:"title" binding:"required"`
	DraftType DraftType   `json:"draft_type" binding:"required"`
	Content   string      `json:"content"`
	Status    DraftStatus `json:"status"`
}

type AIDraftGenerateRequest struct {
	DraftType    DraftType `json:"draft_type" binding:"required"`
	Title        string    `json:"title" binding:"required"`
	Instructions string    `json:"instructions" binding:"required"`
}
