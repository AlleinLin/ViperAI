package utils

import (
	"viperai/internal/domain"

	"github.com/cloudwego/eino/schema"
)

func ConvertToSchemaMessages(messages []*domain.ChatMessage) []*schema.Message {
	result := make([]*schema.Message, 0, len(messages))
	for _, m := range messages {
		role := schema.Assistant
		if m.IsFromUser {
			role = schema.User
		}
		result = append(result, &schema.Message{
			Role:    role,
			Content: m.Content,
		})
	}
	return result
}

func ConvertToDomainMessage(conversationID string, userID int64, msg *schema.Message) *domain.ChatMessage {
	return &domain.ChatMessage{
		ConversationID: conversationID,
		UserID:         userID,
		Content:        msg.Content,
		IsFromUser:     false,
	}
}
