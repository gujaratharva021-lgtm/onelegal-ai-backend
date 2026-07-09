package handlers

import (
	"net/http"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MeetingHandler struct {
	service *services.MeetingService
}

func NewMeetingHandler() *MeetingHandler {
	return &MeetingHandler{service: services.NewMeetingService()}
}

func (h *MeetingHandler) Create(c *gin.Context) {
	userID, _ := currentUserID(c)
	var req models.MeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	meeting, err := h.service.Create(userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create meeting")
		return
	}
	response.Success(c, http.StatusCreated, "Meeting created", meeting)
}

func (h *MeetingHandler) ListUpcoming(c *gin.Context) {
	userID, _ := currentUserID(c)
	meetings, err := h.service.ListUpcoming(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch meetings")
		return
	}
	response.Success(c, http.StatusOK, "Upcoming meetings fetched", meetings)
}

func (h *MeetingHandler) ListHistory(c *gin.Context) {
	userID, _ := currentUserID(c)
	meetings, err := h.service.ListHistory(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch meeting history")
		return
	}
	response.Success(c, http.StatusOK, "Meeting history fetched", meetings)
}

func (h *MeetingHandler) Get(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid meeting id")
		return
	}
	meeting, err := h.service.Get(id, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Meeting not found")
		return
	}
	response.Success(c, http.StatusOK, "Meeting fetched", meeting)
}

func (h *MeetingHandler) Update(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid meeting id")
		return
	}
	var req models.MeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	meeting, err := h.service.Update(id, userID, req)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Meeting not found")
		return
	}
	response.Success(c, http.StatusOK, "Meeting updated", meeting)
}

func (h *MeetingHandler) Join(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid meeting id")
		return
	}
	meeting, err := h.service.UpdateStatus(id, userID, models.MeetingStatusOngoing)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Meeting not found")
		return
	}
	response.Success(c, http.StatusOK, "Meeting joined", meeting)
}

func (h *MeetingHandler) End(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid meeting id")
		return
	}
	meeting, err := h.service.UpdateStatus(id, userID, models.MeetingStatusCompleted)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Meeting not found")
		return
	}
	response.Success(c, http.StatusOK, "Meeting ended", meeting)
}

// Cancel is a distinct, non-destructive alternative to Delete — the meeting
// stays visible in history with Status = Cancelled instead of being removed.
func (h *MeetingHandler) Cancel(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid meeting id")
		return
	}
	meeting, err := h.service.Cancel(id, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Meeting not found")
		return
	}
	response.Success(c, http.StatusOK, "Meeting cancelled", meeting)
}

func (h *MeetingHandler) Delete(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid meeting id")
		return
	}
	if err := h.service.Delete(id, userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete meeting")
		return
	}
	response.Success(c, http.StatusOK, "Meeting deleted", nil)
}
