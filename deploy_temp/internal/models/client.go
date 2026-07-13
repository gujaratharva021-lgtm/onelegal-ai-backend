package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClientStatus string

const (
	ClientStatusActive ClientStatus = "active"
	ClientStatusClosed ClientStatus = "closed"
)

// ClientAccountStatus tracks the linked login account, independent of the
// CRM engagement Status above. An account is Active only while the client
// has at least one non-completed case; it is automatically flipped to
// Inactive (blocking login) once the lawyer closes all of the client's
// cases, and reactivated the moment a case is reopened or a new one is
// created for the same client. See CaseService.syncClientAccountStatus.
type ClientAccountStatus string

const (
	ClientAccountActive   ClientAccountStatus = "active"
	ClientAccountInactive ClientAccountStatus = "inactive"
)

type Client struct {
	ID          uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID    `json:"user_id" gorm:"type:uuid;not null;index"`
	Name        string       `json:"name" gorm:"not null"`
	Email       string       `json:"email"`
	Phone       string       `json:"phone"`
	Address     string       `json:"address"`
	City        string       `json:"city"`
	State       string       `json:"state"`
	DateOfBirth *time.Time   `json:"date_of_birth"`
	Gender      string       `json:"gender"`
	CaseType    string       `json:"case_type"`
	Notes       string       `json:"notes"`
	Status      ClientStatus `json:"status" gorm:"default:active;index"`
	// LoginID is the email or mobile number the lawyer assigned this client
	// to log in with. AccountUserID links to the actual login User row
	// (Role = RoleClient); AccountStatus gates whether that account can log
	// in right now. Both are set once, at client creation, by
	// ClientService.Create — there is no public client signup.
	LoginID       string              `json:"login_id"`
	AccountUserID *uuid.UUID          `json:"account_user_id" gorm:"type:uuid;index"`
	AccountStatus ClientAccountStatus `json:"account_status" gorm:"default:inactive;index"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
	DeletedAt     gorm.DeletedAt      `json:"-" gorm:"index"`
}

type ClientRequest struct {
	Name        string     `json:"name" binding:"required"`
	Email       string     `json:"email" binding:"omitempty,email"`
	Phone       string     `json:"phone"`
	Address     string     `json:"address"`
	City        string     `json:"city"`
	State       string     `json:"state"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	Gender      string     `json:"gender"`
	CaseType    string     `json:"case_type"`
	Notes       string     `json:"notes"`
	// Status is "active" or "closed"; anything else (including omitted)
	// defaults to "active". Use the dedicated archive endpoint to close a
	// client rather than relying on callers to set this on every update.
	Status string `json:"status"`
	// LoginID/TemporaryPassword are only read on Create (client account
	// creation is a one-time step); Update silently ignores them.
	LoginID           string `json:"login_id"`
	TemporaryPassword string `json:"temporary_password"`
}

// ResetClientPasswordRequest is used by the lawyer's "Reset Client Password"
// action in the Client Profile screen.
type ResetClientPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
