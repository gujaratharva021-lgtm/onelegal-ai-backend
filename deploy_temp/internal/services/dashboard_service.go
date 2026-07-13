package services

import (
	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"

	"github.com/google/uuid"
)

type DashboardService struct {
	userRepo         *repositories.UserRepository
	hearingRepo      *repositories.HearingRepository
	taskRepo         *repositories.TaskRepository
	caseRepo         *repositories.CaseRepository
	clientRepo       *repositories.ClientRepository
	researchRepo     *repositories.ResearchRepository
	notificationRepo *repositories.NotificationRepository
}

func NewDashboardService() *DashboardService {
	return &DashboardService{
		userRepo:         repositories.NewUserRepository(),
		hearingRepo:      repositories.NewHearingRepository(),
		taskRepo:         repositories.NewTaskRepository(),
		caseRepo:         repositories.NewCaseRepository(),
		clientRepo:       repositories.NewClientRepository(),
		researchRepo:     repositories.NewResearchRepository(),
		notificationRepo: repositories.NewNotificationRepository(),
	}
}

type DashboardResponse struct {
	User             *models.User             `json:"user"`
	TodaysHearings   []models.Hearing         `json:"todays_hearings"`
	UpcomingTasks    []models.Task            `json:"upcoming_tasks"`
	RecentCases      []models.Case            `json:"recent_cases"`
	RecentResearch   []models.ResearchHistory `json:"recent_research"`
	UnreadNotifCount int64                    `json:"unread_notification_count"`
}

func (s *DashboardService) Get(userID uuid.UUID) (*DashboardResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	hearings, err := s.hearingRepo.ListToday(userID)
	if err != nil {
		return nil, err
	}
	for i := range hearings {
		if c, err := s.caseRepo.FindByID(hearings[i].CaseID, userID); err == nil {
			hearings[i].CaseTitle = c.Title
			hearings[i].CaseStatus = string(c.Status)
			if c.ClientID != nil {
				if client, err := s.clientRepo.FindByID(*c.ClientID, userID); err == nil {
					hearings[i].ClientName = client.Name
				}
			}
		}
	}

	tasks, err := s.taskRepo.ListUpcoming(userID, 5)
	if err != nil {
		return nil, err
	}
	for i := range tasks {
		if tasks[i].CaseID != nil {
			if c, err := s.caseRepo.FindByID(*tasks[i].CaseID, userID); err == nil {
				tasks[i].CaseTitle = c.Title
			}
		}
	}

	cases, err := s.caseRepo.ListRecent(userID, 5)
	if err != nil {
		return nil, err
	}

	research, err := s.researchRepo.ListRecent(userID, 5)
	if err != nil {
		return nil, err
	}

	unreadCount, err := s.notificationRepo.CountUnread(userID)
	if err != nil {
		return nil, err
	}

	return &DashboardResponse{
		User:             user,
		TodaysHearings:   hearings,
		UpcomingTasks:    tasks,
		RecentCases:      cases,
		RecentResearch:   research,
		UnreadNotifCount: unreadCount,
	}, nil
}
