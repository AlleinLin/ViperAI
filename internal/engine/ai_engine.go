package engine

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"

	"viperai/internal/config"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type StreamHandler func(chunk string)

type AIEngine interface {
	Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error)
	Stream(ctx context.Context, messages []*schema.Message, handler StreamHandler) (string, error)
	Type() string
}

type OpenAIEngine struct {
	model model.ToolCallingChatModel
}

func NewOpenAIEngine(ctx context.Context) (*OpenAIEngine, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	modelName := os.Getenv("OPENAI_MODEL_NAME")
	baseURL := os.Getenv("OPENAI_BASE_URL")

	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: baseURL,
		Model:   modelName,
		APIKey:  apiKey,
	})
	if err != nil {
		return nil, err
	}

	return &OpenAIEngine{model: llm}, nil
}

func (e *OpenAIEngine) Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	return e.model.Generate(ctx, messages)
}

func (e *OpenAIEngine) Stream(ctx context.Context, messages []*schema.Message, handler StreamHandler) (string, error) {
	stream, err := e.model.Stream(ctx, messages)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	var result strings.Builder
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if len(msg.Content) > 0 {
			result.WriteString(msg.Content)
			handler(msg.Content)
		}
	}

	return result.String(), nil
}

func (e *OpenAIEngine) Type() string {
	return "openai"
}

type OllamaEngine struct {
	model model.ToolCallingChatModel
}

func NewOllamaEngine(ctx context.Context, baseURL, modelName string) (*OllamaEngine, error) {
	llm, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: baseURL,
		Model:   modelName,
	})
	if err != nil {
		return nil, err
	}

	return &OllamaEngine{model: llm}, nil
}

func (e *OllamaEngine) Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	return e.model.Generate(ctx, messages)
}

func (e *OllamaEngine) Stream(ctx context.Context, messages []*schema.Message, handler StreamHandler) (string, error) {
	stream, err := e.model.Stream(ctx, messages)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	var result strings.Builder
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if len(msg.Content) > 0 {
			result.WriteString(msg.Content)
			handler(msg.Content)
		}
	}

	return result.String(), nil
}

func (e *OllamaEngine) Type() string {
	return "ollama"
}

type AliEngine struct {
	model model.ToolCallingChatModel
}

func NewAliEngine(ctx context.Context) (*AliEngine, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	cfg := config.Get().AIModel

	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: cfg.BaseURL,
		Model:   cfg.ChatModel,
		APIKey:  apiKey,
	})
	if err != nil {
		return nil, err
	}

	return &AliEngine{model: llm}, nil
}

func (e *AliEngine) Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	return e.model.Generate(ctx, messages)
}

func (e *AliEngine) Stream(ctx context.Context, messages []*schema.Message, handler StreamHandler) (string, error) {
	stream, err := e.model.Stream(ctx, messages)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	var result strings.Builder
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if len(msg.Content) > 0 {
			result.WriteString(msg.Content)
			handler(msg.Content)
		}
	}

	return result.String(), nil
}

func (e *AliEngine) Type() string {
	return "ali"
}

type EngineFactory struct {
	creators map[string]func(ctx context.Context, opts map[string]interface{}) (AIEngine, error)
}

func NewEngineFactory() *EngineFactory {
	factory := &EngineFactory{
		creators: make(map[string]func(ctx context.Context, opts map[string]interface{}) (AIEngine, error)),
	}
	factory.registerDefaults()
	return factory
}

func (f *EngineFactory) registerDefaults() {
	f.creators["1"] = func(ctx context.Context, opts map[string]interface{}) (AIEngine, error) {
		return NewOpenAIEngine(ctx)
	}

	f.creators["2"] = func(ctx context.Context, opts map[string]interface{}) (AIEngine, error) {
		userID, ok := opts["user_id"].(int64)
		if !ok {
			return nil, errors.New("user_id is required for RAG engine")
		}
		return NewRAGEngine(ctx, userID)
	}

	f.creators["3"] = func(ctx context.Context, opts map[string]interface{}) (AIEngine, error) {
		userID, ok := opts["user_id"].(int64)
		if !ok {
			return nil, errors.New("user_id is required for MCP engine")
		}
		return NewMCPEngine(ctx, userID)
	}

	f.creators["4"] = func(ctx context.Context, opts map[string]interface{}) (AIEngine, error) {
		baseURL, _ := opts["base_url"].(string)
		modelName, ok := opts["model_name"].(string)
		if !ok {
			return nil, errors.New("model_name is required for ollama engine")
		}
		return NewOllamaEngine(ctx, baseURL, modelName)
	}
}

func (f *EngineFactory) Create(engineType string, ctx context.Context, opts map[string]interface{}) (AIEngine, error) {
	creator, ok := f.creators[engineType]
	if !ok {
		return nil, errors.New("unsupported engine type: " + engineType)
	}
	return creator(ctx, opts)
}

func (f *EngineFactory) Register(engineType string, creator func(ctx context.Context, opts map[string]interface{}) (AIEngine, error)) {
	f.creators[engineType] = creator
}

var defaultFactory *EngineFactory

func GetFactory() *EngineFactory {
	if defaultFactory == nil {
		defaultFactory = NewEngineFactory()
	}
	return defaultFactory
}
