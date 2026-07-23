package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CaseStatus string

const (
	CaseStatusUpcoming  CaseStatus = "upcoming"
	CaseStatusOngoing   CaseStatus = "ongoing"
	CaseStatusCompleted CaseStatus = "completed"
)

type CasePriority string

const (
	CasePriorityLow    CasePriority = "low"
	CasePriorityMedium CasePriority = "medium"
	CasePriorityHigh   CasePriority = "high"
)

type Case struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	ClientID        *uuid.UUID     `json:"client_id" gorm:"type:uuid;index"`
	Title           string         `json:"title" gorm:"not null"`
	CaseNumber      string         `json:"case_number"`
	CourtName       string         `json:"court_name"`
	CourtNumber     string         `json:"court_number"`
	Judge           string         `json:"judge"`
	Opponent        string         `json:"opponent"`
	CaseType        string         `json:"case_type"`
CNRNumber       string         `json:"cnr_number"`
LodgingNumber   string         `json:"lodging_number"`
FilingDate      *time.Time     `json:"filing_date"`
RegNumber       string         `json:"registration_number"`
RegDate         *time.Time     `json:"registration_date"`
Petitioner      string         `json:"petitioner"`
Respondent      string         `json:"respondent"`
PetnAdvocate    string         `json:"petitioner_advocate"`
RespAdvocate    string         `json:"respondent_advocate"`
District        string         `json:"district"`
Bench           string         `json:"bench"`
Category        string         `json:"category"`
Stage           string         `json:"stage"`
Coram           string         `json:"coram"`
LastHearingDate *time.Time     `json:"last_hearing_date"`
LastCoram       string         `json:"last_coram"`
Act             string         `json:"act"`
	Priority        CasePriority   `json:"priority" gorm:"default:medium;index"`
	Status          CaseStatus     `json:"status" gorm:"default:upcoming;index"`
	Description     string         `json:"description"`
	NextHearingDate *time.Time     `json:"next_hearing_date"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

type CaseRequest struct {
	ClientID        *uuid.UUID   `json:"client_id"`
	Title           string       `json:"title" binding:"required"`
	CaseNumber      string       `json:"case_number"`
	CourtName       string       `json:"court_name"`
	CourtNumber     string       `json:"court_number"`
	Judge           string       `json:"judge"`
	Opponent        string       `json:"opponent"`
	CaseType        string       `json:"case_type"`
CNRNumber       string       `json:"cnr_number"`
LodgingNumber   string       `json:"lodging_number"`
FilingDate      *time.Time   `json:"filing_date"`
RegNumber       string       `json:"registration_number"`
RegDate         *time.Time   `json:"registration_date"`
Petitioner      string       `json:"petitioner"`
Respondent      string       `json:"respondent"`
PetnAdvocate    string       `json:"petitioner_advocate"`
RespAdvocate    string       `json:"respondent_advocate"`
District        string       `json:"district"`
Bench           string       `json:"bench"`
Category        string       `json:"category"`
Stage           string       `json:"stage"`
Coram           string       `json:"coram"`
LastHearingDate *time.Time   `json:"last_hearing_date"`
LastCoram       string       `json:"last_coram"`
Act             string       `json:"act"`
	Priority        CasePriority `json:"priority"`
	Status          CaseStatus   `json:"status"`
	Description     string       `json:"description"`
	NextHearingDate *time.Time   `json:"next_hearing_date"`
}
