package handlers

import (
	"net/http"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	service *services.NotificationService
}

func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{service: services.NewNotificationService()}
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	unreadOnly := c.Query("unread") == "true"
	notifications, err := h.service.List(userID, unreadOnly)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch notifications")
		return
	}
	count, _ := h.service.CountUnread(userID)
	response.Success(c, http.StatusOK, "Notifications fetched", gin.H{
		"notifications": notifications,
		"unread_count":  count,
	})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid notification id")
		return
	}
	if err := h.service.MarkRead(id, userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update notification")
		return
	}
	response.Success(c, http.StatusOK, "Notification marked as read", nil)
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if err := h.service.MarkAllRead(userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update notifications")
		return
	}
	response.Success(c, http.StatusOK, "All notifications marked as read", nil)
}

func (h *NotificationHandler) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var req models.NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	n, err := h.service.Create(userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create notification")
		return
	}
	response.Success(c, http.StatusCreated, "Notification created", n)
}

func (h *NotificationHandler) Update(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid notification id")
		return
	}
	var req models.NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	n, err := h.service.Update(id, userID, req)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Notification not found")
		return
	}
	response.Success(c, http.StatusOK, "Notification updated", n)
}

func (h *NotificationHandler) Delete(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid notification id")
		return
	}
	if err := h.service.Delete(id, userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete notification")
		return
	}
	response.Success(c, http.StatusOK, "Notification deleted", nil)
}
