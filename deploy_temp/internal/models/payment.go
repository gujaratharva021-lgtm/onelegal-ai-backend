package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentStatus string

const (
	// PaymentStatusPendingVerification is set the moment a client returns
	// from a successful Android UPI Intent attempt — never proof of payment
	// on its own.
	PaymentStatusPendingVerification PaymentStatus = "pending_verification"
	// PaymentStatusPaid is set only when the lawyer manually confirms the
	// money landed in their bank account and enters the UTR number.
	PaymentStatusPaid PaymentStatus = "paid"
	// PaymentStatusCancelled is the lawyer's manual action — a payment that
	// will never happen (waived, disputed) — see InvoiceService.SetPaymentStatus.
	PaymentStatusCancelled PaymentStatus = "cancelled"
	// PaymentStatusFailed is set automatically when the client's own UPI
	// Intent attempt comes back FAILED or CANCELLED from the UPI app itself
	// (e.g. they backed out, insufficient balance) — distinct from the
	// lawyer's manual PaymentStatusCancelled, so Payment History shows what
	// actually happened at the UPI-app layer vs. a lawyer decision.
	PaymentStatusFailed PaymentStatus = "failed"
)

// Payment is the full audit record of one invoice's payment lifecycle —
// separate from Invoice itself so the UTR number, who verified it, and
// exactly when funds were confirmed are tracked independently of the
// invoice's own content. One invoice has at most one Payment row, created
// the moment the client attempts payment via the Android UPI Intent.
type Payment struct {
	ID          uuid.UUID     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	InvoiceID   uuid.UUID     `json:"invoice_id" gorm:"type:uuid;not null;uniqueIndex"`
	LawyerID    uuid.UUID     `json:"lawyer_id" gorm:"type:uuid;not null;index"`
	ClientID    uuid.UUID     `json:"client_id" gorm:"type:uuid;not null;index"`
	AmountPaise int64         `json:"amount_paise" gorm:"not null"`
	Status      PaymentStatus `json:"status" gorm:"default:pending_verification;index"`
	// UTRNumber is the lawyer's manually-entered bank transaction reference
	// — required before Status can become "paid".
	UTRNumber string `json:"utr_number"`
	// VerifiedBy is the lawyer user id who marked this Paid — always the
	// invoice's own LawyerID today (only the owning lawyer can verify), kept
	// as its own column for a clear audit trail.
	VerifiedBy *uuid.UUID `json:"verified_by" gorm:"type:uuid"`
	// PaymentDate is when the lawyer confirmed/marked Paid; AttemptedAt is
	// when the client's UPI Intent attempt itself came back to the app.
	PaymentDate *time.Time `json:"payment_date"`
	AttemptedAt time.Time  `json:"attempted_at"`
	// TransactionRef/TransactionID/AppUsed are whatever the UPI app itself
	// reported on return — informational only, never trusted as proof.
	TransactionRef string         `json:"transaction_ref"`
	TransactionID  string         `json:"transaction_id"`
	AppUsed        string         `json:"app_used"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}
