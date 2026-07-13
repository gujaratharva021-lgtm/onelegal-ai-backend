package handlers

import (
	"net/http"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// AIDraftingHandler serves the three structured document generators
// (Petition, Agreement, Legal Notice). Each generates text via the existing
// Groq-backed AIService, then saves it as a Document so it appears in
// Saved Documents like any other generated file.
type AIDraftingHandler struct {
	aiService  *services.AIService
	docService *services.DocumentService
}

func NewAIDraftingHandler(aiService *services.AIService, docService *services.DocumentService) *AIDraftingHandler {
	return &AIDraftingHandler{aiService: aiService, docService: docService}
}

func (h *AIDraftingHandler) Petition(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var req models.PetitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	content, err := h.aiService.GeneratePetition(req)
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}

	doc, err := h.docService.Create(userID, models.DocumentRequest{
		Title:        req.PetitionType + " Petition - " + req.ClientName,
		DocumentType: "petition",
		Content:      content,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to save generated petition")
		return
	}

	response.Success(c, http.StatusCreated, "Petition generated", doc)
}

func (h *AIDraftingHandler) Agreement(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var req models.AgreementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	content, err := h.aiService.GenerateAgreement(req)
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}

	doc, err := h.docService.Create(userID, models.DocumentRequest{
		Title:        req.AgreementType + " Agreement - " + req.PartyA + " & " + req.PartyB,
		DocumentType: "agreement",
		Content:      content,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to save generated agreement")
		return
	}

	response.Success(c, http.StatusCreated, "Agreement generated", doc)
}

func (h *AIDraftingHandler) LegalNotice(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var req models.LegalNoticeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	content, err := h.aiService.GenerateLegalNotice(req)
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}

	doc, err := h.docService.Create(userID, models.DocumentRequest{
		Title:        req.NoticeType + " Notice - " + req.Sender + " to " + req.Receiver,
		DocumentType: "legal_notice",
		Content:      content,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to save generated notice")
		return
	}

	response.Success(c, http.StatusCreated, "Legal notice generated", doc)
}
