package services

import (
	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"

	"github.com/google/uuid"
)

type TaskService struct {
	repo             *repositories.TaskRepository
	notificationRepo *repositories.NotificationRepository
}

func NewTaskService() *TaskService {
	return &TaskService{
		repo:             repositories.NewTaskRepository(),
		notificationRepo: repositories.NewNotificationRepository(),
	}
}

func (s *TaskService) Create(userID uuid.UUID, req models.TaskRequest) (*models.Task, error) {
	priority := req.Priority
	if priority == "" {
		priority = models.TaskPriorityMedium
	}
	status := req.Status
	if status == "" {
		status = models.TaskStatusPending
	}
	task := models.Task{
		UserID:      userID,
		CaseID:      req.CaseID,
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
		Priority:    priority,
		Status:      status,
		ReminderAt:  req.ReminderAt,
	}
	if err := s.repo.Create(&task); err != nil {
		return nil, err
	}

	_ = s.notificationRepo.Create(&models.Notification{
		UserID:    userID,
		Title:     "Task created",
		Body:      task.Title,
		Type:      models.NotificationTypeTask,
		RelatedID: &task.ID,
	})

	return &task, nil
}

func (s *TaskService) List(userID uuid.UUID, status, priority string) ([]models.Task, error) {
	return s.repo.ListByUser(userID, status, priority)
}

func (s *TaskService) ListDueToday(userID uuid.UUID) ([]models.Task, error) {
	return s.repo.ListDueToday(userID)
}

func (s *TaskService) ListUpcoming(userID uuid.UUID, limit int) ([]models.Task, error) {
	return s.repo.ListUpcoming(userID, limit)
}

func (s *TaskService) Get(id, userID uuid.UUID) (*models.Task, error) {
	return s.repo.FindByID(id, userID)
}

func (s *TaskService) Update(id, userID uuid.UUID, req models.TaskRequest) (*models.Task, error) {
	task, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, err
	}
	task.CaseID = req.CaseID
	task.Title = req.Title
	task.Description = req.Description
	task.DueDate = req.DueDate
	if req.Priority != "" {
		task.Priority = req.Priority
	}
	if req.Status != "" {
		task.Status = req.Status
	}
	task.ReminderAt = req.ReminderAt
	if err := s.repo.Update(task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) Delete(id, userID uuid.UUID) error {
	return s.repo.Delete(id, userID)
}
