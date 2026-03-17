package auth

import (
	"testing"
)

func TestGenerateAndParseToken(t *testing.T) {
	t.Run("generate and parse token", func(t *testing.T) {
		userID := int64(123)
		account := "testuser"

		token, err := GenerateToken(userID, account)
		if err != nil {
			t.Errorf("Failed to generate token: %v", err)
			return
		}

		if token == "" {
			t.Error("Token should not be empty")
			return
		}

		claims, ok := ParseToken(token)
		if !ok {
			t.Error("Failed to parse token")
			return
		}

		if claims.UserID != userID {
			t.Errorf("Expected userID %d, got %d", userID, claims.UserID)
		}

		if claims.Account != account {
			t.Errorf("Expected account '%s', got '%s'", account, claims.Account)
		}
	})

	t.Run("parse invalid token", func(t *testing.T) {
		_, ok := ParseToken("invalid-token")
		if ok {
			t.Error("Should fail to parse invalid token")
		}
	})

	t.Run("parse empty token", func(t *testing.T) {
		_, ok := ParseToken("")
		if ok {
			t.Error("Should fail to parse empty token")
		}
	})

	t.Run("generate multiple tokens", func(t *testing.T) {
		token1, err := GenerateToken(1, "user1")
		if err != nil {
			t.Errorf("Failed to generate token1: %v", err)
			return
		}

		token2, err := GenerateToken(2, "user2")
		if err != nil {
			t.Errorf("Failed to generate token2: %v", err)
			return
		}

		if token1 == token2 {
			t.Error("Different users should have different tokens")
		}
	})
}
