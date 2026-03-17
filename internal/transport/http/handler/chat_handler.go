package handler

import (
	"fmt"
	"net/http"

	"viperai/internal/domain"
	"viperai/internal/service"
	"viperai/internal/transport/http/middleware"
	"viperai/internal/transport/http/response"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

type SendMessageRequest struct {
	Question   string `json:"question" binding:"required"`
	EngineType string `json:"engineType" binding:"required"`
}

type SendMessageResponse struct {
	response.DataResponse
	ConversationID string `json:"conversationId,omitempty"`
	Content        string `json:"content,omitempty"`
}

func (h *ChatHandler) GetConversations(c *gin.Context) {
	userID := middleware.GetUserID(c)

	conversations, err := h.chatService.GetUserConversations(userID)
	if err != nil {
		response.ServerError(c, "Failed to get conversations")
		return
	}

	response.SuccessWithData(c, gin.H{"conversations": conversations})
}

func (h *ChatHandler) CreateAndSend(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid parameters")
		return
	}

	convID, content, err := h.chatService.CreateConversationAndSend(userID, req.Question, req.EngineType)
	if err != nil {
		if svcErr, ok := err.(*service.ServiceError); ok {
			response.Error(c, svcErr.Code)
			return
		}
		response.Error(c, response.CodeAIModelError)
		return
	}

	response.SuccessWithData(c, gin.H{
		"conversationId": convID,
		"content":        content,
	})
}

type SendToConversationRequest struct {
	Question       string `json:"question" binding:"required"`
	EngineType     string `json:"engineType" binding:"required"`
	ConversationID string `json:"conversationId" binding:"required"`
}

func (h *ChatHandler) Send(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req SendToConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid parameters")
		return
	}

	content, err := h.chatService.Send(userID, req.ConversationID, req.Question, req.EngineType)
	if err != nil {
		if svcErr, ok := err.(*service.ServiceError); ok {
			response.Error(c, svcErr.Code)
			return
		}
		response.Error(c, response.CodeAIModelError)
		return
	}

	response.SuccessWithData(c, gin.H{"content": content})
}

func (h *ChatHandler) CreateAndStream(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "Invalid parameters"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Accel-Buffering", "no")

	convID, err := h.chatService.CreateConversation(userID, req.Question)
	if err != nil {
		c.SSEvent("error", gin.H{"message": "Failed to create conversation"})
		return
	}

	c.Writer.WriteString(fmt.Sprintf("data: {\"conversationId\": \"%s\"}\n\n", convID))
	c.Writer.Flush()

	if err := h.chatService.StreamToConversation(userID, convID, req.Question, req.EngineType, http.ResponseWriter(c.Writer)); err != nil {
		c.SSEvent("error", gin.H{"message": "Failed to send message"})
		return
	}
}

func (h *ChatHandler) Stream(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req SendToConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "Invalid parameters"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Accel-Buffering", "no")

	if err := h.chatService.Stream(userID, req.ConversationID, req.Question, req.EngineType, http.ResponseWriter(c.Writer)); err != nil {
		c.SSEvent("error", gin.H{"message": "Failed to send message"})
		return
	}
}

type HistoryRequest struct {
	ConversationID string `json:"conversationId" binding:"required"`
}

type HistoryResponse struct {
	response.DataResponse
	History []domain.MessageHistory `json:"history"`
}

func (h *ChatHandler) GetHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req HistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid parameters")
		return
	}

	history, err := h.chatService.GetHistory(userID, req.ConversationID)
	if err != nil {
		if svcErr, ok := err.(*service.ServiceError); ok {
			response.Error(c, svcErr.Code)
			return
		}
		response.ServerError(c, "Failed to get history")
		return
	}

	response.SuccessWithData(c, gin.H{"history": history})
}
