package handlers

import (
	"net/http"

	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type DashboardStatsHandler struct {
	service *services.DashboardStatsService
}

func NewDashboardStatsHandler() *DashboardStatsHandler {
	return &DashboardStatsHandler{service: services.NewDashboardStatsService()}
}

func (h *DashboardStatsHandler) Get(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	stats, err := h.service.Get(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load statistics")
		return
	}
	response.Success(c, http.StatusOK, "Statistics fetched", stats)
}
