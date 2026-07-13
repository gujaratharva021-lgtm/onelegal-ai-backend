package handlers

import (
	"errors"
	"net/http"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VideoRoomHandler struct {
	service *services.VideoRoomService
}

func NewVideoRoomHandler(service *services.VideoRoomService) *VideoRoomHandler {
	return &VideoRoomHandler{service: service}
}

// StartCall — lawyer presses "Start Meeting" on a scheduled meeting.
func (h *VideoRoomHandler) StartCall(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var req models.StartMeetingCallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	room, clientName, delivered, err := h.service.StartCall(userID, req.MeetingID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusCreated, "Meeting call started", gin.H{
		"room":          room,
		"client_name":   clientName,
		"client_online": delivered,
	})
}

// ActiveForMe — the client-side check performed on login/app open.
func (h *VideoRoomHandler) ActiveForMe(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	room, lawyerName, err := h.service.ActiveForClient(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Success(c, http.StatusOK, "No active meeting", gin.H{"room": nil})
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to check for an active meeting")
		return
	}
	response.Success(c, http.StatusOK, "Active meeting found", gin.H{
		"room":        room,
		"lawyer_name": lawyerName,
	})
}

func (h *VideoRoomHandler) Join(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid room id")
		return
	}
	room, lawyerName, err := h.service.Join(id, userID)
	if err != nil {
		response.Error(c, http.StatusForbidden, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Joined meeting", gin.H{
		"room":        room,
		"lawyer_name": lawyerName,
	})
}

// End covers hangup, decline (reason:"rejected"), and a client-side ring
// timeout (reason:"missed") — all the same endpoint, since the meaningful
// distinction is just whether the room was ever accepted (see
// VideoRoomService.End).
func (h *VideoRoomHandler) End(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid room id")
		return
	}
	var req models.EndCallRequest
	_ = c.ShouldBindJSON(&req) // optional body — a plain hangup sends none

	room, err := h.service.End(id, userID, req.Reason)
	if err != nil {
		response.Error(c, http.StatusForbidden, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Meeting ended", room)
}
