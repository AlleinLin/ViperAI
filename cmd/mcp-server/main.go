package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"viperai/internal/infrastructure/mcp"
)

func main() {
	port := 8081
	if p := os.Getenv("MCP_SERVER_PORT"); p != "" {
		fmt.Sscanf(p, "%d", &port)
	}

	server := mcp.NewServer(port)

	mcp.RegisterDefaultTools(server)

	registerCustomTools(server)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down MCP server...")
		cancel()
	}()

	log.Printf("Starting ViperAI MCP Server on port %d", port)
	if err := server.Start(ctx); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}

func registerCustomTools(server *mcp.Server) {
	server.RegisterTool(mcp.ToolDefinition{
		Name:        "search_knowledge",
		Description: "Search knowledge base for relevant information",
		Parameters: map[string]mcp.ParameterDef{
			"query": {
				Type:        "string",
				Description: "Search query string",
				Required:    true,
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			query, _ := args["query"].(string)
			return fmt.Sprintf(`{"query": "%s", "results": [], "message": "Knowledge search completed"}`, query), nil
		},
	})

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "execute_code",
		Description: "Execute code in a sandboxed environment",
		Parameters: map[string]mcp.ParameterDef{
			"language": {
				Type:        "string",
				Description: "Programming language (python, javascript, go)",
				Required:    true,
			},
			"code": {
				Type:        "string",
				Description: "Code to execute",
				Required:    true,
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			language, _ := args["language"].(string)
			code, _ := args["code"].(string)
			return fmt.Sprintf(`{"language": "%s", "status": "executed", "output": "Code execution simulated", "code_length": %d}`, language, len(code)), nil
		},
	})
}
