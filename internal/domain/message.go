package domain

import (
	"time"
)

type ChatMessage struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ConversationID string    `gorm:"index;not null;type:varchar(36)" json:"conversation_id"`
	UserID         int64     `gorm:"type:bigint" json:"user_id"`
	Content        string    `gorm:"type:text" json:"content"`
	IsFromUser     bool      `gorm:"not null" json:"is_from_user"`
	CreatedAt      time.Time `json:"created_at"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}

type MessageHistory struct {
	IsFromUser bool   `json:"is_from_user"`
	Content    string `json:"content"`
}
