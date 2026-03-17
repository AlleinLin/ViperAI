package engine

import (
	"context"
	"sync"

	"viperai/internal/domain"
	"viperai/internal/infrastructure/queue"
	"viperai/internal/pkg/utils"
)

type Assistant struct {
	engine      AIEngine
	messages    []*domain.ChatMessage
	mu          sync.RWMutex
	conversationID string
	userID      int64
	saveFunc    func(*domain.ChatMessage) error
}

func NewAssistant(engine AIEngine, conversationID string, userID int64) *Assistant {
	return &Assistant{
		engine:         engine,
		messages:       make([]*domain.ChatMessage, 0),
		conversationID: conversationID,
		userID:         userID,
		saveFunc: func(msg *domain.ChatMessage) error {
			data := queue.EncodeMessage(msg)
			mq := queue.NewMessageQueue("chat_messages")
			return mq.Publish(data)
		},
	}
}

func (a *Assistant) AddMessage(content string, isFromUser bool, persist bool) {
	msg := &domain.ChatMessage{
		ConversationID: a.conversationID,
		UserID:         a.userID,
		Content:        content,
		IsFromUser:     isFromUser,
	}

	a.mu.Lock()
	a.messages = append(a.messages, msg)
	a.mu.Unlock()

	if persist {
		a.saveFunc(msg)
	}
}

func (a *Assistant) SetSaveFunc(saveFunc func(*domain.ChatMessage) error) {
	a.saveFunc = saveFunc
}

func (a *Assistant) GetMessages() []*domain.ChatMessage {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]*domain.ChatMessage, len(a.messages))
	copy(result, a.messages)
	return result
}

func (a *Assistant) Generate(ctx context.Context, question string) (*domain.ChatMessage, error) {
	a.AddMessage(question, true, true)

	a.mu.RLock()
	schemaMsgs := utils.ConvertToSchemaMessages(a.messages)
	a.mu.RUnlock()

	resp, err := a.engine.Generate(ctx, schemaMsgs)
	if err != nil {
		return nil, err
	}

	modelMsg := utils.ConvertToDomainMessage(a.conversationID, a.userID, resp)
	a.AddMessage(modelMsg.Content, false, true)

	return modelMsg, nil
}

func (a *Assistant) Stream(ctx context.Context, question string, handler StreamHandler) (*domain.ChatMessage, error) {
	a.AddMessage(question, true, true)

	a.mu.RLock()
	schemaMsgs := utils.ConvertToSchemaMessages(a.messages)
	a.mu.RUnlock()

	content, err := a.engine.Stream(ctx, schemaMsgs, handler)
	if err != nil {
		return nil, err
	}

	modelMsg := &domain.ChatMessage{
		ConversationID: a.conversationID,
		UserID:         a.userID,
		Content:        content,
		IsFromUser:     false,
	}

	a.AddMessage(modelMsg.Content, false, true)

	return modelMsg, nil
}

func (a *Assistant) EngineType() string {
	return a.engine.Type()
}

type AssistantManager struct {
	assistants map[int64]map[string]*Assistant
	mu         sync.RWMutex
}

func NewAssistantManager() *AssistantManager {
	return &AssistantManager{
		assistants: make(map[int64]map[string]*Assistant),
	}
}

func (m *AssistantManager) GetOrCreate(userID int64, conversationID string, engineType string, opts map[string]interface{}) (*Assistant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	userAssistants, exists := m.assistants[userID]
	if !exists {
		userAssistants = make(map[string]*Assistant)
		m.assistants[userID] = userAssistants
	}

	assistant, exists := userAssistants[conversationID]
	if exists {
		return assistant, nil
	}

	factory := GetFactory()
	engine, err := factory.Create(engineType, context.Background(), opts)
	if err != nil {
		return nil, err
	}

	assistant = NewAssistant(engine, conversationID, userID)
	userAssistants[conversationID] = assistant

	return assistant, nil
}

func (m *AssistantManager) Get(userID int64, conversationID string) (*Assistant, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userAssistants, exists := m.assistants[userID]
	if !exists {
		return nil, false
	}

	assistant, exists := userAssistants[conversationID]
	return assistant, exists
}

func (m *AssistantManager) Remove(userID int64, conversationID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	userAssistants, exists := m.assistants[userID]
	if !exists {
		return
	}

	delete(userAssistants, conversationID)

	if len(userAssistants) == 0 {
		delete(m.assistants, userID)
	}
}

func (m *AssistantManager) GetUserConversations(userID int64) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userAssistants, exists := m.assistants[userID]
	if !exists {
		return []string{}
	}

	ids := make([]string, 0, len(userAssistants))
	for id := range userAssistants {
		ids = append(ids, id)
	}

	return ids
}

var (
	globalManager     *AssistantManager
	globalManagerOnce sync.Once
)

func GetManager() *AssistantManager {
	globalManagerOnce.Do(func() {
		globalManager = NewAssistantManager()
	})
	return globalManager
}
