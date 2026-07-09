package services

import (
	"time"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"

	"github.com/google/uuid"
)

type CalendarService struct {
	repo *repositories.CalendarRepository
}

func NewCalendarService() *CalendarService {
	return &CalendarService{repo: repositories.NewCalendarRepository()}
}

func (s *CalendarService) Create(userID uuid.UUID, req models.CalendarEventRequest) (*models.CalendarEvent, error) {
	eventType := req.EventType
	if eventType == "" {
		eventType = models.CalendarEventPersonal
	}
	event := models.CalendarEvent{
		UserID:           userID,
		Title:            req.Title,
		Description:      req.Description,
		EventType:        eventType,
		StartTime:        req.StartTime,
		EndTime:          req.EndTime,
		Location:         req.Location,
		RelatedCaseID:    req.RelatedCaseID,
		RelatedMeetingID: req.RelatedMeetingID,
	}
	if err := s.repo.Create(&event); err != nil {
		return nil, err
	}
	return &event, nil
}

func (s *CalendarService) ListRange(userID uuid.UUID, from, to time.Time) ([]models.CalendarEvent, error) {
	return s.repo.ListByRange(userID, from, to)
}

func (s *CalendarService) Get(id, userID uuid.UUID) (*models.CalendarEvent, error) {
	return s.repo.FindByID(id, userID)
}

func (s *CalendarService) Update(id, userID uuid.UUID, req models.CalendarEventRequest) (*models.CalendarEvent, error) {
	event, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, err
	}
	event.Title = req.Title
	event.Description = req.Description
	if req.EventType != "" {
		event.EventType = req.EventType
	}
	event.StartTime = req.StartTime
	event.EndTime = req.EndTime
	event.Location = req.Location
	event.RelatedCaseID = req.RelatedCaseID
	event.RelatedMeetingID = req.RelatedMeetingID
	if err := s.repo.Update(event); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *CalendarService) Delete(id, userID uuid.UUID) error {
	return s.repo.Delete(id, userID)
}
