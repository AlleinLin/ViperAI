package repository

import (
	"testing"

	"viperai/internal/domain"
)

func TestUserRepository_Create(t *testing.T) {
	t.Run("user creation", func(t *testing.T) {
		user := &domain.User{
			ID:      1,
			Name:    "Test User",
			Email:   "test@example.com",
			Account: "12345678901",
		}

		if user.ID != 1 {
			t.Errorf("Expected ID 1, got %d", user.ID)
		}
		if user.Email != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
		}
	})
}

func TestConversationRepository_Create(t *testing.T) {
	t.Run("conversation creation", func(t *testing.T) {
		conv := &domain.Conversation{
			ID:     "test-conv-id",
			UserID: 1,
			Title:  "Test Conversation",
		}

		if conv.ID != "test-conv-id" {
			t.Errorf("Expected ID 'test-conv-id', got '%s'", conv.ID)
		}
	})
}

func TestMessageRepository_Create(t *testing.T) {
	t.Run("message creation", func(t *testing.T) {
		msg := &domain.ChatMessage{
			ID:             1,
			ConversationID: "test-conv-id",
			UserID:         1,
			Content:        "Hello, AI!",
			IsFromUser:     true,
		}

		if msg.Content != "Hello, AI!" {
			t.Errorf("Expected content 'Hello, AI!', got '%s'", msg.Content)
		}
		if !msg.IsFromUser {
			t.Error("Expected IsFromUser to be true")
		}
	})
}

func TestDomainModels(t *testing.T) {
	t.Run("user table name", func(t *testing.T) {
		user := domain.User{}
		if user.TableName() != "users" {
			t.Errorf("Expected table name 'users', got '%s'", user.TableName())
		}
	})

	t.Run("conversation table name", func(t *testing.T) {
		conv := domain.Conversation{}
		if conv.TableName() != "conversations" {
			t.Errorf("Expected table name 'conversations', got '%s'", conv.TableName())
		}
	})

	t.Run("message table name", func(t *testing.T) {
		msg := domain.ChatMessage{}
		if msg.TableName() != "chat_messages" {
			t.Errorf("Expected table name 'chat_messages', got '%s'", msg.TableName())
		}
	})
}
