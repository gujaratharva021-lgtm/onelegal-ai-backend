package handlers

import (
	"net/http"
	"time"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CalendarHandler struct {
	service *services.CalendarService
}

func NewCalendarHandler() *CalendarHandler {
	return &CalendarHandler{service: services.NewCalendarService()}
}

func (h *CalendarHandler) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var req models.CalendarEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	event, err := h.service.Create(userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create event")
		return
	}
	response.Success(c, http.StatusCreated, "Event created", event)
}

func (h *CalendarHandler) List(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	from := time.Now().AddDate(0, -1, 0)
	to := time.Now().AddDate(0, 2, 0)

	if v := c.Query("from"); v != "" {
		if parsed, err := time.Parse(time.RFC3339, v); err == nil {
			from = parsed
		}
	}
	if v := c.Query("to"); v != "" {
		if parsed, err := time.Parse(time.RFC3339, v); err == nil {
			to = parsed
		}
	}

	events, err := h.service.ListRange(userID, from, to)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch events")
		return
	}
	response.Success(c, http.StatusOK, "Events fetched", events)
}

func (h *CalendarHandler) Get(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	event, err := h.service.Get(id, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Event not found")
		return
	}
	response.Success(c, http.StatusOK, "Event fetched", event)
}

func (h *CalendarHandler) Update(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	var req models.CalendarEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	event, err := h.service.Update(id, userID, req)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Event not found")
		return
	}
	response.Success(c, http.StatusOK, "Event updated", event)
}

func (h *CalendarHandler) Delete(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid event id")
		return
	}
	if err := h.service.Delete(id, userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete event")
		return
	}
	response.Success(c, http.StatusOK, "Event deleted", nil)
}
