package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"legaltech-backend/internal/ai"
	"legaltech-backend/internal/config"
	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"

	"github.com/google/uuid"
)

// requestTimeout bounds every call into the AI provider so a slow/unreachable
// provider can't hang a request indefinitely.
const requestTimeout = 20 * time.Second

type AIService struct {
	cfg  *config.Config
	repo *repositories.AIRepository
}

func NewAIService(cfg *config.Config) *AIService {
	return &AIService{
		cfg:  cfg,
		repo: repositories.NewAIRepository(),
	}
}

// complete sends a single flattened prompt to the AI provider (Groq) and
// returns its reply. Any provider-side failure is already normalized to a
// generic, user-safe message by the ai package.
func (s *AIService) complete(prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	reply, err := ai.GenerateResponse(ctx, prompt)
	if err != nil {
		return "", err
	}
	return reply, nil
}

// Chat sends a message within a conversation, persisting both sides of the exchange.
func (s *AIService) Chat(userID uuid.UUID, req models.AIChatRequest) (*models.AIChatResponse, error) {
	var conversation *models.AIConversation
	var err error

	if req.ConversationID != nil {
		conversation, err = s.repo.FindConversationByID(*req.ConversationID, userID)
		if err != nil {
			return nil, errors.New("conversation not found")
		}
	} else {
		title := req.Message
		if runes := []rune(title); len(runes) > 60 {
			title = string(runes[:60])
		}
		conversation = &models.AIConversation{UserID: userID, Title: title}
		if err := s.repo.CreateConversation(conversation); err != nil {
			return nil, err
		}
	}

	history, err := s.repo.ListMessages(conversation.ID)
	if err != nil {
		return nil, err
	}

	// Cap how much history is sent per turn to keep token usage low; the
	// full history is still persisted and returned by ListMessages/history
	// endpoints — only what's sent to the model is bounded.
	const maxHistoryMessages = 6
	if len(history) > maxHistoryMessages {
		history = history[len(history)-maxHistoryMessages:]
	}

	var b strings.Builder
	for _, m := range history {
		if m.Role == models.AIMessageRoleUser {
			fmt.Fprintf(&b, "User: %s\n", m.Content)
		} else {
			fmt.Fprintf(&b, "Assistant: %s\n", m.Content)
		}
	}
	fmt.Fprintf(&b, "User: %s", req.Message)

	reply, err := s.complete(b.String())
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateMessage(&models.AIMessage{
		ConversationID: conversation.ID,
		Role:           models.AIMessageRoleUser,
		Content:        req.Message,
	}); err != nil {
		return nil, err
	}

	if err := s.repo.CreateMessage(&models.AIMessage{
		ConversationID: conversation.ID,
		Role:           models.AIMessageRoleAssistant,
		Content:        reply,
	}); err != nil {
		return nil, err
	}

	_ = s.repo.TouchConversation(conversation)

	return &models.AIChatResponse{ConversationID: conversation.ID, Reply: reply}, nil
}

func (s *AIService) ListConversations(userID uuid.UUID) ([]models.AIConversation, error) {
	return s.repo.ListConversations(userID)
}

func (s *AIService) ListMessages(conversationID, userID uuid.UUID) ([]models.AIMessage, error) {
	if _, err := s.repo.FindConversationByID(conversationID, userID); err != nil {
		return nil, errors.New("conversation not found")
	}
	return s.repo.ListMessages(conversationID)
}

func (s *AIService) DeleteConversation(id, userID uuid.UUID) error {
	return s.repo.DeleteConversation(id, userID)
}

func (s *AIService) Summarize(text string) (string, error) {
	prompt := "Summarize this legal document concisely. Highlight parties, obligations, dates, and risks:\n\n" + text
	return s.complete(prompt)
}

func (s *AIService) GenerateDraft(draftType models.DraftType, title, instructions string) (string, error) {
	prompt := fmt.Sprintf(
		"Draft a complete, properly formatted %s titled \"%s\" under Indian law. "+
			"Use professional legal language and correct structure, with placeholders in [BRACKETS] for missing information. "+
			"Output only the document text. Instructions: %s",
		draftType, title, instructions,
	)
	return s.complete(prompt)
}

func (s *AIService) LegalResearch(query, category string) (string, error) {
	prompt := fmt.Sprintf(
		"Legal research request (category: %s) under Indian law. Provide relevant statutes, sections, and leading judgements where applicable. "+
			"Query: %s", category, query,
	)
	return s.complete(prompt)
}

func (s *AIService) AnalyzeContract(text string) (string, error) {
	prompt := "Analyze this contract under Indian law. Return three sections in order: 'Key Clauses', 'Risks', and 'Suggestions', " +
		"referencing the actual clauses found:\n\n" + text
	return s.complete(prompt)
}

func (s *AIService) RecommendForCase(caseSummary string) (string, error) {
	prompt := "Based on these case details, give practical, actionable recommendations: next steps, ways to strengthen the case, " +
		"risks to watch for, and procedural considerations under Indian law:\n\n" + caseSummary
	return s.complete(prompt)
}

func (s *AIService) GeneratePetition(req models.PetitionRequest) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "Draft a complete %s petition under Indian law. Output only the petition text, properly formatted.\n", req.PetitionType)
	fmt.Fprintf(&b, "Client (Petitioner): %s\n", req.ClientName)
	if req.Opponent != "" {
		fmt.Fprintf(&b, "Opponent (Respondent): %s\n", req.Opponent)
	}
	if req.Court != "" {
		fmt.Fprintf(&b, "Court: %s\n", req.Court)
	}
	fmt.Fprintf(&b, "Case Facts: %s\n", req.CaseFacts)
	if req.ReliefSought != "" {
		fmt.Fprintf(&b, "Relief Sought: %s\n", req.ReliefSought)
	}
	if req.AdditionalNotes != "" {
		fmt.Fprintf(&b, "Additional Notes: %s\n", req.AdditionalNotes)
	}
	return s.complete(b.String())
}

func (s *AIService) GenerateAgreement(req models.AgreementRequest) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "Draft a complete %s agreement under Indian law. Output only the agreement text, properly formatted with numbered clauses.\n", req.AgreementType)
	fmt.Fprintf(&b, "Party A: %s\n", req.PartyA)
	fmt.Fprintf(&b, "Party B: %s\n", req.PartyB)
	if req.Terms != "" {
		fmt.Fprintf(&b, "Terms: %s\n", req.Terms)
	}
	if req.Duration != "" {
		fmt.Fprintf(&b, "Duration: %s\n", req.Duration)
	}
	if req.Payment != "" {
		fmt.Fprintf(&b, "Payment: %s\n", req.Payment)
	}
	if req.SpecialClauses != "" {
		fmt.Fprintf(&b, "Special Clauses: %s\n", req.SpecialClauses)
	}
	return s.complete(b.String())
}

func (s *AIService) GenerateLegalNotice(req models.LegalNoticeRequest) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "Draft a complete %s legal notice under Indian law. Output only the notice text, properly formatted.\n", req.NoticeType)
	fmt.Fprintf(&b, "Sender: %s\n", req.Sender)
	fmt.Fprintf(&b, "Receiver: %s\n", req.Receiver)
	if req.Reason != "" {
		fmt.Fprintf(&b, "Reason: %s\n", req.Reason)
	}
	if req.Facts != "" {
		fmt.Fprintf(&b, "Facts: %s\n", req.Facts)
	}
	if req.LegalDemand != "" {
		fmt.Fprintf(&b, "Legal Demand: %s\n", req.LegalDemand)
	}
	if req.TimeLimit != "" {
		fmt.Fprintf(&b, "Time Limit to Comply: %s\n", req.TimeLimit)
	}
	return s.complete(b.String())
}
