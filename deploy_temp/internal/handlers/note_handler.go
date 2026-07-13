package handlers

import (
	"net/http"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NoteHandler struct {
	service *services.NoteService
}

func NewNoteHandler() *NoteHandler {
	return &NoteHandler{service: services.NewNoteService()}
}

func (h *NoteHandler) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var req models.NoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	note, err := h.service.Create(userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create note")
		return
	}
	response.Success(c, http.StatusCreated, "Note created", note)
}

func (h *NoteHandler) List(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	notes, err := h.service.List(userID, c.Query("category"), c.Query("search"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch notes")
		return
	}
	response.Success(c, http.StatusOK, "Notes fetched", notes)
}

func (h *NoteHandler) Get(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid note id")
		return
	}
	note, err := h.service.Get(id, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Note not found")
		return
	}
	response.Success(c, http.StatusOK, "Note fetched", note)
}

func (h *NoteHandler) Update(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid note id")
		return
	}
	var req models.NoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	note, err := h.service.Update(id, userID, req)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Note not found")
		return
	}
	response.Success(c, http.StatusOK, "Note updated", note)
}

func (h *NoteHandler) Delete(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid note id")
		return
	}
	if err := h.service.Delete(id, userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete note")
		return
	}
	response.Success(c, http.StatusOK, "Note deleted", nil)
}
