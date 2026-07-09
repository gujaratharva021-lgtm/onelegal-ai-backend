package services

import (
	"fmt"
	"log"
	"net/smtp"

	"legaltech-backend/internal/config"
)

type EmailService struct {
	cfg *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{cfg: cfg}
}

func (s *EmailService) SendPasswordResetEmail(toEmail, toName, code string) {
	subject := "Reset your LegalTech AI password"
	body := fmt.Sprintf(`<div style="font-family:Arial,sans-serif;max-width:480px;margin:0 auto;padding:24px;">
<h2 style="color:#0A1F44;">Password Reset Request</h2>
<p>Hello %s,</p>
<p>We received a request to reset the password for your LegalTech AI account. Use the code below in the app to set a new password. This code expires in 15 minutes.</p>
<div style="background:#0A1F44;color:#D4AF37;font-size:28px;letter-spacing:6px;font-weight:bold;text-align:center;padding:16px;border-radius:8px;margin:24px 0;">%s</div>
<p>If you did not request this, you can safely ignore this email.</p>
<p style="color:#888;font-size:12px;">LegalTech AI Legal Assistant</p>
</div>`, toName, code)

	if err := s.send(toEmail, subject, body); err != nil {
		log.Printf("email_service: failed to send password reset email to %s: %v", toEmail, err)
	}
}

// SendInvoiceEmail notifies the client that a new invoice is ready to pay —
// sent once, right when the lawyer transitions an invoice Draft -> Sent (see
// InvoiceService.Send). Best-effort: a missing/invalid client email or SMTP
// failure is logged and swallowed, matching SendPasswordResetEmail's
// convention — it never blocks the Send action itself.
func (s *EmailService) SendInvoiceEmail(toEmail, clientName, lawyerName, invoiceNumber string, amountRupees float64, description string) {
	subject := fmt.Sprintf("Invoice %s from %s", invoiceNumber, lawyerName)
	body := fmt.Sprintf(`<div style="font-family:Arial,sans-serif;max-width:480px;margin:0 auto;padding:24px;">
<h2 style="color:#0A1F44;">New Invoice</h2>
<p>Hello %s,</p>
<p>%s has issued you a new invoice via LegalTech AI.</p>
<table style="width:100%%;border-collapse:collapse;margin:20px 0;">
<tr><td style="padding:8px 0;color:#888;">Invoice Number</td><td style="padding:8px 0;text-align:right;font-weight:bold;">%s</td></tr>
<tr><td style="padding:8px 0;color:#888;">Description</td><td style="padding:8px 0;text-align:right;">%s</td></tr>
<tr><td style="padding:8px 0;color:#888;border-top:1px solid #eee;">Amount Due</td><td style="padding:8px 0;text-align:right;font-weight:bold;font-size:18px;color:#0A1F44;border-top:1px solid #eee;">&#8377;%.2f</td></tr>
</table>
<p>Open the LegalTech AI app, go to My Invoices, and tap <b>Pay Now</b> to pay instantly via UPI (Google Pay, PhonePe, Paytm, BHIM, or Amazon Pay).</p>
<p style="color:#888;font-size:12px;">LegalTech AI Legal Assistant</p>
</div>`, clientName, lawyerName, invoiceNumber, description, amountRupees)

	if err := s.send(toEmail, subject, body); err != nil {
		log.Printf("email_service: failed to send invoice email to %s: %v", toEmail, err)
	}
}

func (s *EmailService) send(to, subject, body string) error {
	if s.cfg == nil || s.cfg.SMTPHost == "" || s.cfg.SMTPUser == "" {
		return fmt.Errorf("smtp is not configured: set SMTP_HOST, SMTP_USER, SMTP_PASSWORD in .env")
	}

	addr := fmt.Sprintf("%s:%s", s.cfg.SMTPHost, s.cfg.SMTPPort)
	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n%s",
		s.cfg.SMTPFrom, to, subject, body,
	))

	return smtp.SendMail(addr, auth, s.cfg.SMTPFrom, []string{to}, msg)
}
