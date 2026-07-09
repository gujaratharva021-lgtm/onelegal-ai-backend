package repositories

import (
	"time"

	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"

	"github.com/google/uuid"
)

type TaskRepository struct{}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{}
}

func (r *TaskRepository) Create(t *models.Task) error {
	return database.DB.Create(t).Error
}

func (r *TaskRepository) FindByID(id, userID uuid.UUID) (*models.Task, error) {
	var t models.Task
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TaskRepository) ListByUser(userID uuid.UUID, status, priority string) ([]models.Task, error) {
	var tasks []models.Task
	q := database.DB.Where("user_id = ?", userID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if priority != "" {
		q = q.Where("priority = ?", priority)
	}
	if err := q.Order("due_date ASC NULLS LAST").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *TaskRepository) ListDueToday(userID uuid.UUID) ([]models.Task, error) {
	var tasks []models.Task
	// See the identical fix/comment in HearingRepository.ListToday — Truncate
	// aligns to UTC-based boundaries, not local calendar midnight.
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)
	if err := database.DB.Where("user_id = ? AND status = ? AND due_date >= ? AND due_date < ?", userID, models.TaskStatusPending, start, end).
		Order("due_date ASC").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// ListUpcoming is "pending tasks due today or later" — bounded by the start
// of today (calendar day), not the exact current instant. Using time.Now()
// as the lower bound would wrongly exclude a task still pending and due
// earlier today (e.g. due 9am, checked at 3pm) even though it's not done.
func (r *TaskRepository) ListUpcoming(userID uuid.UUID, limit int) ([]models.Task, error) {
	var tasks []models.Task
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if err := database.DB.Where("user_id = ? AND status = ? AND due_date >= ?", userID, models.TaskStatusPending, startOfToday).
		Order("due_date ASC").Limit(limit).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *TaskRepository) Update(t *models.Task) error {
	return database.DB.Save(t).Error
}

func (r *TaskRepository) Delete(id, userID uuid.UUID) error {
	return database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Task{}).Error
}
