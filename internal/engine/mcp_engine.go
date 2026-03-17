package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"viperai/internal/config"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

type MCPEngine struct {
	model     *openai.ChatModel
	mcpClient *client.Client
	userID    int64
	baseURL   string
}

func NewMCPEngine(ctx context.Context, userID int64) (*MCPEngine, error) {
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

	return &MCPEngine{
		model:   llm,
		baseURL: "http://localhost:8081/mcp",
		userID:  userID,
	}, nil
}

func (e *MCPEngine) getMCPClient(ctx context.Context) (*client.Client, error) {
	if e.mcpClient == nil {
		httpTransport, err := transport.NewStreamableHTTP(e.baseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create MCP transport: %w", err)
		}

		e.mcpClient = client.NewClient(httpTransport)

		initRequest := mcp.InitializeRequest{}
		initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
		initRequest.Params.ClientInfo = mcp.Implementation{
			Name:    "ViperAI-MCP-Client",
			Version: "1.0.0",
		}
		initRequest.Params.Capabilities = mcp.ClientCapabilities{}

		if _, err := e.mcpClient.Initialize(ctx, initRequest); err != nil {
			return nil, fmt.Errorf("MCP client initialization failed: %w", err)
		}
	}
	return e.mcpClient, nil
}

func (e *MCPEngine) Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	if len(messages) == 0 {
		return nil, errors.New("no messages provided")
	}

	lastMessage := messages[len(messages)-1]
	query := lastMessage.Content

	firstPrompt := e.buildFirstPrompt(query)
	firstMessages := make([]*schema.Message, len(messages))
	copy(firstMessages, messages)
	firstMessages[len(firstMessages)-1] = &schema.Message{
		Role:    schema.User,
		Content: firstPrompt,
	}

	firstResp, err := e.model.Generate(ctx, firstMessages)
	if err != nil {
		return nil, fmt.Errorf("MCP first generate failed: %w", err)
	}

	log.Println("MCP first response:", firstResp)

	toolCall, err := e.parseAIResponse(firstResp.Content)
	if err != nil {
		log.Printf("Failed to parse AI response: %v", err)
		return firstResp, nil
	}

	if !toolCall.IsToolCall {
		log.Println("No tool call required")
		return firstResp, nil
	}

	log.Println("Tool call required:", toolCall)

	mcpClient, err := e.getMCPClient(ctx)
	if err != nil {
		log.Printf("MCP client error: %v", err)
		return firstResp, nil
	}

	toolResult, err := e.callMCPTool(ctx, mcpClient, toolCall.ToolName, toolCall.Args)
	if err != nil {
		log.Printf("MCP tool call failed: %v", err)
		return firstResp, nil
	}

	secondPrompt := e.buildSecondPrompt(query, toolCall.ToolName, toolCall.Args, toolResult)
	secondMessages := make([]*schema.Message, len(messages))
	copy(secondMessages, messages)
	secondMessages[len(secondMessages)-1] = &schema.Message{
		Role:    schema.User,
		Content: secondPrompt,
	}

	finalResp, err := e.model.Generate(ctx, secondMessages)
	if err != nil {
		return nil, fmt.Errorf("MCP second generate failed: %w", err)
	}

	log.Println("Final response:", finalResp)
	return finalResp, nil
}

func (e *MCPEngine) Stream(ctx context.Context, messages []*schema.Message, handler StreamHandler) (string, error) {
	if len(messages) == 0 {
		return "", errors.New("no messages provided")
	}

	lastMessage := messages[len(messages)-1]
	query := lastMessage.Content

	firstPrompt := e.buildFirstPrompt(query)
	firstMessages := make([]*schema.Message, len(messages))
	copy(firstMessages, messages)
	firstMessages[len(firstMessages)-1] = &schema.Message{
		Role:    schema.User,
		Content: firstPrompt,
	}

	firstResp, err := e.model.Generate(ctx, firstMessages)
	if err != nil {
		return "", fmt.Errorf("MCP first generate failed: %w", err)
	}

	toolCall, err := e.parseAIResponse(firstResp.Content)
	if err != nil {
		log.Printf("Failed to parse AI response: %v", err)
		return firstResp.Content, nil
	}

	if !toolCall.IsToolCall {
		return firstResp.Content, nil
	}

	mcpClient, err := e.getMCPClient(ctx)
	if err != nil {
		log.Printf("MCP client error: %v", err)
		return firstResp.Content, nil
	}

	toolResult, err := e.callMCPTool(ctx, mcpClient, toolCall.ToolName, toolCall.Args)
	if err != nil {
		log.Printf("MCP tool call failed: %v", err)
		return firstResp.Content, nil
	}

	secondPrompt := e.buildSecondPrompt(query, toolCall.ToolName, toolCall.Args, toolResult)
	secondMessages := make([]*schema.Message, len(messages))
	copy(secondMessages, messages)
	secondMessages[len(secondMessages)-1] = &schema.Message{
		Role:    schema.User,
		Content: secondPrompt,
	}

	stream, err := e.model.Stream(ctx, secondMessages)
	if err != nil {
		return "", fmt.Errorf("MCP stream failed: %w", err)
	}
	defer stream.Close()

	var result strings.Builder
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("MCP stream recv failed: %w", err)
		}
		if len(msg.Content) > 0 {
			result.WriteString(msg.Content)
			handler(msg.Content)
		}
	}

	return result.String(), nil
}

type ToolCallRequest struct {
	IsToolCall bool                   `json:"isToolCall"`
	ToolName   string                 `json:"toolName"`
	Args       map[string]interface{} `json:"args"`
}

func (e *MCPEngine) buildFirstPrompt(query string) string {
	return fmt.Sprintf(`You are an intelligent assistant that can call MCP tools to get information.

Available tools:
- get_weather: Get weather information for a specified city. Parameters: city (city name, supports Chinese and English, e.g., Beijing, Shanghai)

Important rules:
1. If you need to call a tool, you must strictly return the following JSON format:
{
  "isToolCall": true,
  "toolName": "tool_name",
  "args": {"param_name": "param_value"}
}
2. If you don't need to call a tool, return a natural language answer directly
3. Please decide whether you need to call a tool based on the user's question

User question: %s

Please call the appropriate tool if needed, then provide a comprehensive answer.`, query)
}

func (e *MCPEngine) buildSecondPrompt(query, toolName string, args map[string]interface{}, toolResult string) string {
	return fmt.Sprintf(`You are an intelligent assistant that can call MCP tools to get information.

Tool execution result:
Tool name: %s
Tool arguments: %v
Tool result: %s

User question: %s

Please provide a final comprehensive answer based on the tool result and user question.`, toolName, args, toolResult, query)
}

func (e *MCPEngine) parseAIResponse(response string) (*ToolCallRequest, error) {
	var toolCall ToolCallRequest
	if err := json.Unmarshal([]byte(response), &toolCall); err == nil {
		return &toolCall, nil
	}

	if strings.Contains(response, "get_weather") {
		city := e.extractCityFromResponse(response)
		if city != "" {
			return &ToolCallRequest{
				IsToolCall: true,
				ToolName:   "get_weather",
				Args:       map[string]interface{}{"city": city},
			}, nil
		}
	}

	return &ToolCallRequest{IsToolCall: false}, nil
}

func (e *MCPEngine) callMCPTool(ctx context.Context, c *client.Client, toolName string, args map[string]interface{}) (string, error) {
	callToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		},
	}

	result, err := c.CallTool(ctx, callToolRequest)
	if err != nil {
		return "", fmt.Errorf("MCP tool call failed: %w", err)
	}

	var text string
	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			text += textContent.Text + "\n"
		}
	}

	return text, nil
}

func (e *MCPEngine) extractCityFromResponse(response string) string {
	var toolCall ToolCallRequest
	if err := json.Unmarshal([]byte(response), &toolCall); err == nil {
		if args, ok := toolCall.Args["city"].(string); ok {
			return args
		}
	}

	return ""
}

func (e *MCPEngine) Type() string {
	return "mcp"
}

func (e *MCPEngine) Close() {
	if e.mcpClient != nil {
		e.mcpClient.Close()
	}
}
