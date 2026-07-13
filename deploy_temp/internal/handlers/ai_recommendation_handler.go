package handlers

import (
	"net/http"

	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AIRecommendationHandler struct {
	service *services.AIRecommendationService
}

func NewAIRecommendationHandler(service *services.AIRecommendationService) *AIRecommendationHandler {
	return &AIRecommendationHandler{service: service}
}

func (h *AIRecommendationHandler) Generate(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid case id")
		return
	}
	rec, err := h.service.Generate(userID, caseID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusCreated, "Recommendation generated", rec)
}

func (h *AIRecommendationHandler) List(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid case id")
		return
	}
	recs, err := h.service.List(caseID, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Case not found")
		return
	}
	response.Success(c, http.StatusOK, "Recommendations fetched", recs)
}
