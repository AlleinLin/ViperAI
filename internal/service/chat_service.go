package service

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"viperai/internal/domain"
	"viperai/internal/engine"
	"viperai/internal/repository"

	"github.com/google/uuid"
)

type ChatService struct {
	convRepo    *repository.ConversationRepository
	msgRepo     *repository.MessageRepository
	assistantMgr *engine.AssistantManager
}

func NewChatService(convRepo *repository.ConversationRepository, msgRepo *repository.MessageRepository) *ChatService {
	return &ChatService{
		convRepo:     convRepo,
		msgRepo:      msgRepo,
		assistantMgr: engine.GetManager(),
	}
}

func (s *ChatService) GetUserConversations(userID int64) ([]domain.ConversationInfo, error) {
	ids := s.assistantMgr.GetUserConversations(userID)

	infos := make([]domain.ConversationInfo, 0, len(ids))
	for _, id := range ids {
		infos = append(infos, domain.ConversationInfo{
			ID:    id,
			Title: id,
		})
	}

	return infos, nil
}

func (s *ChatService) CreateConversationAndSend(userID int64, question, engineType string) (string, string, error) {
	convID := uuid.New().String()

	conv := &domain.Conversation{
		ID:     convID,
		UserID: userID,
		Title:  question,
	}

	if err := s.convRepo.Create(conv); err != nil {
		log.Println("Failed to create conversation:", err)
		return "", "", ErrConversationCreate
	}

	opts := map[string]interface{}{
		"user_id": userID,
	}

	assistant, err := s.assistantMgr.GetOrCreate(userID, convID, engineType, opts)
	if err != nil {
		log.Println("Failed to create assistant:", err)
		return "", "", ErrAIEngine
	}

	msg, err := assistant.Generate(context.Background(), question)
	if err != nil {
		log.Println("Failed to generate response:", err)
		return "", "", ErrAIEngine
	}

	return convID, msg.Content, nil
}

func (s *ChatService) CreateConversation(userID int64, question string) (string, error) {
	convID := uuid.New().String()

	conv := &domain.Conversation{
		ID:     convID,
		UserID: userID,
		Title:  question,
	}

	if err := s.convRepo.Create(conv); err != nil {
		return "", ErrConversationCreate
	}

	return convID, nil
}

func (s *ChatService) StreamToConversation(userID int64, convID, question, engineType string, writer http.ResponseWriter) error {
	flusher, ok := writer.(http.Flusher)
	if !ok {
		return ErrStreamingUnsupported
	}

	opts := map[string]interface{}{
		"user_id": userID,
	}

	assistant, err := s.assistantMgr.GetOrCreate(userID, convID, engineType, opts)
	if err != nil {
		log.Println("Failed to create assistant:", err)
		return ErrAIEngine
	}

	handler := func(chunk string) {
		log.Printf("[SSE] Sending chunk: %s (len=%d)\n", chunk, len(chunk))
		_, err := writer.Write([]byte("data: " + chunk + "\n\n"))
		if err != nil {
			log.Println("[SSE] Write error:", err)
			return
		}
		flusher.Flush()
	}

	_, err = assistant.Stream(context.Background(), question, handler)
	if err != nil {
		log.Println("Stream error:", err)
		return ErrAIEngine
	}

	_, err = writer.Write([]byte("data: [DONE]\n\n"))
	if err != nil {
		return ErrAIEngine
	}
	flusher.Flush()

	return nil
}

func (s *ChatService) CreateAndStream(userID int64, question, engineType string, writer http.ResponseWriter) (string, error) {
	convID, err := s.CreateConversation(userID, question)
	if err != nil {
		return "", err
	}

	if err := s.StreamToConversation(userID, convID, question, engineType, writer); err != nil {
		return convID, err
	}

	return convID, nil
}

func (s *ChatService) Send(userID int64, convID, question, engineType string) (string, error) {
	opts := map[string]interface{}{
		"user_id": userID,
	}

	assistant, err := s.assistantMgr.GetOrCreate(userID, convID, engineType, opts)
	if err != nil {
		log.Println("Failed to get assistant:", err)
		return "", ErrAIEngine
	}

	msg, err := assistant.Generate(context.Background(), question)
	if err != nil {
		log.Println("Failed to generate response:", err)
		return "", ErrAIEngine
	}

	return msg.Content, nil
}

func (s *ChatService) GetHistory(userID int64, convID string) ([]domain.MessageHistory, error) {
	assistant, exists := s.assistantMgr.Get(userID, convID)
	if !exists {
		return nil, ErrConversationNotFound
	}

	messages := assistant.GetMessages()
	history := make([]domain.MessageHistory, 0, len(messages))

	for i, msg := range messages {
		isFromUser := i%2 == 0
		history = append(history, domain.MessageHistory{
			IsFromUser: isFromUser,
			Content:    msg.Content,
		})
	}

	return history, nil
}

func (s *ChatService) Stream(userID int64, convID, question, engineType string, writer http.ResponseWriter) error {
	return s.StreamToConversation(userID, convID, question, engineType, writer)
}

var (
	ErrConversationCreate    = NewServiceError(4001, "Failed to create conversation")
	ErrAIEngine              = NewServiceError(5001, "AI engine error")
	ErrConversationNotFound  = NewServiceError(2009, "Conversation not found")
	ErrStreamingUnsupported  = NewServiceError(4001, "Streaming not supported")
)

func init() {
	_ = fmt.Sprintf("")
}
