package handlers

import (
	"net/http"

	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	service *services.DashboardService
}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{service: services.NewDashboardService()}
}

func (h *DashboardHandler) Get(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	dashboard, err := h.service.Get(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load dashboard")
		return
	}
	response.Success(c, http.StatusOK, "Dashboard fetched", dashboard)
}
