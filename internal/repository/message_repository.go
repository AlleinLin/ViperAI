package repository

import (
	"viperai/internal/domain"

	"gorm.io/gorm"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(msg *domain.ChatMessage) error {
	return r.db.Create(msg).Error
}

func (r *MessageRepository) FindByConversationID(conversationID string) ([]domain.ChatMessage, error) {
	var messages []domain.ChatMessage
	err := r.db.Where("conversation_id = ?", conversationID).Order("created_at asc").Find(&messages).Error
	return messages, err
}

func (r *MessageRepository) FindByConversationIDs(conversationIDs []string) ([]domain.ChatMessage, error) {
	var messages []domain.ChatMessage
	if len(conversationIDs) == 0 {
		return messages, nil
	}
	err := r.db.Where("conversation_id IN ?", conversationIDs).Order("created_at asc").Find(&messages).Error
	return messages, err
}

func (r *MessageRepository) FindAll() ([]domain.ChatMessage, error) {
	var messages []domain.ChatMessage
	err := r.db.Order("created_at asc").Find(&messages).Error
	return messages, err
}
