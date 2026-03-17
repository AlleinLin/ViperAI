package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	port     int
	handlers map[string]ToolHandler
	mu       sync.RWMutex
}

type ToolHandler func(ctx context.Context, args map[string]interface{}) (string, error)

type ToolDefinition struct {
	Name        string
	Description string
	Parameters  map[string]ParameterDef
	Handler     ToolHandler
}

type ParameterDef struct {
	Type        string
	Description string
	Required    bool
}

func NewServer(port int) *Server {
	return &Server{
		port:     port,
		handlers: make(map[string]ToolHandler),
	}
}

func (s *Server) RegisterTool(def ToolDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.handlers[def.Name]; exists {
		return fmt.Errorf("tool %s already registered", def.Name)
	}

	s.handlers[def.Name] = def.Handler

	log.Printf("Tool registered: %s", def.Name)
	return nil
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/mcp", s.handleMCP)
	mux.HandleFunc("/health", s.handleHealth)

	addr := fmt.Sprintf(":%d", s.port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	log.Printf("MCP Server starting on port %d", s.port)
	return srv.ListenAndServe()
}

func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	switch req.Method {
	case "initialize":
		s.handleInitialize(w, req)
	case "tools/list":
		s.handleToolsList(w, req)
	case "tools/call":
		s.handleToolsCall(w, r, req)
	default:
		http.Error(w, fmt.Sprintf("Unknown method: %s", req.Method), http.StatusBadRequest)
	}
}

func (s *Server) handleInitialize(w http.ResponseWriter, req MCPRequest) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "ViperAI-MCP-Server",
				"version": "1.0.0",
			},
		},
	}
	s.writeJSON(w, resp)
}

func (s *Server) handleToolsList(w http.ResponseWriter, req MCPRequest) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]map[string]interface{}, 0, len(s.handlers))
	for name := range s.handlers {
		tools = append(tools, map[string]interface{}{
			"name":        name,
			"description": fmt.Sprintf("Tool: %s", name),
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		})
	}

	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
	s.writeJSON(w, resp)
}

func (s *Server) handleToolsCall(w http.ResponseWriter, r *http.Request, req MCPRequest) {
	params, ok := req.Params.(map[string]interface{})
	if !ok {
		http.Error(w, "Invalid params", http.StatusBadRequest)
		return
	}

	toolName, _ := params["name"].(string)
	args, _ := params["arguments"].(map[string]interface{})

	s.mu.RLock()
	handler, exists := s.handlers[toolName]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, fmt.Sprintf("Tool not found: %s", toolName), http.StatusNotFound)
		return
	}

	result, err := handler(r.Context(), args)
	if err != nil {
		resp := MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -1,
				Message: err.Error(),
			},
		}
		s.writeJSON(w, resp)
		return
	}

	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": result,
				},
			},
		},
	}
	s.writeJSON(w, resp)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func RegisterDefaultTools(srv *Server) {
	srv.RegisterTool(ToolDefinition{
		Name:        "get_weather",
		Description: "Get current weather information for a specified city",
		Parameters: map[string]ParameterDef{
			"city": {
				Type:        "string",
				Description: "City name (supports Chinese and English, e.g., Beijing, Shanghai)",
				Required:    true,
			},
		},
		Handler: handleGetWeather,
	})

	srv.RegisterTool(ToolDefinition{
		Name:        "get_time",
		Description: "Get current time for a specified timezone",
		Parameters: map[string]ParameterDef{
			"timezone": {
				Type:        "string",
				Description: "Timezone name (e.g., Asia/Shanghai, America/New_York)",
				Required:    false,
			},
		},
		Handler: handleGetTime,
	})

	srv.RegisterTool(ToolDefinition{
		Name:        "calculate",
		Description: "Perform mathematical calculations",
		Parameters: map[string]ParameterDef{
			"expression": {
				Type:        "string",
				Description: "Mathematical expression to evaluate",
				Required:    true,
			},
		},
		Handler: handleCalculate,
	})
}

func handleGetWeather(ctx context.Context, args map[string]interface{}) (string, error) {
	city, ok := args["city"].(string)
	if !ok {
		return "", fmt.Errorf("city parameter is required")
	}

	weather := fmt.Sprintf(`{"city": "%s", "temperature": "22°C", "condition": "Sunny", "humidity": "65%%", "wind": "Light breeze", "update_time": "%s"}`, 
		city, time.Now().Format("2006-01-02 15:04:05"))
	
	return weather, nil
}

func handleGetTime(ctx context.Context, args map[string]interface{}) (string, error) {
	timezone := "Local"
	if tz, ok := args["timezone"].(string); ok && tz != "" {
		timezone = tz
	}

	now := time.Now()
	if timezone != "Local" {
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			return "", fmt.Errorf("invalid timezone: %s", timezone)
		}
		now = now.In(loc)
	}

	return fmt.Sprintf(`{"timezone": "%s", "current_time": "%s", "unix_timestamp": %d}`,
		timezone, now.Format("2006-01-02 15:04:05"), now.Unix()), nil
}

func handleCalculate(ctx context.Context, args map[string]interface{}) (string, error) {
	expr, ok := args["expression"].(string)
	if !ok {
		return "", fmt.Errorf("expression parameter is required")
	}

	return fmt.Sprintf(`{"expression": "%s", "result": "calculated_value", "note": "This is a mock calculator"}`, expr), nil
}
