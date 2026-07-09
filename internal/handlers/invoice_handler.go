package handlers

import (
	"net/http"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// InvoiceHandler manages a lawyer's own invoices: create a draft, list/view
// them, send one (Draft -> Sent), generate its PDF, or delete a draft.
type InvoiceHandler struct {
	service *services.InvoiceService
}

func NewInvoiceHandler() *InvoiceHandler {
	return &InvoiceHandler{service: services.NewInvoiceService()}
}

func (h *InvoiceHandler) Create(c *gin.Context) {
	userID, _ := currentUserID(c)
	var req models.CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	inv, err := h.service.Create(userID, req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusCreated, "Invoice created", inv)
}

func (h *InvoiceHandler) List(c *gin.Context) {
	userID, _ := currentUserID(c)
	invoices, err := h.service.List(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch invoices")
		return
	}
	response.Success(c, http.StatusOK, "Invoices fetched", invoices)
}

func (h *InvoiceHandler) Get(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid invoice id")
		return
	}
	inv, err := h.service.Get(id, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Invoice not found")
		return
	}
	response.Success(c, http.StatusOK, "Invoice fetched", inv)
}

// Update edits an existing invoice's content — rejected once a payment has
// been submitted or confirmed (see InvoiceService.Update).
func (h *InvoiceHandler) Update(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid invoice id")
		return
	}
	var req models.UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	inv, err := h.service.Update(id, userID, req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Invoice updated", inv)
}

func (h *InvoiceHandler) Send(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid invoice id")
		return
	}
	inv, err := h.service.Send(id, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Invoice sent", inv)
}

// MarkPaid is the lawyer's manual "I've confirmed the money is in my bank
// account" action — payment is never auto-marked Paid from the client's UPI
// Intent result alone.
func (h *InvoiceHandler) MarkPaid(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid invoice id")
		return
	}
	inv, err := h.service.MarkPaid(id, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Invoice marked as paid", inv)
}

// SetPaymentStatus is the Payments module's manual lawyer action —
// Pending/Paid/Cancelled — separate from editing the invoice's own content.
func (h *InvoiceHandler) SetPaymentStatus(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid invoice id")
		return
	}
	var req models.SetPaymentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	inv, err := h.service.SetPaymentStatus(id, userID, req.Status, req.UTRNumber)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Payment status updated", inv)
}

// GetPayment returns the invoice's Payment audit record (UTR, verified by,
// payment date) — lawyer-scoped.
func (h *InvoiceHandler) GetPayment(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid invoice id")
		return
	}
	payment, err := h.service.GetPayment(id, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Payment not found")
		return
	}
	response.Success(c, http.StatusOK, "Payment fetched", payment)
}

// PaymentHistory is the lawyer's full Payment History across every invoice.
func (h *InvoiceHandler) PaymentHistory(c *gin.Context) {
	userID, _ := currentUserID(c)
	payments, err := h.service.ListPaymentHistory(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch payment history")
		return
	}
	response.Success(c, http.StatusOK, "Payment history fetched", payments)
}

func (h *InvoiceHandler) PDF(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid invoice id")
		return
	}
	inv, lawyer, client, err := h.service.PDFContext(id, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Invoice not found")
		return
	}
	writeInvoicePDF(c, inv, lawyer, client)
}

func (h *InvoiceHandler) Delete(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid invoice id")
		return
	}
	if err := h.service.Delete(id, userID); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Invoice deleted", nil)
}
