package handlers

import (
	"net/http"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CaseTimelineHandler struct {
	service *services.CaseTimelineService
}

func NewCaseTimelineHandler() *CaseTimelineHandler {
	return &CaseTimelineHandler{service: services.NewCaseTimelineService()}
}

func (h *CaseTimelineHandler) Create(c *gin.Context) {
	userID, _ := currentUserID(c)
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid case id")
		return
	}
	var req models.CaseTimelineEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	event, err := h.service.Create(caseID, userID, req)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Case not found")
		return
	}
	response.Success(c, http.StatusCreated, "Timeline event created", event)
}

func (h *CaseTimelineHandler) List(c *gin.Context) {
	userID, _ := currentUserID(c)
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid case id")
		return
	}
	events, err := h.service.List(caseID, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Case not found")
		return
	}
	response.Success(c, http.StatusOK, "Timeline fetched", events)
}

func (h *CaseTimelineHandler) Update(c *gin.Context) {
	userID, _ := currentUserID(c)
	eventID, err := uuid.Parse(c.Param("eventId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	var req models.CaseTimelineEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	event, err := h.service.Update(eventID, userID, req)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Timeline event not found")
		return
	}
	response.Success(c, http.StatusOK, "Timeline event updated", event)
}
