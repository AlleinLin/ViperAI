package engine

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestEngineFactory_Create(t *testing.T) {
	factory := NewEngineFactory()

	t.Run("unsupported engine type", func(t *testing.T) {
		_, err := factory.Create("invalid", context.Background(), nil)
		if err == nil {
			t.Error("Expected error for unsupported engine type")
		}
	})

	t.Run("registered engine types", func(t *testing.T) {
		types := []string{"1", "2", "3", "4"}
		for _, engineType := range types {
			if _, ok := factory.creators[engineType]; !ok {
				t.Errorf("Engine type %s not registered", engineType)
			}
		}
	})
}

func TestEngineFactory_Register(t *testing.T) {
	factory := NewEngineFactory()

	customType := "custom"
	factory.Register(customType, func(ctx context.Context, opts map[string]interface{}) (AIEngine, error) {
		return nil, nil
	})

	if _, ok := factory.creators[customType]; !ok {
		t.Error("Custom engine type should be registered")
	}
}

func TestGetFactory(t *testing.T) {
	f1 := GetFactory()
	f2 := GetFactory()

	if f1 != f2 {
		t.Error("GetFactory should return singleton instance")
	}
}

type mockEngine struct {
	engineType string
}

func (m *mockEngine) Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	return &schema.Message{Content: "mock response"}, nil
}

func (m *mockEngine) Stream(ctx context.Context, messages []*schema.Message, handler StreamHandler) (string, error) {
	handler("mock ")
	handler("stream")
	return "mock stream", nil
}

func (m *mockEngine) Type() string {
	return m.engineType
}

func TestMockEngine(t *testing.T) {
	engine := &mockEngine{engineType: "test"}

	t.Run("type", func(t *testing.T) {
		if engine.Type() != "test" {
			t.Errorf("Expected type 'test', got '%s'", engine.Type())
		}
	})

	t.Run("generate", func(t *testing.T) {
		resp, err := engine.Generate(context.Background(), nil)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if resp.Content != "mock response" {
			t.Errorf("Expected 'mock response', got '%s'", resp.Content)
		}
	})

	t.Run("stream", func(t *testing.T) {
		var result string
		resp, err := engine.Stream(context.Background(), nil, func(chunk string) {
			result += chunk
		})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if resp != "mock stream" {
			t.Errorf("Expected 'mock stream', got '%s'", resp)
		}
		if result != "mock stream" {
			t.Errorf("Expected streamed 'mock stream', got '%s'", result)
		}
	})
}
