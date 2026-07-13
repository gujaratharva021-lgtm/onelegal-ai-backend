package handlers

import (
	"net/http"

	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type OCRHandler struct {
	service *services.OCRService
}

func NewOCRHandler(service *services.OCRService) *OCRHandler {
	return &OCRHandler{service: service}
}

func (h *OCRHandler) Extract(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "No file provided")
		return
	}

	text, err := h.service.Extract(file)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Text extracted", gin.H{"text": text})
}
