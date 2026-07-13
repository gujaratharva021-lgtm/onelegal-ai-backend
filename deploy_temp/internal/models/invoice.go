package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InvoiceStatus string

// Lifecycle: Draft (lawyer is still preparing it) -> Sent ("Pending" in the
// UI — issued, awaiting payment) -> Submitted (client completed a UPI Intent
// payment; NOT auto-verified) -> Paid (lawyer manually confirms funds landed
// in their bank account). Failed/cancelled UPI attempts leave the invoice in
// Sent/"Pending" — they never move it to Failed automatically.
const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusSent      InvoiceStatus = "sent"
	InvoiceStatusSubmitted InvoiceStatus = "submitted"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusFailed    InvoiceStatus = "failed"
	// InvoiceStatusCancelled is set only by the lawyer's manual Payments
	// action — a payment that will never happen (e.g. waived, disputed).
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
)

// Invoice is a lawyer-issued bill for a client.
type Invoice struct {
	ID            uuid.UUID     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	LawyerID      uuid.UUID     `json:"lawyer_id" gorm:"type:uuid;not null;index"`
	ClientID      uuid.UUID     `json:"client_id" gorm:"type:uuid;not null;index"`
	CaseID        *uuid.UUID    `json:"case_id" gorm:"type:uuid;index"`
	InvoiceNumber string        `json:"invoice_number" gorm:"uniqueIndex;not null"`
	Description   string        `json:"description"`
	AmountPaise   int64         `json:"amount_paise" gorm:"not null"`
	Currency      string        `json:"currency" gorm:"default:INR"`
	Status        InvoiceStatus `json:"status" gorm:"default:draft;index"`
	DueDate       *time.Time    `json:"due_date"`
	PaidAt        *time.Time    `json:"paid_at"`
	// LawyerUpiID snapshots the lawyer's UPI ID at the moment the invoice
	// was generated, so a later change to the lawyer's payment settings
	// never alters where an already-issued invoice's payment is directed.
	LawyerUpiID string `json:"lawyer_upi_id"`
	// LawyerName is a read-only convenience field for the client's Pay Now
	// screen (the UPI intent's payee-name parameter) — populated by
	// InvoiceService at read time, never persisted as its own column.
	LawyerName string `json:"lawyer_name" gorm:"-"`
	// ClientName/ClientPhone are read-only conveniences for the lawyer's
	// "Share via WhatsApp" action on this invoice — populated by
	// InvoiceService at read time, never persisted as their own columns.
	ClientName  string `json:"client_name" gorm:"-"`
	ClientPhone string `json:"client_phone" gorm:"-"`
	// Populated once the client completes (or the UPI app reports) a
	// payment attempt via the Android UPI Intent. Never trusted as proof of
	// payment on its own — only a manual lawyer verification sets Paid.
	TransactionRef  string     `json:"transaction_ref"`
	TransactionID   string     `json:"transaction_id"`
	PaymentDate     *time.Time `json:"payment_date"`
	PaidAmountPaise int64      `json:"paid_amount_paise"`
	// PaymentApp is the UPI app the client picked (e.g. "Google Pay"), when
	// Android was able to report it — purely informational display, never
	// used to verify a payment.
	PaymentApp string `json:"payment_app"`
	// UTRNumber is the bank UTR/transaction reference the lawyer manually
	// enters when confirming a payment landed in their account — the actual
	// proof of payment, distinct from TransactionRef (whatever the UPI app
	// itself reported, which is never trusted alone). Denormalized here from
	// the Payment row of the same name so every screen that already reads
	// Invoice fields shows it with no extra fetch.
	UTRNumber string         `json:"utr_number"`
	Notes     string         `json:"notes"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// SubmitPaymentRequest is sent by the client's app right after the Android
// UPI Intent returns control to it.
type SubmitPaymentRequest struct {
	// Status is one of "SUCCESS", "FAILED", "CANCELLED" as read from the UPI
	// app's intent response.
	Status          string `json:"status" binding:"required"`
	TransactionRef  string `json:"transaction_ref"`
	TransactionID   string `json:"transaction_id"`
	PaidAmountPaise int64  `json:"paid_amount_paise"`
	AppUsed         string `json:"app_used"`
}

type CreateInvoiceRequest struct {
	ClientID    uuid.UUID  `json:"client_id" binding:"required"`
	CaseID      *uuid.UUID `json:"case_id"`
	Description string     `json:"description" binding:"required"`
	// AmountPaise is the amount in the smallest currency unit (paise for
	// INR).
	AmountPaise int64      `json:"amount_paise" binding:"required,min=100"`
	DueDate     *time.Time `json:"due_date"`
	Notes       string     `json:"notes"`
}

// UpdateInvoiceRequest edits an existing invoice's content — only allowed
// while it hasn't been paid/submitted for payment yet (see
// InvoiceService.Update).
type UpdateInvoiceRequest struct {
	Description string     `json:"description" binding:"required"`
	AmountPaise int64      `json:"amount_paise" binding:"required,min=100"`
	DueDate     *time.Time `json:"due_date"`
	Notes       string     `json:"notes"`
}

// SetPaymentStatusRequest is the lawyer's manual Payments-module action —
// Status is one of "pending", "paid", "cancelled". UTRNumber is required
// when Status is "paid" — the lawyer's manually-entered bank transaction
// reference, the actual proof funds arrived.
type SetPaymentStatusRequest struct {
	Status    string `json:"status" binding:"required"`
	UTRNumber string `json:"utr_number"`
}
