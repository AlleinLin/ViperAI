package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"viperai/internal/domain"
	"viperai/internal/service"
	"viperai/internal/transport/http/response"

	"github.com/gin-gonic/gin"
)

type mockUserSvc struct{}

func (m *mockUserSvc) Login(account, password string) (string, error) {
	if account == "testuser" && password == "password123" {
		return "mock-token", nil
	}
	return "", service.ErrInvalidPassword
}

func (m *mockUserSvc) Register(email, password, captcha string) (string, error) {
	return "mock-token", nil
}

func (m *mockUserSvc) SendCaptcha(email string) error {
	return nil
}

func (m *mockUserSvc) GetByID(userID int64) (*domain.User, error) {
	return &domain.User{ID: userID, Name: "Test User"}, nil
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestUserHandler_Login(t *testing.T) {
	router := setupTestRouter()
	mockSvc := &mockUserSvc{}
	handler := NewUserHandler(mockSvc)

	router.POST("/login", handler.Login)

	t.Run("successful login", func(t *testing.T) {
		body := LoginRequest{Account: "testuser", Password: "password123"}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("missing parameters", func(t *testing.T) {
		body := LoginRequest{Account: "", Password: ""}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var resp response.Response
		json.Unmarshal(w.Body.Bytes(), &resp)

		if resp.Code == response.CodeSuccess {
			t.Error("Should not succeed with empty parameters")
		}
	})
}

func TestUserHandler_SendCaptcha(t *testing.T) {
	router := setupTestRouter()
	mockSvc := &mockUserSvc{}
	handler := NewUserHandler(mockSvc)

	router.POST("/captcha", handler.SendCaptcha)

	t.Run("send captcha", func(t *testing.T) {
		body := CaptchaRequest{Email: "test@example.com"}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest(http.MethodPost, "/captcha", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}
