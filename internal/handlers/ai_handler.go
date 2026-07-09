package handlers

import (
	"net/http"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AIHandler struct {
	service *services.AIService
}

func NewAIHandler(service *services.AIService) *AIHandler {
	return &AIHandler{service: service}
}

func (h *AIHandler) Chat(c *gin.Context) {
	userID, _ := currentUserID(c)
	var req models.AIChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	reply, err := h.service.Chat(userID, req)
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Reply generated", reply)
}

func (h *AIHandler) ListConversations(c *gin.Context) {
	userID, _ := currentUserID(c)
	conversations, err := h.service.ListConversations(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch conversations")
		return
	}
	response.Success(c, http.StatusOK, "Conversations fetched", conversations)
}

func (h *AIHandler) GetMessages(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid conversation id")
		return
	}
	messages, err := h.service.ListMessages(id, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Messages fetched", messages)
}

func (h *AIHandler) DeleteConversation(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid conversation id")
		return
	}
	if err := h.service.DeleteConversation(id, userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete conversation")
		return
	}
	response.Success(c, http.StatusOK, "Conversation deleted", nil)
}

func (h *AIHandler) Summarize(c *gin.Context) {
	var req models.AISummarizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	summary, err := h.service.Summarize(req.Text)
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Summary generated", models.AISummarizeResponse{Summary: summary})
}
