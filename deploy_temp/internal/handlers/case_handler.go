package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"
	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CaseHandler struct {
	service *services.CaseService
}

func NewCaseHandler() *CaseHandler {
	return &CaseHandler{service: services.NewCaseService()}
}

func (h *CaseHandler) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var req models.CaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.service.Create(userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create case")
		return
	}
	response.Success(c, http.StatusCreated, "Case created", result)
}

func (h *CaseHandler) List(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	page := 0
	if v := c.Query("page"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid page")
			return
		}
		page = parsed
	}
	pageSize := 0
	if v := c.Query("page_size"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid page_size")
			return
		}
		pageSize = parsed
	}
	result, err := h.service.List(userID, repositories.CaseListFilter{
		Search:   c.Query("search"),
		Status:   c.Query("status"),
		ClientID: c.Query("client_id"),
		Priority: c.Query("priority"),
		Court:    c.Query("court"),
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch cases")
		return
	}
	response.Success(c, http.StatusOK, "Cases fetched", result)
}

func (h *CaseHandler) Get(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid case id")
		return
	}
	result, err := h.service.Get(id, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Case not found")
		return
	}
	response.Success(c, http.StatusOK, "Case fetched", result)
}

func (h *CaseHandler) Update(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid case id")
		return
	}
	var req models.CaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.service.Update(id, userID, req)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Case not found")
		return
	}
	response.Success(c, http.StatusOK, "Case updated", result)
}

func (h *CaseHandler) Delete(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid case id")
		return
	}
	if err := h.service.Delete(id, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, "Case not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to delete case")
		return
	}
	response.Success(c, http.StatusOK, "Case deleted", nil)
}
