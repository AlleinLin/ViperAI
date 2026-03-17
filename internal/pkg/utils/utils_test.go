package utils

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateRandomCode(t *testing.T) {
	t.Run("generate code with correct length", func(t *testing.T) {
		for length := 1; length <= 20; length++ {
			code := GenerateRandomCode(length)
			if len(code) != length {
				t.Errorf("Expected length %d, got %d", length, len(code))
			}
		}
	})

	t.Run("generate numeric code", func(t *testing.T) {
		code := GenerateRandomCode(100)
		for _, c := range code {
			if c < '0' || c > '9' {
				t.Errorf("Expected numeric character, got '%c'", c)
			}
		}
	})

	t.Run("generate different codes", func(t *testing.T) {
		codes := make(map[string]bool)
		for i := 0; i < 100; i++ {
			time.Sleep(time.Nanosecond)
			code := GenerateRandomCode(10)
			codes[code] = true
		}
		if len(codes) < 50 {
			t.Errorf("Expected mostly unique codes, got %d unique out of 100", len(codes))
		}
	})
}

func TestHashPassword(t *testing.T) {
	t.Run("hash password", func(t *testing.T) {
		password := "testpassword123"
		hash := HashPassword(password)

		if hash == "" {
			t.Error("Hash should not be empty")
		}

		if hash == password {
			t.Error("Hash should be different from password")
		}
	})

	t.Run("consistent hash", func(t *testing.T) {
		password := "testpassword123"
		hash1 := HashPassword(password)
		hash2 := HashPassword(password)

		if hash1 != hash2 {
			t.Error("Same password should produce same hash")
		}
	})

	t.Run("different passwords different hashes", func(t *testing.T) {
		password1 := "password1"
		password2 := "password2"
		hash1 := HashPassword(password1)
		hash2 := HashPassword(password2)

		if hash1 == hash2 {
			t.Error("Different passwords should produce different hashes")
		}
	})
}

func TestGenerateUUID(t *testing.T) {
	t.Run("generate non-empty UUID", func(t *testing.T) {
		uuid := GenerateUUID()
		if uuid == "" {
			t.Error("UUID should not be empty")
		}
	})

	t.Run("generate valid UUID format", func(t *testing.T) {
		uuid := GenerateUUID()
		if len(uuid) != 36 {
			t.Errorf("Expected UUID length 36, got %d", len(uuid))
		}

		parts := strings.Split(uuid, "-")
		if len(parts) != 5 {
			t.Errorf("Expected 5 parts in UUID, got %d", len(parts))
		}
	})

	t.Run("generate unique UUIDs", func(t *testing.T) {
		uuids := make(map[string]bool)
		for i := 0; i < 1000; i++ {
			uuid := GenerateUUID()
			uuids[uuid] = true
		}
		if len(uuids) != 1000 {
			t.Errorf("Expected 1000 unique UUIDs, got %d", len(uuids))
		}
	})
}
