package service

import (
	"fmt"

	"viperai/internal/domain"
	"viperai/internal/infrastructure/cache"
	"viperai/internal/pkg/auth"
	"viperai/internal/pkg/utils"
	"viperai/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) Login(account, password string) (string, error) {
	user, err := s.userRepo.FindByAccount(account)
	if err != nil {
		return "", ErrUserNotFound
	}

	if user.Password != utils.HashPassword(password) {
		return "", ErrInvalidPassword
	}

	token, err := auth.GenerateToken(user.ID, user.Account)
	if err != nil {
		return "", ErrTokenGeneration
	}

	return token, nil
}

func (s *UserService) Register(email, password, captcha string) (string, error) {
	if s.userRepo.ExistsByEmail(email) {
		return "", ErrUserExists
	}

	valid, err := cache.VerifyCaptcha(email, captcha)
	if err != nil || !valid {
		return "", ErrInvalidCaptcha
	}

	account := utils.GenerateRandomCode(11)

	user := &domain.User{
		Name:     account,
		Email:    email,
		Account:  account,
		Password: utils.HashPassword(password),
	}

	if err := s.userRepo.Create(user); err != nil {
		return "", ErrRegistrationFailed
	}

	token, err := auth.GenerateToken(user.ID, user.Account)
	if err != nil {
		return "", ErrTokenGeneration
	}

	return token, nil
}

func (s *UserService) SendCaptcha(email string) error {
	code := utils.GenerateRandomCode(6)

	if err := cache.SetCaptcha(email, code); err != nil {
		return err
	}

	return sendEmail(email, code, "Your verification code is (valid for 2 minutes): ")
}

func (s *UserService) GetByID(userID int64) (*domain.User, error) {
	return s.userRepo.FindByAccount(fmt.Sprintf("%d", userID))
}

var (
	ErrUserNotFound       = NewServiceError(2003, "User not found")
	ErrInvalidPassword    = NewServiceError(2004, "Invalid password")
	ErrUserExists         = NewServiceError(2002, "User already exists")
	ErrInvalidCaptcha     = NewServiceError(2007, "Invalid captcha")
	ErrTokenGeneration    = NewServiceError(4001, "Token generation failed")
	ErrRegistrationFailed = NewServiceError(4001, "Registration failed")
)

type ServiceError struct {
	Code    int
	Message string
}

func NewServiceError(code int, message string) *ServiceError {
	return &ServiceError{Code: code, Message: message}
}

func (e *ServiceError) Error() string {
	return e.Message
}
