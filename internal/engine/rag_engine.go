package engine

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"viperai/internal/config"
	"viperai/internal/infrastructure/cache"

	embeddingArk "github.com/cloudwego/eino-ext/components/embedding/ark"
	redisIndexer "github.com/cloudwego/eino-ext/components/indexer/redis"
	redisRetriever "github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	redis "github.com/redis/go-redis/v9"
)

type RAGIndexer struct {
	embedder embedding.Embedder
	indexer   *redisIndexer.Indexer
}

type RAGRetriever struct {
	embedder  embedding.Embedder
	retriever retriever.Retriever
}

func NewRAGIndexer(filename, embeddingModel string) (*RAGIndexer, error) {
	ctx := context.Background()
	apiKey := os.Getenv("OPENAI_API_KEY")
	cfg := config.Get().AIModel

	embedConfig := &embeddingArk.EmbeddingConfig{
		BaseURL: cfg.BaseURL,
		APIKey:  apiKey,
		Model:   embeddingModel,
	}

	embedder, err := embeddingArk.NewEmbedder(ctx, embedConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	if err := cache.CreateVectorIndex(ctx, filename, cfg.Dimension); err != nil {
		return nil, fmt.Errorf("failed to create vector index: %w", err)
	}

	rdb := cache.Client

	indexerConfig := &redisIndexer.IndexerConfig{
		Client:    rdb,
		KeyPrefix: cache.GenerateIndexPrefix(filename),
		BatchSize: 10,
		DocumentToHashes: func(ctx context.Context, doc *schema.Document) (*redisIndexer.Hashes, error) {
			source := ""
			if s, ok := doc.MetaData["source"].(string); ok {
				source = s
			}

			return &redisIndexer.Hashes{
				Key: fmt.Sprintf("%s:%s", filename, doc.ID),
				Field2Value: map[string]redisIndexer.FieldValue{
					"content":  {Value: doc.Content, EmbedKey: "vector"},
					"metadata": {Value: source},
				},
			}, nil
		},
	}
	indexerConfig.Embedding = embedder

	idx, err := redisIndexer.NewIndexer(ctx, indexerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create indexer: %w", err)
	}

	return &RAGIndexer{
		embedder: embedder,
		indexer:   idx,
	}, nil
}

func (r *RAGIndexer) IndexFile(ctx context.Context, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	doc := &schema.Document{
		ID:      "doc_1",
		Content: string(content),
		MetaData: map[string]any{
			"source": filePath,
		},
	}

	_, err = r.indexer.Store(ctx, []*schema.Document{doc})
	return err
}

func DeleteRAGIndex(ctx context.Context, filename string) error {
	return cache.DeleteVectorIndex(ctx, filename)
}

func NewRAGRetriever(ctx context.Context, userID int64) (*RAGRetriever, error) {
	cfg := config.Get().AIModel
	apiKey := os.Getenv("OPENAI_API_KEY")

	embedConfig := &embeddingArk.EmbeddingConfig{
		BaseURL: cfg.BaseURL,
		APIKey:  apiKey,
		Model:   cfg.EmbeddingModel,
	}
	embedder, err := embeddingArk.NewEmbedder(ctx, embedConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	userDir := fmt.Sprintf("uploads/%d", userID)
	files, err := os.ReadDir(userDir)
	if err != nil || len(files) == 0 {
		return nil, fmt.Errorf("no uploaded file found for user %d", userID)
	}

	var filename string
	for _, f := range files {
		if !f.IsDir() {
			filename = f.Name()
			break
		}
	}

	if filename == "" {
		return nil, fmt.Errorf("no valid file found for user %d", userID)
	}

	rdb := cache.Client
	indexName := cache.GenerateIndexName(filename)

	retrieverConfig := &redisRetriever.RetrieverConfig{
		Client:       rdb,
		Index:        indexName,
		Dialect:      2,
		ReturnFields: []string{"content", "metadata", "distance"},
		TopK:         5,
		VectorField:  "vector",
		DocumentConverter: func(ctx context.Context, doc redis.Document) (*schema.Document, error) {
			resp := &schema.Document{
				ID:       doc.ID,
				Content:  "",
				MetaData: map[string]any{},
			}
			for field, val := range doc.Fields {
				if field == "content" {
					resp.Content = val
				} else {
					resp.MetaData[field] = val
				}
			}
			return resp, nil
		},
	}
	retrieverConfig.Embedding = embedder

	rtr, err := redisRetriever.NewRetriever(ctx, retrieverConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create retriever: %w", err)
	}

	return &RAGRetriever{
		embedder:  embedder,
		retriever: rtr,
	}, nil
}

func (r *RAGRetriever) Retrieve(ctx context.Context, query string) ([]*schema.Document, error) {
	return r.retriever.Retrieve(ctx, query)
}

func BuildRAGPrompt(query string, docs []*schema.Document) string {
	if len(docs) == 0 {
		return query
	}

	var contextText strings.Builder
	for i, doc := range docs {
		contextText.WriteString(fmt.Sprintf("[Document %d]: %s\n\n", i+1, doc.Content))
	}

	return fmt.Sprintf(`Based on the following reference documents, answer the user's question. If the documents don't contain relevant information, please state that.

Reference Documents:
%s

User Question: %s

Please provide an accurate and complete answer:`, contextText.String(), query)
}

type RAGEngine struct {
	model    *openai.ChatModel
	userID   int64
}

func NewRAGEngine(ctx context.Context, userID int64) (*RAGEngine, error) {
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

	return &RAGEngine{
		model:  llm,
		userID: userID,
	}, nil
}

func (e *RAGEngine) Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	retriever, err := NewRAGRetriever(ctx, e.userID)
	if err != nil {
		log.Printf("RAG retriever error (user may not have uploaded file): %v", err)
		return e.model.Generate(ctx, messages)
	}

	if len(messages) == 0 {
		return nil, errors.New("no messages provided")
	}

	lastMessage := messages[len(messages)-1]
	docs, err := retriever.Retrieve(ctx, lastMessage.Content)
	if err != nil {
		log.Printf("Document retrieval failed: %v", err)
		return e.model.Generate(ctx, messages)
	}

	ragPrompt := BuildRAGPrompt(lastMessage.Content, docs)

	ragMessages := make([]*schema.Message, len(messages))
	copy(ragMessages, messages)
	ragMessages[len(ragMessages)-1] = &schema.Message{
		Role:    schema.User,
		Content: ragPrompt,
	}

	return e.model.Generate(ctx, ragMessages)
}

func (e *RAGEngine) Stream(ctx context.Context, messages []*schema.Message, handler StreamHandler) (string, error) {
	retriever, err := NewRAGRetriever(ctx, e.userID)
	if err != nil {
		log.Printf("RAG retriever error: %v", err)
		return e.streamWithoutRAG(ctx, messages, handler)
	}

	if len(messages) == 0 {
		return "", errors.New("no messages provided")
	}

	lastMessage := messages[len(messages)-1]
	docs, err := retriever.Retrieve(ctx, lastMessage.Content)
	if err != nil {
		log.Printf("Document retrieval failed: %v", err)
		return e.streamWithoutRAG(ctx, messages, handler)
	}

	ragPrompt := BuildRAGPrompt(lastMessage.Content, docs)

	ragMessages := make([]*schema.Message, len(messages))
	copy(ragMessages, messages)
	ragMessages[len(ragMessages)-1] = &schema.Message{
		Role:    schema.User,
		Content: ragPrompt,
	}

	stream, err := e.model.Stream(ctx, ragMessages)
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

func (e *RAGEngine) streamWithoutRAG(ctx context.Context, messages []*schema.Message, handler StreamHandler) (string, error) {
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

func (e *RAGEngine) Type() string {
	return "rag"
}
