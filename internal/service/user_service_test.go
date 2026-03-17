package service

import (
	"errors"
	"testing"

	"viperai/internal/domain"
)

type UserRepository interface {
	Create(user *domain.User) error
	FindByAccount(account string) (*domain.User, error)
	ExistsByAccount(account string) bool
	ExistsByEmail(email string) bool
}

type mockUserRepository struct {
	users map[int64]*domain.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[int64]*domain.User),
	}
}

func (m *mockUserRepository) Create(user *domain.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) FindByAccount(account string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Account == account {
			return u, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockUserRepository) ExistsByAccount(account string) bool {
	for _, u := range m.users {
		if u.Account == account {
			return true
		}
	}
	return false
}

func (m *mockUserRepository) ExistsByEmail(email string) bool {
	for _, u := range m.users {
		if u.Email == email {
			return true
		}
	}
	return false
}

func TestServiceError(t *testing.T) {
	err := NewServiceError(1001, "test error")
	if err.Code != 1001 {
		t.Errorf("Expected code 1001, got %d", err.Code)
	}
	if err.Message != "test error" {
		t.Errorf("Expected message 'test error', got '%s'", err.Message)
	}
	if err.Error() != "test error" {
		t.Errorf("Expected error string 'test error', got '%s'", err.Error())
	}
}

func TestNewServiceError(t *testing.T) {
	err := NewServiceError(2001, "parameter error")
	if err == nil {
		t.Error("NewServiceError should not return nil")
	}
}
