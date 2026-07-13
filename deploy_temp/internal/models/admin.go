package models

import "time"

// AdminDashboardStats is the Admin Dashboard's top summary — every number is
// a live COUNT/SUM query against the existing tables, never mock data.
type AdminDashboardStats struct {
	TotalLawyers    int64 `json:"total_lawyers"`
	TotalClients    int64 `json:"total_clients"`
	TotalCases      int64 `json:"total_cases"`
	TotalMeetings   int64 `json:"total_meetings"`
	PendingPayments int64 `json:"pending_payments"`
	RevenuePaise    int64 `json:"revenue_paise"`
}

// AdminLawyerSummary is one row of Admin > Lawyer Management — a read-only
// cross-account view, reusing the existing User table (Role = advocate).
type AdminLawyerSummary struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	LawFirm     string    `json:"law_firm"`
	IsOnline    bool      `json:"is_online"`
	ClientCount int64     `json:"client_count"`
	CaseCount   int64     `json:"case_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// AdminClientSummary is one row of Admin > Client Management, reusing the
// existing Client (CRM) table across every lawyer.
type AdminClientSummary struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	Status        string    `json:"status"`
	AccountStatus string    `json:"account_status"`
	LawyerName    string    `json:"lawyer_name"`
	CreatedAt     time.Time `json:"created_at"`
}

// AdminCaseSummary is one row of Admin > Case Management, reusing the
// existing Case table across every lawyer.
type AdminCaseSummary struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	CaseNumber string    `json:"case_number"`
	Status     string    `json:"status"`
	Priority   string    `json:"priority"`
	LawyerName string    `json:"lawyer_name"`
	ClientName string    `json:"client_name"`
	CreatedAt  time.Time `json:"created_at"`
}

// AdminPaymentSummary is one row of Admin > Payments, reusing the existing
// Invoice table across every lawyer — no separate payment table.
type AdminPaymentSummary struct {
	ID            string     `json:"id"`
	InvoiceNumber string     `json:"invoice_number"`
	AmountPaise   int64      `json:"amount_paise"`
	Status        string     `json:"status"`
	LawyerName    string     `json:"lawyer_name"`
	ClientName    string     `json:"client_name"`
	PaymentDate   *time.Time `json:"payment_date"`
	CreatedAt     time.Time  `json:"created_at"`
}
