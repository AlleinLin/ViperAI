package handler

import (
	"viperai/internal/service"
	"viperai/internal/transport/http/middleware"
	"viperai/internal/transport/http/response"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	fileService *service.FileService
}

func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{fileService: fileService}
}

func (h *FileHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "File is required")
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	filePath, err := h.fileService.UploadRAGFile(userID, file)
	if err != nil {
		if svcErr, ok := err.(*service.ServiceError); ok {
			response.Error(c, svcErr.Code)
			return
		}
		response.ServerError(c, "Failed to upload file")
		return
	}

	response.SuccessWithData(c, gin.H{"filePath": filePath})
}
