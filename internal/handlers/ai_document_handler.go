package handlers

import (
	"net/http"

	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type AIDocumentHandler struct {
	service *services.AIDocumentService
}

func NewAIDocumentHandler(service *services.AIDocumentService) *AIDocumentHandler {
	return &AIDocumentHandler{service: service}
}

func (h *AIDocumentHandler) Summarize(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "No file provided")
		return
	}
	analysis, err := h.service.UploadAndSummarize(userID, file)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusCreated, "Summary generated", analysis)
}

func (h *AIDocumentHandler) AnalyzeContract(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "No file provided")
		return
	}
	analysis, err := h.service.UploadAndAnalyzeContract(userID, file)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusCreated, "Contract analysis generated", analysis)
}

func (h *AIDocumentHandler) History(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	history, err := h.service.History(userID, c.Query("type"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch history")
		return
	}
	response.Success(c, http.StatusOK, "History fetched", history)
}
