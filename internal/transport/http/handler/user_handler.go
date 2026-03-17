package handler

import (
	"viperai/internal/domain"
	"viperai/internal/service"
	"viperai/internal/transport/http/middleware"
	"viperai/internal/transport/http/response"

	"github.com/gin-gonic/gin"
)

type UserService interface {
	Login(account, password string) (string, error)
	Register(email, password, captcha string) (string, error)
	SendCaptcha(email string) error
	GetByID(userID int64) (*domain.User, error)
}

type UserHandler struct {
	userService UserService
}

func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type LoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid parameters")
		return
	}

	token, err := h.userService.Login(req.Account, req.Password)
	if err != nil {
		if svcErr, ok := err.(*service.ServiceError); ok {
			response.Error(c, svcErr.Code)
			return
		}
		response.ServerError(c, "Login failed")
		return
	}

	response.SuccessWithData(c, gin.H{"token": token})
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Captcha  string `json:"captcha" binding:"required"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid parameters")
		return
	}

	token, err := h.userService.Register(req.Email, req.Password, req.Captcha)
	if err != nil {
		if svcErr, ok := err.(*service.ServiceError); ok {
			response.Error(c, svcErr.Code)
			return
		}
		response.ServerError(c, "Registration failed")
		return
	}

	response.SuccessWithData(c, gin.H{"token": token})
}

type CaptchaRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *UserHandler) SendCaptcha(c *gin.Context) {
	var req CaptchaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid parameters")
		return
	}

	if err := h.userService.SendCaptcha(req.Email); err != nil {
		response.ServerError(c, "Failed to send captcha")
		return
	}

	response.Success(c)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "User not found")
		return
	}

	user, err := h.userService.GetByID(userID)
	if err != nil {
		response.Error(c, response.CodeUserNotFound)
		return
	}

	response.SuccessWithData(c, gin.H{
		"id":      user.ID,
		"name":    user.Name,
		"email":   user.Email,
		"account": user.Account,
	})
}
