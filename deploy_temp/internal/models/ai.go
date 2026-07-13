package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AIMessageRole string

const (
	AIMessageRoleUser      AIMessageRole = "user"
	AIMessageRoleAssistant AIMessageRole = "assistant"
)

type AIConversation struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	Title     string         `json:"title" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type AIMessage struct {
	ID             uuid.UUID     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ConversationID uuid.UUID     `json:"conversation_id" gorm:"type:uuid;not null;index"`
	Role           AIMessageRole `json:"role" gorm:"not null"`
	Content        string        `json:"content" gorm:"type:text;not null"`
	CreatedAt      time.Time     `json:"created_at"`
}

type AIChatRequest struct {
	ConversationID *uuid.UUID `json:"conversation_id"`
	Message        string     `json:"message" binding:"required"`
}

type AIChatResponse struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	Reply          string    `json:"reply"`
}

type AISummarizeRequest struct {
	Text string `json:"text" binding:"required"`
}

type AISummarizeResponse struct {
	Summary string `json:"summary"`
}
