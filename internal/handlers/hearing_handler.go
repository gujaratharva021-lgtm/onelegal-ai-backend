package handlers

import (
	"net/http"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type HearingHandler struct {
	service *services.HearingService
}

func NewHearingHandler() *HearingHandler {
	return &HearingHandler{service: services.NewHearingService()}
}

func (h *HearingHandler) Create(c *gin.Context) {
	userID, _ := currentUserID(c)
	var req models.HearingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	hearing, err := h.service.Create(userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create hearing")
		return
	}
	response.Success(c, http.StatusCreated, "Hearing created", hearing)
}

func (h *HearingHandler) List(c *gin.Context) {
	userID, _ := currentUserID(c)
	if c.Query("today") == "true" {
		hearings, err := h.service.ListToday(userID)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to fetch hearings")
			return
		}
		response.Success(c, http.StatusOK, "Today's hearings fetched", hearings)
		return
	}
	hearings, err := h.service.List(userID, c.Query("status"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch hearings")
		return
	}
	response.Success(c, http.StatusOK, "Hearings fetched", hearings)
}

func (h *HearingHandler) Get(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid hearing id")
		return
	}
	hearing, err := h.service.Get(id, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Hearing not found")
		return
	}
	response.Success(c, http.StatusOK, "Hearing fetched", hearing)
}

func (h *HearingHandler) Update(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid hearing id")
		return
	}
	var req models.HearingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	hearing, err := h.service.Update(id, userID, req)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Hearing not found")
		return
	}
	response.Success(c, http.StatusOK, "Hearing updated", hearing)
}

func (h *HearingHandler) Delete(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid hearing id")
		return
	}
	if err := h.service.Delete(id, userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete hearing")
		return
	}
	response.Success(c, http.StatusOK, "Hearing deleted", nil)
}
