package handler

import (
	"viperai/internal/service"
	"viperai/internal/transport/http/response"

	"github.com/gin-gonic/gin"
)

type TTSHandler struct {
	ttsService *service.TTSService
}

func NewTTSHandler(ttsService *service.TTSService) *TTSHandler {
	return &TTSHandler{ttsService: ttsService}
}

type CreateTTSRequest struct {
	Text string `json:"text" binding:"required"`
}

func (h *TTSHandler) CreateTask(c *gin.Context) {
	var req CreateTTSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Text is required")
		return
	}

	taskID, err := h.ttsService.CreateTask(c.Request.Context(), req.Text)
	if err != nil {
		response.Error(c, response.CodeTTSError)
		return
	}

	response.SuccessWithData(c, gin.H{"taskId": taskID})
}

func (h *TTSHandler) QueryTask(c *gin.Context) {
	taskID := c.Query("taskId")
	if taskID == "" {
		response.BadRequest(c, "Task ID is required")
		return
	}

	result, err := h.ttsService.QueryTask(c.Request.Context(), taskID)
	if err != nil {
		response.Error(c, response.CodeTTSError)
		return
	}

	if len(result.TasksInfo) == 0 {
		response.Error(c, response.CodeTTSError)
		return
	}

	task := result.TasksInfo[0]
	respData := gin.H{
		"taskId":     task.TaskID,
		"taskStatus": task.TaskStatus,
	}

	if task.TaskResult != nil {
		respData["taskResult"] = task.TaskResult.SpeechURL
	}

	response.SuccessWithData(c, respData)
}
