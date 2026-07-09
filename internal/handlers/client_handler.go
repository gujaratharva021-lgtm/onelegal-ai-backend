package handlers

import (
	"errors"
	"net/http"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ClientHandler struct {
	service         *services.ClientService
	documentService *services.DocumentService
	meetingService  *services.MeetingService
	invoiceService  *services.InvoiceService
}

func NewClientHandler() *ClientHandler {
	return &ClientHandler{
		service:         services.NewClientService(),
		documentService: services.NewDocumentService(),
		meetingService:  services.NewMeetingService(),
		invoiceService:  services.NewInvoiceService(),
	}
}

func (h *ClientHandler) Create(c *gin.Context) {
	userID, _ := currentUserID(c)
	var req models.ClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	client, err := h.service.Create(userID, req)
	if err != nil {
		if errors.Is(err, services.ErrDuplicateEmail) || errors.Is(err, services.ErrDuplicatePhone) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, services.ErrLoginIDRequired) ||
			errors.Is(err, services.ErrTemporaryPasswordRequired) ||
			errors.Is(err, services.ErrLoginIDTaken) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to create client")
		return
	}
	response.Success(c, http.StatusCreated, "Client account created successfully", client)
}

func (h *ClientHandler) List(c *gin.Context) {
	userID, _ := currentUserID(c)
	clients, err := h.service.List(userID, c.Query("search"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch clients")
		return
	}
	response.Success(c, http.StatusOK, "Clients fetched", clients)
}

func (h *ClientHandler) Get(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid client id")
		return
	}
	client, err := h.service.Get(id, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Client not found")
		return
	}
	response.Success(c, http.StatusOK, "Client fetched", client)
}

func (h *ClientHandler) Update(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid client id")
		return
	}
	var req models.ClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	client, err := h.service.Update(id, userID, req)
	if err != nil {
		if errors.Is(err, services.ErrDuplicateEmail) || errors.Is(err, services.ErrDuplicatePhone) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusNotFound, "Client not found")
		return
	}
	response.Success(c, http.StatusOK, "Client updated", client)
}

func (h *ClientHandler) Archive(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid client id")
		return
	}
	client, err := h.service.Archive(id, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Client not found")
		return
	}
	response.Success(c, http.StatusOK, "Client archived", client)
}

// ListDocuments returns every document attached to this client, either
// directly or via one of their cases — the client profile's Documents tab.
func (h *ClientHandler) ListDocuments(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid client id")
		return
	}
	if _, err := h.service.Get(id, userID); err != nil {
		response.Error(c, http.StatusNotFound, "Client not found")
		return
	}
	docs, err := h.documentService.ListForClient(userID, id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch documents")
		return
	}
	response.Success(c, http.StatusOK, "Documents fetched", docs)
}

// ListMeetings returns every meeting scheduled with this client — the
// client profile's Meetings tab.
func (h *ClientHandler) ListMeetings(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid client id")
		return
	}
	if _, err := h.service.Get(id, userID); err != nil {
		response.Error(c, http.StatusNotFound, "Client not found")
		return
	}
	meetings, err := h.meetingService.ListForClient(userID, id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch meetings")
		return
	}
	response.Success(c, http.StatusOK, "Meetings fetched", meetings)
}

// ListInvoices returns every invoice issued to this client — the client
// profile's Billing tab.
func (h *ClientHandler) ListInvoices(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid client id")
		return
	}
	if _, err := h.service.Get(id, userID); err != nil {
		response.Error(c, http.StatusNotFound, "Client not found")
		return
	}
	invoices, err := h.invoiceService.ListForClient(userID, id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch invoices")
		return
	}
	response.Success(c, http.StatusOK, "Invoices fetched", invoices)
}

// ResetPassword — lawyer-only "Reset Client Password" action in the Client
// Profile screen. The client's Login ID is unchanged; only the password
// rotates, and the client is notified.
func (h *ClientHandler) ResetPassword(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid client id")
		return
	}
	var req models.ResetClientPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	client, err := h.service.ResetPassword(id, userID, req.NewPassword)
	if err != nil {
		if errors.Is(err, services.ErrNoLinkedAccount) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(c, http.StatusNotFound, "Client not found")
		return
	}
	response.Success(c, http.StatusOK, "Client password reset successfully", client)
}

func (h *ClientHandler) Delete(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid client id")
		return
	}
	if err := h.service.Delete(id, userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete client")
		return
	}
	response.Success(c, http.StatusOK, "Client deleted", nil)
}
