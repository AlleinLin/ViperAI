package handler

import (
	"viperai/internal/service"
	"viperai/internal/transport/http/response"

	"github.com/gin-gonic/gin"
)

type ImageHandler struct {
	imageService *service.ImageService
}

func NewImageHandler(imageService *service.ImageService) *ImageHandler {
	return &ImageHandler{imageService: imageService}
}

func (h *ImageHandler) Recognize(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		response.BadRequest(c, "Image is required")
		return
	}

	className, err := h.imageService.Recognize(file)
	if err != nil {
		response.ServerError(c, "Failed to recognize image")
		return
	}

	response.SuccessWithData(c, gin.H{"className": className})
}
