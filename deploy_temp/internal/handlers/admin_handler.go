package handlers

import (
	"net/http"

	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdminHandler serves the hidden Admin Dashboard. Every route this backs is
// registered under a route group gated by middleware.RequireRole("admin") —
// see routes.go — so a lawyer or client JWT gets 403 Forbidden here even if
// they somehow guess the URL.
type AdminHandler struct {
	service *services.AdminService
}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{service: services.NewAdminService()}
}

func (h *AdminHandler) DashboardStats(c *gin.Context) {
	stats, err := h.service.DashboardStats()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load dashboard stats")
		return
	}
	response.Success(c, http.StatusOK, "Dashboard stats fetched successfully", stats)
}

func (h *AdminHandler) Lawyers(c *gin.Context) {
	lawyers, err := h.service.ListLawyers()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load lawyers")
		return
	}
	response.Success(c, http.StatusOK, "Lawyers fetched successfully", lawyers)
}

func (h *AdminHandler) Clients(c *gin.Context) {
	clients, err := h.service.ListClients()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load clients")
		return
	}
	response.Success(c, http.StatusOK, "Clients fetched successfully", clients)
}

func (h *AdminHandler) Cases(c *gin.Context) {
	cases, err := h.service.ListCases()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load cases")
		return
	}
	response.Success(c, http.StatusOK, "Cases fetched successfully", cases)
}

func (h *AdminHandler) Payments(c *gin.Context) {
	payments, err := h.service.ListPayments()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load payments")
		return
	}
	response.Success(c, http.StatusOK, "Payments fetched successfully", payments)
}
