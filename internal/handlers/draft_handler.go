package handlers

import (
	"net/http"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DraftHandler struct {
	service *services.DraftService
}

func NewDraftHandler(service *services.DraftService) *DraftHandler {
	return &DraftHandler{service: service}
}

func (h *DraftHandler) Create(c *gin.Context) {
	userID, _ := currentUserID(c)
	var req models.DraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	draft, err := h.service.Create(userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create draft")
		return
	}
	response.Success(c, http.StatusCreated, "Draft created", draft)
}

func (h *DraftHandler) GenerateWithAI(c *gin.Context) {
	userID, _ := currentUserID(c)
	var req models.AIDraftGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	draft, err := h.service.GenerateWithAI(userID, req)
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}
	response.Success(c, http.StatusCreated, "Draft generated", draft)
}

func (h *DraftHandler) List(c *gin.Context) {
	userID, _ := currentUserID(c)
	drafts, err := h.service.List(userID, c.Query("type"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch drafts")
		return
	}
	response.Success(c, http.StatusOK, "Drafts fetched", drafts)
}

func (h *DraftHandler) Get(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid draft id")
		return
	}
	draft, err := h.service.Get(id, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Draft not found")
		return
	}
	response.Success(c, http.StatusOK, "Draft fetched", draft)
}

func (h *DraftHandler) Update(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid draft id")
		return
	}
	var req models.DraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	draft, err := h.service.Update(id, userID, req)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Draft not found")
		return
	}
	response.Success(c, http.StatusOK, "Draft updated", draft)
}

func (h *DraftHandler) Delete(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid draft id")
		return
	}
	if err := h.service.Delete(id, userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete draft")
		return
	}
	response.Success(c, http.StatusOK, "Draft deleted", nil)
}
