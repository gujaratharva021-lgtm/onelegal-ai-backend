package handlers

import (
	"net/http"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TaskHandler struct {
	service *services.TaskService
}

func NewTaskHandler() *TaskHandler {
	return &TaskHandler{service: services.NewTaskService()}
}

func (h *TaskHandler) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var req models.TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	task, err := h.service.Create(userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create task")
		return
	}
	response.Success(c, http.StatusCreated, "Task created", task)
}

func (h *TaskHandler) List(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	tasks, err := h.service.List(userID, c.Query("status"), c.Query("priority"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch tasks")
		return
	}
	response.Success(c, http.StatusOK, "Tasks fetched", tasks)
}

func (h *TaskHandler) Get(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid task id")
		return
	}
	task, err := h.service.Get(id, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Task not found")
		return
	}
	response.Success(c, http.StatusOK, "Task fetched", task)
}

func (h *TaskHandler) Update(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid task id")
		return
	}
	var req models.TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	task, err := h.service.Update(id, userID, req)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Task not found")
		return
	}
	response.Success(c, http.StatusOK, "Task updated", task)
}

func (h *TaskHandler) Delete(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid task id")
		return
	}
	if err := h.service.Delete(id, userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete task")
		return
	}
	response.Success(c, http.StatusOK, "Task deleted", nil)
}
