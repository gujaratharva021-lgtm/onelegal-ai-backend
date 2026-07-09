package handlers

import (
	"net/http"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"
	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ClientPortalHandler serves the read-only client-facing dashboard (My
// Cases, My Documents, My Meetings, My Invoices, My Profile), plus the
// client's UPI Intent "Pay Now" actions on their own invoices. Every method
// resolves "my own data" server-side from the authenticated client's linked
// Client record — the client never supplies a client id, so there is no way
// for one client to request another's data. Routes are additionally gated
// to Role == "client" by middleware.RequireRole.
type ClientPortalHandler struct {
	clientRepo     *repositories.ClientRepository
	caseRepo       *repositories.CaseRepository
	documentRepo   *repositories.DocumentRepository
	meetingRepo    *repositories.MeetingRepository
	invoiceService *services.InvoiceService
}

func NewClientPortalHandler() *ClientPortalHandler {
	return &ClientPortalHandler{
		clientRepo:     repositories.NewClientRepository(),
		caseRepo:       repositories.NewCaseRepository(),
		documentRepo:   repositories.NewDocumentRepository(),
		meetingRepo:    repositories.NewMeetingRepository(),
		invoiceService: services.NewInvoiceService(),
	}
}

// myClient resolves the Client CRM record linked to the currently
// authenticated client-role user. Every other handler in this file calls
// this first and 404s if it fails, rather than trusting any client-supplied id.
func (h *ClientPortalHandler) myClient(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := currentUserID(c)
	if !exists {
		return uuid.UUID{}, false
	}
	client, err := h.clientRepo.FindByAccountUserID(userID)
	if err != nil {
		return uuid.UUID{}, false
	}
	return client.ID, true
}

func (h *ClientPortalHandler) Profile(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	client, err := h.clientRepo.FindByAccountUserID(userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Client profile not found")
		return
	}
	response.Success(c, http.StatusOK, "Profile fetched", client)
}

func (h *ClientPortalHandler) Cases(c *gin.Context) {
	clientID, ok := h.myClient(c)
	if !ok {
		response.Error(c, http.StatusNotFound, "Client profile not found")
		return
	}
	cases, err := h.caseRepo.ListByClientID(clientID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch cases")
		return
	}
	response.Success(c, http.StatusOK, "Cases fetched", cases)
}

func (h *ClientPortalHandler) Documents(c *gin.Context) {
	clientID, ok := h.myClient(c)
	if !ok {
		response.Error(c, http.StatusNotFound, "Client profile not found")
		return
	}
	docs, err := h.documentRepo.ListByClientID(clientID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch documents")
		return
	}
	response.Success(c, http.StatusOK, "Documents fetched", docs)
}

func (h *ClientPortalHandler) Meetings(c *gin.Context) {
	clientID, ok := h.myClient(c)
	if !ok {
		response.Error(c, http.StatusNotFound, "Client profile not found")
		return
	}
	meetings, err := h.meetingRepo.ListByClientID(clientID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch meetings")
		return
	}
	response.Success(c, http.StatusOK, "Meetings fetched", meetings)
}

func (h *ClientPortalHandler) Invoices(c *gin.Context) {
	clientID, ok := h.myClient(c)
	if !ok {
		response.Error(c, http.StatusNotFound, "Client profile not found")
		return
	}
	invoices, err := h.invoiceService.ListForClientPortal(clientID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch invoices")
		return
	}
	response.Success(c, http.StatusOK, "Invoices fetched", invoices)
}

// GetInvoice is used by the client's Invoice Details / Pay Now screen to
// refresh a single invoice (e.g. after submitting a payment result).
func (h *ClientPortalHandler) GetInvoice(c *gin.Context) {
	clientID, ok := h.myClient(c)
	if !ok {
		response.Error(c, http.StatusNotFound, "Client profile not found")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid invoice id")
		return
	}
	inv, err := h.invoiceService.GetForClient(id, clientID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Invoice not found")
		return
	}
	response.Success(c, http.StatusOK, "Invoice fetched", inv)
}

// SubmitInvoicePayment records the result of the Android UPI Intent right
// after it returns control to the app. SUCCESS moves the invoice to
// "Payment Submitted" (never straight to Paid); FAILED/CANCELLED leave it
// exactly as it was.
func (h *ClientPortalHandler) SubmitInvoicePayment(c *gin.Context) {
	clientID, ok := h.myClient(c)
	if !ok {
		response.Error(c, http.StatusNotFound, "Client profile not found")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid invoice id")
		return
	}
	var req models.SubmitPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	inv, err := h.invoiceService.SubmitPayment(id, clientID, req)
	if err != nil {
		if err.Error() == "invoice is already paid" {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusNotFound, "Invoice not found")
		return
	}
	response.Success(c, http.StatusOK, "Payment result recorded", inv)
}

// GetPayment returns the invoice's Payment audit record (status, UTR once
// verified, payment date) — client-scoped.
func (h *ClientPortalHandler) GetPayment(c *gin.Context) {
	clientID, ok := h.myClient(c)
	if !ok {
		response.Error(c, http.StatusNotFound, "Client profile not found")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid invoice id")
		return
	}
	payment, err := h.invoiceService.GetPaymentForClient(id, clientID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Payment not found")
		return
	}
	response.Success(c, http.StatusOK, "Payment fetched", payment)
}

// PaymentHistory is the client's own Payment History across every invoice.
func (h *ClientPortalHandler) PaymentHistory(c *gin.Context) {
	clientID, ok := h.myClient(c)
	if !ok {
		response.Error(c, http.StatusNotFound, "Client profile not found")
		return
	}
	payments, err := h.invoiceService.ListPaymentHistoryForClient(clientID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch payment history")
		return
	}
	response.Success(c, http.StatusOK, "Payment history fetched", payments)
}
