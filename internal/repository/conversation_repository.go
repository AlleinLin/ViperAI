package repository

import (
	"viperai/internal/domain"

	"gorm.io/gorm"
)

type ConversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) Create(conv *domain.Conversation) error {
	return r.db.Create(conv).Error
}

func (r *ConversationRepository) FindByID(id string) (*domain.Conversation, error) {
	var conv domain.Conversation
	err := r.db.Where("id = ?", id).First(&conv).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *ConversationRepository) FindByUserID(userID int64) ([]domain.Conversation, error) {
	var conversations []domain.Conversation
	err := r.db.Where("user_id = ?", userID).Find(&conversations).Error
	return conversations, err
}

func (r *ConversationRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.Conversation{}).Error
}
