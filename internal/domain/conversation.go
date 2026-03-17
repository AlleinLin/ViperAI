package domain

import (
	"time"

	"gorm.io/gorm"
)

type Conversation struct {
	ID        string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID    int64          `gorm:"index;not null" json:"user_id"`
	Title     string         `gorm:"type:varchar(100)" json:"title"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Conversation) TableName() string {
	return "conversations"
}

type ConversationInfo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
