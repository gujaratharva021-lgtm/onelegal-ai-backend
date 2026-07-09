package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"

	"github.com/google/uuid"
)

var ErrNotSubmittedForPayment = errors.New("only an invoice with a submitted payment can be marked paid")
var ErrInvoiceLocked = errors.New("this invoice can no longer be edited because a payment has been submitted or confirmed")

// InvoiceService is the lawyer-facing half of billing: create, list, view,
// send, and delete a lawyer's own invoices. It also backs the client-facing
// UPI Intent payment flow (submit payment result) and the lawyer's manual
// payment verification (mark paid) — reusing this same service/table, no
// second invoice/payment system.
type InvoiceService struct {
	repo                *repositories.InvoiceRepository
	clientRepo          *repositories.ClientRepository
	userRepo            *repositories.UserRepository
	paymentRepo         *repositories.PaymentRepository
	notificationService *NotificationService
	emailService        *EmailService
}

func NewInvoiceService() *InvoiceService {
	return &InvoiceService{
		repo:                repositories.NewInvoiceRepository(),
		clientRepo:          repositories.NewClientRepository(),
		userRepo:            repositories.NewUserRepository(),
		paymentRepo:         repositories.NewPaymentRepository(),
		notificationService: NewNotificationService(),
		emailService:        NewEmailService(GetConfig()),
	}
}

func (s *InvoiceService) enrichInvoice(inv *models.Invoice) {
	if lawyer, err := s.userRepo.FindByID(inv.LawyerID); err == nil {
		inv.LawyerName = lawyer.Name
	}
	if client, err := s.clientRepo.FindByID(inv.ClientID, inv.LawyerID); err == nil {
		inv.ClientName = client.Name
		inv.ClientPhone = client.Phone
	}
}

func (s *InvoiceService) Create(lawyerID uuid.UUID, req models.CreateInvoiceRequest) (*models.Invoice, error) {
	client, err := s.clientRepo.FindByID(req.ClientID, lawyerID)
	if err != nil {
		return nil, errors.New("client not found")
	}
	lawyer, err := s.userRepo.FindByID(lawyerID)
	if err != nil {
		return nil, errors.New("lawyer account not found")
	}

	inv := &models.Invoice{
		LawyerID:      lawyerID,
		ClientID:      req.ClientID,
		CaseID:        req.CaseID,
		InvoiceNumber: fmt.Sprintf("INV-%d", time.Now().UnixNano()/int64(time.Millisecond)),
		Description:   req.Description,
		AmountPaise:   req.AmountPaise,
		Currency:      "INR",
		Status:        models.InvoiceStatusDraft,
		DueDate:       req.DueDate,
		Notes:         req.Notes,
		// Snapshotted now so a later change to the lawyer's UPI ID never
		// alters where this invoice's payment goes.
		LawyerUpiID: lawyer.UpiID,
	}
	if err := s.repo.Create(inv); err != nil {
		return nil, err
	}
	inv.LawyerName = lawyer.Name
	inv.ClientName = client.Name
	inv.ClientPhone = client.Phone
	return inv, nil
}

func (s *InvoiceService) List(lawyerID uuid.UUID) ([]models.Invoice, error) {
	invoices, err := s.repo.ListByLawyer(lawyerID)
	if err != nil {
		return nil, err
	}
	for i := range invoices {
		s.enrichInvoice(&invoices[i])
	}
	return invoices, nil
}

func (s *InvoiceService) Get(id, lawyerID uuid.UUID) (*models.Invoice, error) {
	inv, err := s.repo.FindByID(id, lawyerID)
	if err != nil {
		return nil, err
	}
	s.enrichInvoice(inv)
	return inv, nil
}

// PDFContext gathers everything the invoice PDF needs beyond the invoice row
// itself — the lawyer's own business details (name, law firm, GST, UPI) and
// the client's details — reusing the existing User/Client tables, no new
// model.
func (s *InvoiceService) PDFContext(id, lawyerID uuid.UUID) (*models.Invoice, *models.User, *models.Client, error) {
	inv, err := s.repo.FindByID(id, lawyerID)
	if err != nil {
		return nil, nil, nil, err
	}
	s.enrichInvoice(inv)
	lawyer, err := s.userRepo.FindByID(lawyerID)
	if err != nil {
		return nil, nil, nil, err
	}
	client, err := s.clientRepo.FindByID(inv.ClientID, lawyerID)
	if err != nil {
		return nil, nil, nil, err
	}
	return inv, lawyer, client, nil
}

func (s *InvoiceService) ListForClient(lawyerID, clientID uuid.UUID) ([]models.Invoice, error) {
	invoices, err := s.repo.ListByLawyerAndClient(lawyerID, clientID)
	if err != nil {
		return nil, err
	}
	for i := range invoices {
		s.enrichInvoice(&invoices[i])
	}
	return invoices, nil
}

// GetForClient / ListForClientPortal are the client-role equivalents of
// Get/List above, scoped by the caller's own linked Client id instead of a
// lawyer id.
func (s *InvoiceService) GetForClient(id, clientID uuid.UUID) (*models.Invoice, error) {
	inv, err := s.repo.FindByIDForClient(id, clientID)
	if err != nil {
		return nil, err
	}
	s.enrichInvoice(inv)
	return inv, nil
}

func (s *InvoiceService) ListForClientPortal(clientID uuid.UUID) ([]models.Invoice, error) {
	invoices, err := s.repo.ListByClientID(clientID)
	if err != nil {
		return nil, err
	}
	for i := range invoices {
		s.enrichInvoice(&invoices[i])
	}
	return invoices, nil
}

// Update edits an existing invoice's content (description/amount/due
// date/payment method/notes). Only allowed before a payment has been
// submitted or confirmed, so an in-flight or completed payment's record
// can never be silently altered after the fact.
func (s *InvoiceService) Update(id, lawyerID uuid.UUID, req models.UpdateInvoiceRequest) (*models.Invoice, error) {
	inv, err := s.repo.FindByID(id, lawyerID)
	if err != nil {
		return nil, errors.New("invoice not found")
	}
	if inv.Status == models.InvoiceStatusSubmitted || inv.Status == models.InvoiceStatusPaid {
		return nil, ErrInvoiceLocked
	}
	inv.Description = req.Description
	inv.AmountPaise = req.AmountPaise
	inv.DueDate = req.DueDate
	inv.Notes = req.Notes
	if err := s.repo.Update(inv); err != nil {
		return nil, err
	}
	s.enrichInvoice(inv)
	return inv, nil
}

// Send transitions Draft -> Sent, making the invoice issued — and, best
// effort, notifies the client by email and push/in-app notification. Never
// fails the Send itself if either notification channel is unavailable.
func (s *InvoiceService) Send(id, lawyerID uuid.UUID) (*models.Invoice, error) {
	inv, err := s.repo.FindByID(id, lawyerID)
	if err != nil {
		return nil, errors.New("invoice not found")
	}
	if inv.Status != models.InvoiceStatusDraft {
		return nil, errors.New("only a draft invoice can be sent")
	}
	inv.Status = models.InvoiceStatusSent
	if err := s.repo.Update(inv); err != nil {
		return nil, err
	}
	s.enrichInvoice(inv)

	if client, err := s.clientRepo.FindByID(inv.ClientID, lawyerID); err == nil {
		if client.Email != "" {
			s.emailService.SendInvoiceEmail(
				client.Email, client.Name, inv.LawyerName, inv.InvoiceNumber,
				float64(inv.AmountPaise)/100, inv.Description,
			)
		}
		if client.AccountUserID != nil {
			_, _ = s.notificationService.Create(*client.AccountUserID, models.NotificationRequest{
				Title: fmt.Sprintf("Invoice %s received", inv.InvoiceNumber),
				Body:  fmt.Sprintf("%s sent you an invoice for ₹%.2f. Open it to pay via UPI.", inv.LawyerName, float64(inv.AmountPaise)/100),
				Type:  models.NotificationTypeGeneral,
			})
		}
	}

	return inv, nil
}

// SubmitPayment records the client's UPI Intent result. SUCCESS moves the
// invoice to "Payment Submitted" and stores whatever transaction reference
// the UPI app returned — this is NOT proof of payment, only the lawyer's
// manual MarkPaid confirms funds actually arrived. FAILED/CANCELLED leave
// the invoice exactly as it was (still "Pending"/Sent).
func (s *InvoiceService) SubmitPayment(id, clientID uuid.UUID, req models.SubmitPaymentRequest) (*models.Invoice, error) {
	inv, err := s.repo.FindByIDForClient(id, clientID)
	if err != nil {
		return nil, errors.New("invoice not found")
	}
	if inv.Status == models.InvoiceStatusPaid {
		return nil, errors.New("invoice is already paid")
	}
	if strings.EqualFold(req.Status, "SUCCESS") {
		inv.Status = models.InvoiceStatusSubmitted
		inv.TransactionRef = req.TransactionRef
		inv.TransactionID = req.TransactionID
		inv.PaidAmountPaise = req.PaidAmountPaise
		inv.PaymentApp = req.AppUsed
		now := time.Now()
		inv.PaymentDate = &now
		if err := s.repo.Update(inv); err != nil {
			return nil, err
		}

		// Create the Payment audit record — one per invoice, created the
		// moment the client's UPI Intent attempt returns. If a prior attempt
		// already created one (e.g. the client retried after a failed try),
		// refresh it instead of creating a second row for the same invoice.
		if existing, err := s.paymentRepo.FindByInvoiceID(inv.ID); err == nil {
			existing.AmountPaise = req.PaidAmountPaise
			existing.Status = models.PaymentStatusPendingVerification
			existing.TransactionRef = req.TransactionRef
			existing.TransactionID = req.TransactionID
			existing.AppUsed = req.AppUsed
			existing.AttemptedAt = now
			_ = s.paymentRepo.Update(existing)
		} else {
			_ = s.paymentRepo.Create(&models.Payment{
				InvoiceID:      inv.ID,
				LawyerID:       inv.LawyerID,
				ClientID:       inv.ClientID,
				AmountPaise:    req.PaidAmountPaise,
				Status:         models.PaymentStatusPendingVerification,
				TransactionRef: req.TransactionRef,
				TransactionID:  req.TransactionID,
				AppUsed:        req.AppUsed,
				AttemptedAt:    now,
			})
		}

		// Best-effort: let the lawyer know a payment attempt landed and
		// needs manual verification.
		_, _ = s.notificationService.Create(inv.LawyerID, models.NotificationRequest{
			Title: fmt.Sprintf("Payment submitted for %s", inv.InvoiceNumber),
			Body:  "A client payment attempt is pending your verification.",
			Type:  models.NotificationTypeGeneral,
		})
	} else {
		// FAILED/CANCELLED at the UPI-app layer (user backed out, insufficient
		// balance, etc.) — the invoice itself stays exactly as it was (still
		// Sent/"Pending"), but the attempt is still recorded so Payment
		// History shows what actually happened rather than nothing at all.
		now := time.Now()
		if existing, err := s.paymentRepo.FindByInvoiceID(inv.ID); err == nil {
			existing.Status = models.PaymentStatusFailed
			existing.TransactionRef = req.TransactionRef
			existing.TransactionID = req.TransactionID
			existing.AppUsed = req.AppUsed
			existing.AttemptedAt = now
			_ = s.paymentRepo.Update(existing)
		} else {
			_ = s.paymentRepo.Create(&models.Payment{
				InvoiceID:      inv.ID,
				LawyerID:       inv.LawyerID,
				ClientID:       inv.ClientID,
				AmountPaise:    req.PaidAmountPaise,
				Status:         models.PaymentStatusFailed,
				TransactionRef: req.TransactionRef,
				TransactionID:  req.TransactionID,
				AppUsed:        req.AppUsed,
				AttemptedAt:    now,
			})
		}
	}
	s.enrichInvoice(inv)
	return inv, nil
}

// MarkPaid is the lawyer's manual verification step, after confirming the
// money actually reached their bank account — never set automatically.
func (s *InvoiceService) MarkPaid(id, lawyerID uuid.UUID) (*models.Invoice, error) {
	inv, err := s.repo.FindByID(id, lawyerID)
	if err != nil {
		return nil, errors.New("invoice not found")
	}
	if inv.Status != models.InvoiceStatusSubmitted && inv.Status != models.InvoiceStatusSent {
		return nil, ErrNotSubmittedForPayment
	}
	inv.Status = models.InvoiceStatusPaid
	now := time.Now()
	inv.PaidAt = &now
	if err := s.repo.Update(inv); err != nil {
		return nil, err
	}
	s.enrichInvoice(inv)
	return inv, nil
}

// ErrInvalidPaymentStatus is returned by SetPaymentStatus for any value
// other than pending/paid/cancelled.
var ErrInvalidPaymentStatus = errors.New("status must be one of: pending, paid, cancelled")

// ErrUTRRequired is returned by SetPaymentStatus when marking an invoice
// Paid without a UTR (bank transaction reference) number — the lawyer's
// manual proof that funds actually landed in their account.
var ErrUTRRequired = errors.New("UTR (transaction reference) number is required to mark a payment as paid")

// SetPaymentStatus is the Payments module's lawyer-facing manual action —
// deliberately separate from the invoice's own content (Update above).
// Reuses this same Invoice row for fast-path display, and keeps the
// invoice's linked Payment row (audit trail: UTR, verified by, payment
// date) in sync.
func (s *InvoiceService) SetPaymentStatus(id, lawyerID uuid.UUID, status, utrNumber string) (*models.Invoice, error) {
	inv, err := s.repo.FindByID(id, lawyerID)
	if err != nil {
		return nil, errors.New("invoice not found")
	}

	payment, _ := s.paymentRepo.FindByInvoiceID(inv.ID)

	switch strings.ToLower(status) {
	case "pending":
		inv.Status = models.InvoiceStatusSent
		inv.PaidAt = nil
		if payment != nil {
			payment.Status = models.PaymentStatusPendingVerification
			_ = s.paymentRepo.Update(payment)
		}
	case "paid":
		utrNumber = strings.TrimSpace(utrNumber)
		if utrNumber == "" {
			return nil, ErrUTRRequired
		}
		inv.Status = models.InvoiceStatusPaid
		inv.UTRNumber = utrNumber
		now := time.Now()
		inv.PaidAt = &now
		if payment == nil {
			// The lawyer can verify a payment even if the client's UPI Intent
			// result never made it back to the app (e.g. app was killed) —
			// still create the audit row rather than silently skipping it.
			payment = &models.Payment{
				InvoiceID:   inv.ID,
				LawyerID:    inv.LawyerID,
				ClientID:    inv.ClientID,
				AmountPaise: inv.AmountPaise,
				AttemptedAt: now,
			}
		}
		payment.Status = models.PaymentStatusPaid
		payment.UTRNumber = utrNumber
		payment.VerifiedBy = &lawyerID
		payment.PaymentDate = &now
		if payment.ID == uuid.Nil {
			_ = s.paymentRepo.Create(payment)
		} else {
			_ = s.paymentRepo.Update(payment)
		}
	case "cancelled":
		inv.Status = models.InvoiceStatusCancelled
		if payment != nil {
			payment.Status = models.PaymentStatusCancelled
			_ = s.paymentRepo.Update(payment)
		}
	default:
		return nil, ErrInvalidPaymentStatus
	}
	if err := s.repo.Update(inv); err != nil {
		return nil, err
	}
	s.enrichInvoice(inv)

	if strings.ToLower(status) == "paid" {
		if client, err := s.clientRepo.FindByID(inv.ClientID, lawyerID); err == nil && client.AccountUserID != nil {
			_, _ = s.notificationService.Create(*client.AccountUserID, models.NotificationRequest{
				Title: fmt.Sprintf("Payment confirmed for %s", inv.InvoiceNumber),
				Body:  fmt.Sprintf("Your payment of ₹%.2f has been verified and marked Paid.", float64(inv.AmountPaise)/100),
				Type:  models.NotificationTypeGeneral,
			})
		}
	}

	return inv, nil
}

// GetPayment returns the Payment audit record for an invoice — lawyer-scoped.
func (s *InvoiceService) GetPayment(invoiceID, lawyerID uuid.UUID) (*models.Payment, error) {
	if _, err := s.repo.FindByID(invoiceID, lawyerID); err != nil {
		return nil, errors.New("invoice not found")
	}
	return s.paymentRepo.FindByInvoiceID(invoiceID)
}

// GetPaymentForClient returns the Payment audit record for an invoice —
// client-scoped.
func (s *InvoiceService) GetPaymentForClient(invoiceID, clientID uuid.UUID) (*models.Payment, error) {
	if _, err := s.repo.FindByIDForClient(invoiceID, clientID); err != nil {
		return nil, errors.New("invoice not found")
	}
	return s.paymentRepo.FindByInvoiceID(invoiceID)
}

// ListPaymentHistory is the lawyer's full Payment History across every
// client/invoice.
func (s *InvoiceService) ListPaymentHistory(lawyerID uuid.UUID) ([]models.Payment, error) {
	return s.paymentRepo.ListByLawyer(lawyerID)
}

// ListPaymentHistoryForClient is the client's own Payment History.
func (s *InvoiceService) ListPaymentHistoryForClient(clientID uuid.UUID) ([]models.Payment, error) {
	return s.paymentRepo.ListByClientID(clientID)
}

func (s *InvoiceService) Delete(id, lawyerID uuid.UUID) error {
	inv, err := s.repo.FindByID(id, lawyerID)
	if err != nil {
		return errors.New("invoice not found")
	}
	if inv.Status != models.InvoiceStatusDraft {
		return errors.New("only a draft invoice can be deleted")
	}
	return repositories.NewInvoiceRepository().DeleteDraft(id, lawyerID)
}
