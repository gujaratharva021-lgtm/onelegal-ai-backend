package repositories

import (
	"legaltech-backend/internal/database"
	"legaltech-backend/internal/models"

	"github.com/google/uuid"
)

type AIRepository struct{}

func NewAIRepository() *AIRepository {
	return &AIRepository{}
}

func (r *AIRepository) CreateConversation(c *models.AIConversation) error {
	return database.DB.Create(c).Error
}

func (r *AIRepository) FindConversationByID(id, userID uuid.UUID) (*models.AIConversation, error) {
	var c models.AIConversation
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *AIRepository) ListConversations(userID uuid.UUID) ([]models.AIConversation, error) {
	var conversations []models.AIConversation
	if err := database.DB.Where("user_id = ?", userID).Order("updated_at DESC").Find(&conversations).Error; err != nil {
		return nil, err
	}
	return conversations, nil
}

func (r *AIRepository) TouchConversation(c *models.AIConversation) error {
	return database.DB.Save(c).Error
}

func (r *AIRepository) CreateMessage(m *models.AIMessage) error {
	return database.DB.Create(m).Error
}

func (r *AIRepository) ListMessages(conversationID uuid.UUID) ([]models.AIMessage, error) {
	var messages []models.AIMessage
	if err := database.DB.Where("conversation_id = ?", conversationID).Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *AIRepository) DeleteConversation(id, userID uuid.UUID) error {
	// Verify ownership first — deleting messages by conversation_id alone,
	// before confirming this conversation belongs to userID, would let a
	// user wipe another user's conversation's messages just by guessing an
	// id, even though the conversation row itself is correctly scoped below.
	var conv models.AIConversation
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&conv).Error; err != nil {
		return err
	}
	if err := database.DB.Where("conversation_id = ?", id).Delete(&models.AIMessage{}).Error; err != nil {
		return err
	}
	return database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.AIConversation{}).Error
}
