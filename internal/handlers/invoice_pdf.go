package handlers

import (
	"fmt"

	"legaltech-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

// writeInvoicePDF renders the invoice PDF straight to the response: company
// (law firm) header, lawyer details, client details, invoice metadata, GST
// (when the lawyer has set one), amount, and a UPI QR placeholder block —
// all sourced from the existing Invoice/User/Client rows, no new model.
func writeInvoicePDF(c *gin.Context, inv *models.Invoice, lawyer *models.User, client *models.Client) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// --- Header: company/law firm name stands in for a logo image ---
	companyName := lawyer.LawFirm
	if companyName == "" {
		companyName = lawyer.Name
	}
	pdf.SetFont("Arial", "B", 18)
	pdf.CellFormat(0, 10, companyName, "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	if lawyer.Bio != "" {
		pdf.CellFormat(0, 6, lawyer.Bio, "", 1, "L", false, 0, "")
	}
	pdf.Ln(2)
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(6)

	pdf.SetFont("Arial", "B", 20)
	pdf.CellFormat(0, 10, "INVOICE", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	// --- Lawyer details ---
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(0, 7, "From (Advocate)", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 6, lawyer.Name, "", 1, "L", false, 0, "")
	if lawyer.Email != "" {
		pdf.CellFormat(0, 6, lawyer.Email, "", 1, "L", false, 0, "")
	}
	if lawyer.Phone != "" {
		pdf.CellFormat(0, 6, lawyer.Phone, "", 1, "L", false, 0, "")
	}
	if lawyer.BarNumber != "" {
		pdf.CellFormat(0, 6, fmt.Sprintf("Bar Number: %s", lawyer.BarNumber), "", 1, "L", false, 0, "")
	}
	if lawyer.GSTNumber != "" {
		pdf.CellFormat(0, 6, fmt.Sprintf("GSTIN: %s", lawyer.GSTNumber), "", 1, "L", false, 0, "")
	}
	pdf.Ln(4)

	// --- Client details ---
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(0, 7, "Bill To (Client)", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 6, client.Name, "", 1, "L", false, 0, "")
	if client.Email != "" {
		pdf.CellFormat(0, 6, client.Email, "", 1, "L", false, 0, "")
	}
	if client.Phone != "" {
		pdf.CellFormat(0, 6, client.Phone, "", 1, "L", false, 0, "")
	}
	pdf.Ln(4)

	// --- Invoice metadata ---
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(0, 7, "Invoice Details", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 6, fmt.Sprintf("Invoice Number: %s", inv.InvoiceNumber), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, fmt.Sprintf("Invoice Date: %s", inv.CreatedAt.Format("Jan 2, 2006")), "", 1, "L", false, 0, "")
	if inv.DueDate != nil {
		pdf.CellFormat(0, 6, fmt.Sprintf("Due Date: %s", inv.DueDate.Format("Jan 2, 2006")), "", 1, "L", false, 0, "")
	}
	pdf.CellFormat(0, 6, fmt.Sprintf("Status: %s", inv.Status), "", 1, "L", false, 0, "")
	if inv.PaidAt != nil {
		pdf.CellFormat(0, 6, fmt.Sprintf("Paid On: %s", inv.PaidAt.Format("Jan 2, 2006")), "", 1, "L", false, 0, "")
	}
	pdf.Ln(4)

	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(0, 7, "Description", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(0, 6, inv.Description, "", "L", false)
	pdf.Ln(2)

	if lawyer.GSTNumber != "" {
		// 18% GST shown as an illustrative line item — the invoiced amount
		// itself is treated as GST-inclusive since no separate tax rate is
		// configurable in the invoice model.
		gstAmount := float64(inv.AmountPaise) / 100 * 18 / 118
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(0, 6, fmt.Sprintf("Includes GST (18%%): %s %.2f", inv.Currency, gstAmount), "", 1, "L", false, 0, "")
	}

	pdf.SetFont("Arial", "B", 14)
	pdf.Ln(2)
	pdf.CellFormat(0, 10, fmt.Sprintf("Total Amount: %s %.2f", inv.Currency, float64(inv.AmountPaise)/100), "", 1, "L", false, 0, "")
	pdf.Ln(6)

	// --- Payment / QR placeholder ---
	if inv.LawyerUpiID != "" {
		pdf.SetDrawColor(180, 180, 180)
		qrY := pdf.GetY()
		pdf.Rect(10, qrY, 30, 30, "D")
		pdf.SetFont("Arial", "", 8)
		pdf.SetXY(10, qrY+13)
		pdf.CellFormat(30, 4, "Scan to Pay (UPI)", "", 0, "C", false, 0, "")
		pdf.SetXY(45, qrY+8)
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(0, 6, fmt.Sprintf("Pay via UPI: %s", inv.LawyerUpiID), "", 1, "L", false, 0, "")
		pdf.SetY(qrY + 34)
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=invoice-%s.pdf", inv.InvoiceNumber))
	_ = pdf.Output(c.Writer)
}
