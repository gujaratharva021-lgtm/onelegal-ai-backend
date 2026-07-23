package services

import (
	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"

	"github.com/google/uuid"
)

type CaseService struct {
	repo                *repositories.CaseRepository
	clientRepo          *repositories.ClientRepository
	notificationService *NotificationService
}

func NewCaseService() *CaseService {
	return &CaseService{
		repo:                repositories.NewCaseRepository(),
		clientRepo:          repositories.NewClientRepository(),
		notificationService: NewNotificationService(),
	}
}

// syncClientAccountStatus is called after every case create/update that
// touches a client's cases. A client's login account is temporary: it stays
// Active only while they have at least one non-completed case, and is
// automatically flipped to Inactive (blocking login immediately) the moment
// their lawyer closes the last one — then reactivated the instant a case is
// reopened or a new one is created for them. The client's Login ID never
// changes across this cycle.
func (s *CaseService) syncClientAccountStatus(userID, clientID uuid.UUID) {
	client, err := s.clientRepo.FindByID(clientID, userID)
	if err != nil || client.AccountUserID == nil {
		return
	}
	openCount, err := s.repo.CountOpenByClient(clientID, userID)
	if err != nil {
		return
	}
	newStatus := models.ClientAccountInactive
	if openCount > 0 {
		newStatus = models.ClientAccountActive
	}
	if newStatus == client.AccountStatus {
		return
	}
	client.AccountStatus = newStatus
	if err := s.clientRepo.Update(client); err != nil {
		return
	}
	if newStatus == models.ClientAccountInactive {
		_, _ = s.notificationService.Create(*client.AccountUserID, models.NotificationRequest{
			Title: "Your account is now inactive",
			Body:  "Your case has been closed. Please contact your advocate if you need further access.",
			Type:  models.NotificationTypeGeneral,
		})
	}
}

func (s *CaseService) Create(userID uuid.UUID, req models.CaseRequest) (*models.Case, error) {
	status := req.Status
	if status == "" {
		status = models.CaseStatusUpcoming
	}
	priority := req.Priority
	if priority == "" {
		priority = models.CasePriorityMedium
	}
	c := models.Case{
		UserID:          userID,
		ClientID:        req.ClientID,
		Title:           req.Title,
		CaseNumber:      req.CaseNumber,
		CourtName:       req.CourtName,
		CourtNumber:     req.CourtNumber,
		Judge:           req.Judge,
		Opponent:        req.Opponent,
		CaseType:        req.CaseType,
CNRNumber:       req.CNRNumber,
LodgingNumber:   req.LodgingNumber,
FilingDate:      req.FilingDate,
RegNumber:       req.RegNumber,
RegDate:         req.RegDate,
Petitioner:      req.Petitioner,
Respondent:      req.Respondent,
PetnAdvocate:    req.PetnAdvocate,
RespAdvocate:    req.RespAdvocate,
District:        req.District,
Bench:           req.Bench,
Category:        req.Category,
Stage:           req.Stage,
Coram:           req.Coram,
LastHearingDate: req.LastHearingDate,
LastCoram:       req.LastCoram,
Act:             req.Act,
		Priority:        priority,
		Status:          status,
		Description:     req.Description,
		NextHearingDate: req.NextHearingDate,
	}
	if err := s.repo.Create(&c); err != nil {
		return nil, err
	}
	if c.ClientID != nil {
		s.syncClientAccountStatus(userID, *c.ClientID)
	}
	return &c, nil
}

type CaseListResult struct {
	Cases []models.Case `json:"cases"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
}

func (s *CaseService) List(userID uuid.UUID, f repositories.CaseListFilter) (*CaseListResult, error) {
	cases, total, err := s.repo.ListByUser(userID, f)
	if err != nil {
		return nil, err
	}
	page := f.Page
	if page < 1 {
		page = 1
	}
	return &CaseListResult{Cases: cases, Total: total, Page: page}, nil
}

func (s *CaseService) ListRecent(userID uuid.UUID, limit int) ([]models.Case, error) {
	return s.repo.ListRecent(userID, limit)
}

func (s *CaseService) Get(id, userID uuid.UUID) (*models.Case, error) {
	return s.repo.FindByID(id, userID)
}

func (s *CaseService) Update(id, userID uuid.UUID, req models.CaseRequest) (*models.Case, error) {
	c, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, err
	}
	previousClientID := c.ClientID
	c.ClientID = req.ClientID
	c.Title = req.Title
	c.CaseNumber = req.CaseNumber
	c.CourtName = req.CourtName
	c.CourtNumber = req.CourtNumber
	c.Judge = req.Judge
	c.Opponent = req.Opponent
	c.CaseType = req.CaseType
c.CNRNumber = req.CNRNumber
c.LodgingNumber = req.LodgingNumber
c.FilingDate = req.FilingDate
c.RegNumber = req.RegNumber
c.RegDate = req.RegDate
c.Petitioner = req.Petitioner
c.Respondent = req.Respondent
c.PetnAdvocate = req.PetnAdvocate
c.RespAdvocate = req.RespAdvocate
c.District = req.District
c.Bench = req.Bench
c.Category = req.Category
c.Stage = req.Stage
c.Coram = req.Coram
c.LastHearingDate = req.LastHearingDate
c.LastCoram = req.LastCoram
c.Act = req.Act
	if req.Priority != "" {
		c.Priority = req.Priority
	}
	if req.Status != "" {
		c.Status = req.Status
	}
	c.Description = req.Description
	c.NextHearingDate = req.NextHearingDate
	if err := s.repo.Update(c); err != nil {
		return nil, err
	}
	// Re-sync both the previous and (if reassigned) new client, since either
	// one's open-case count may have just changed.
	if previousClientID != nil {
		s.syncClientAccountStatus(userID, *previousClientID)
	}
	if c.ClientID != nil && (previousClientID == nil || *c.ClientID != *previousClientID) {
		s.syncClientAccountStatus(userID, *c.ClientID)
	}
	return c, nil
}

func (s *CaseService) Delete(id, userID uuid.UUID) error {
	return s.repo.Delete(id, userID)
}
