package handlers

import (
	"net/http"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ResearchHandler struct {
	service *services.ResearchService
}

func NewResearchHandler(service *services.ResearchService) *ResearchHandler {
	return &ResearchHandler{service: service}
}

func (h *ResearchHandler) Search(c *gin.Context) {
	userID, _ := currentUserID(c)
	var req models.ResearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	entry, err := h.service.Search(userID, req)
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Research completed", entry)
}

func (h *ResearchHandler) List(c *gin.Context) {
	userID, _ := currentUserID(c)
	history, err := h.service.List(userID, c.Query("bookmarked") == "true")
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch research history")
		return
	}
	response.Success(c, http.StatusOK, "Research history fetched", history)
}

func (h *ResearchHandler) SetBookmark(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid research id")
		return
	}
	var req models.ResearchBookmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.service.SetBookmark(id, userID, req.Bookmarked); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update bookmark")
		return
	}
	response.Success(c, http.StatusOK, "Bookmark updated", nil)
}

func (h *ResearchHandler) Delete(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid research id")
		return
	}
	if err := h.service.Delete(id, userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete research entry")
		return
	}
	response.Success(c, http.StatusOK, "Research entry deleted", nil)
}
