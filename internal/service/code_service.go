package service

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type CodeService struct {
	timeout     time.Duration
	maxOutput   int
	sandboxDir  string
	mu          sync.Mutex
}

type CodeExecutionRequest struct {
	Language string `json:"language"`
	Code     string `json:"code"`
	Input    string `json:"input,omitempty"`
}

type CodeExecutionResult struct {
	Success   bool     `json:"success"`
	Output    string   `json:"output"`
	Error     string   `json:"error,omitempty"`
	Duration  int64    `json:"duration_ms"`
	Language  string   `json:"language"`
	TestCases []TestCase `json:"test_cases,omitempty"`
}

type TestCase struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Passed   bool   `json:"passed"`
}

func NewCodeService() *CodeService {
	sandboxDir := filepath.Join(os.TempDir(), "viperai_sandbox")
	os.MkdirAll(sandboxDir, 0755)

	return &CodeService{
		timeout:    30 * time.Second,
		maxOutput:  10000,
		sandboxDir: sandboxDir,
	}
}

func (s *CodeService) Execute(ctx context.Context, req *CodeExecutionRequest) (*CodeExecutionResult, error) {
	switch strings.ToLower(req.Language) {
	case "python", "python3":
		return s.executePython(ctx, req)
	case "javascript", "js", "node":
		return s.executeJavaScript(ctx, req)
	case "go", "golang":
		return s.executeGo(ctx, req)
	case "java":
		return s.executeJava(ctx, req)
	case "cpp", "c++":
		return s.executeCpp(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported language: %s", req.Language)
	}
}

func (s *CodeService) executePython(ctx context.Context, req *CodeExecutionRequest) (*CodeExecutionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := fmt.Sprintf("code_%d.py", time.Now().UnixNano())
	filePath := filepath.Join(s.sandboxDir, filename)
	defer os.Remove(filePath)

	if err := os.WriteFile(filePath, []byte(req.Code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write code file: %w", err)
	}

	start := time.Now()
	cmd := exec.CommandContext(ctx, "python3", filePath)

	if req.Input != "" {
		cmd.Stdin = strings.NewReader(req.Input)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start).Milliseconds()

	result := &CodeExecutionResult{
		Duration: duration,
		Language: "python",
	}

	if err != nil {
		result.Success = false
		result.Error = stderr.String()
		if result.Error == "" {
			result.Error = err.Error()
		}
	} else {
		result.Success = true
		result.Output = s.truncateOutput(stdout.String())
	}

	return result, nil
}

func (s *CodeService) executeJavaScript(ctx context.Context, req *CodeExecutionRequest) (*CodeExecutionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := fmt.Sprintf("code_%d.js", time.Now().UnixNano())
	filePath := filepath.Join(s.sandboxDir, filename)
	defer os.Remove(filePath)

	if err := os.WriteFile(filePath, []byte(req.Code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write code file: %w", err)
	}

	start := time.Now()
	cmd := exec.CommandContext(ctx, "node", filePath)

	if req.Input != "" {
		cmd.Stdin = strings.NewReader(req.Input)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start).Milliseconds()

	result := &CodeExecutionResult{
		Duration: duration,
		Language: "javascript",
	}

	if err != nil {
		result.Success = false
		result.Error = stderr.String()
		if result.Error == "" {
			result.Error = err.Error()
		}
	} else {
		result.Success = true
		result.Output = s.truncateOutput(stdout.String())
	}

	return result, nil
}

func (s *CodeService) executeGo(ctx context.Context, req *CodeExecutionRequest) (*CodeExecutionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dirName := fmt.Sprintf("go_%d", time.Now().UnixNano())
	dirPath := filepath.Join(s.sandboxDir, dirName)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}
	defer os.RemoveAll(dirPath)

	mainFile := filepath.Join(dirPath, "main.go")
	if err := os.WriteFile(mainFile, []byte(req.Code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write code file: %w", err)
	}

	execFile := filepath.Join(dirPath, "main")

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", execFile, mainFile)
	buildCmd.Dir = dirPath
	var buildStderr bytes.Buffer
	buildCmd.Stderr = &buildStderr

	if err := buildCmd.Run(); err != nil {
		return &CodeExecutionResult{
			Success:  false,
			Error:    buildStderr.String(),
			Language: "go",
		}, nil
	}

	start := time.Now()
	cmd := exec.CommandContext(ctx, execFile)

	if req.Input != "" {
		cmd.Stdin = strings.NewReader(req.Input)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start).Milliseconds()

	result := &CodeExecutionResult{
		Duration: duration,
		Language: "go",
	}

	if err != nil {
		result.Success = false
		result.Error = stderr.String()
		if result.Error == "" {
			result.Error = err.Error()
		}
	} else {
		result.Success = true
		result.Output = s.truncateOutput(stdout.String())
	}

	return result, nil
}

func (s *CodeService) executeJava(ctx context.Context, req *CodeExecutionRequest) (*CodeExecutionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	className := "Main"
	filename := fmt.Sprintf("%s.java", className)
	filePath := filepath.Join(s.sandboxDir, filename)
	defer os.Remove(filePath)

	if err := os.WriteFile(filePath, []byte(req.Code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write code file: %w", err)
	}

	compileCmd := exec.CommandContext(ctx, "javac", filePath)
	var compileStderr bytes.Buffer
	compileCmd.Stderr = &compileStderr

	if err := compileCmd.Run(); err != nil {
		return &CodeExecutionResult{
			Success:  false,
			Error:    compileStderr.String(),
			Language: "java",
		}, nil
	}
	defer os.Remove(filepath.Join(s.sandboxDir, fmt.Sprintf("%s.class", className)))

	start := time.Now()
	cmd := exec.CommandContext(ctx, "java", "-cp", s.sandboxDir, className)

	if req.Input != "" {
		cmd.Stdin = strings.NewReader(req.Input)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start).Milliseconds()

	result := &CodeExecutionResult{
		Duration: duration,
		Language: "java",
	}

	if err != nil {
		result.Success = false
		result.Error = stderr.String()
		if result.Error == "" {
			result.Error = err.Error()
		}
	} else {
		result.Success = true
		result.Output = s.truncateOutput(stdout.String())
	}

	return result, nil
}

func (s *CodeService) executeCpp(ctx context.Context, req *CodeExecutionRequest) (*CodeExecutionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := fmt.Sprintf("code_%d.cpp", time.Now().UnixNano())
	filePath := filepath.Join(s.sandboxDir, filename)
	execFile := filepath.Join(s.sandboxDir, fmt.Sprintf("code_%d", time.Now().UnixNano()))
	defer os.Remove(filePath)
	defer os.Remove(execFile)

	if err := os.WriteFile(filePath, []byte(req.Code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write code file: %w", err)
	}

	compileCmd := exec.CommandContext(ctx, "g++", "-o", execFile, filePath)
	var compileStderr bytes.Buffer
	compileCmd.Stderr = &compileStderr

	if err := compileCmd.Run(); err != nil {
		return &CodeExecutionResult{
			Success:  false,
			Error:    compileStderr.String(),
			Language: "cpp",
		}, nil
	}

	start := time.Now()
	cmd := exec.CommandContext(ctx, execFile)

	if req.Input != "" {
		cmd.Stdin = strings.NewReader(req.Input)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start).Milliseconds()

	result := &CodeExecutionResult{
		Duration: duration,
		Language: "cpp",
	}

	if err != nil {
		result.Success = false
		result.Error = stderr.String()
		if result.Error == "" {
			result.Error = err.Error()
		}
	} else {
		result.Success = true
		result.Output = s.truncateOutput(stdout.String())
	}

	return result, nil
}

func (s *CodeService) RunTests(ctx context.Context, req *CodeExecutionRequest, testCases []TestCase) (*CodeExecutionResult, error) {
	results := make([]TestCase, len(testCases))

	for i, tc := range testCases {
		testReq := &CodeExecutionRequest{
			Language: req.Language,
			Code:     req.Code,
			Input:    tc.Input,
		}

		result, err := s.Execute(ctx, testReq)
		if err != nil {
			results[i] = TestCase{
				Input:    tc.Input,
				Expected: tc.Expected,
				Actual:   "",
				Passed:   false,
			}
			continue
		}

		actual := strings.TrimSpace(result.Output)
		expected := strings.TrimSpace(tc.Expected)

		results[i] = TestCase{
			Input:    tc.Input,
			Expected: tc.Expected,
			Actual:   actual,
			Passed:   actual == expected,
		}
	}

	return &CodeExecutionResult{
		Success:   true,
		Language:  req.Language,
		TestCases: results,
	}, nil
}

func (s *CodeService) truncateOutput(output string) string {
	if len(output) > s.maxOutput {
		return output[:s.maxOutput] + "\n... (output truncated)"
	}
	return output
}

func (s *CodeService) GetSupportedLanguages() []string {
	return []string{"python", "javascript", "go", "java", "cpp"}
}

func (s *CodeService) AnalyzeCode(language, code string) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"language":     language,
		"lines":        len(strings.Split(code, "\n")),
		"characters":   len(code),
		"suggestions":  []string{},
		"complexity":   "low",
	}

	if strings.Contains(code, "for") || strings.Contains(code, "while") {
		result["complexity"] = "medium"
	}
	if strings.Contains(code, "for") && strings.Contains(code, "for") {
		result["complexity"] = "high"
	}

	suggestions := []string{}
	if strings.Contains(code, "print") {
		suggestions = append(suggestions, "Consider using logging instead of print statements")
	}
	if strings.Contains(code, "TODO") {
		suggestions = append(suggestions, "Code contains TODO comments")
	}
	if strings.Contains(code, "password") || strings.Contains(code, "secret") {
		suggestions = append(suggestions, "Warning: Code may contain sensitive information")
	}

	result["suggestions"] = suggestions

	return result, nil
}

func (s *CodeService) FormatCode(language, code string) (string, error) {
	switch strings.ToLower(language) {
	case "python":
		return s.formatPython(code)
	case "javascript", "js":
		return s.formatJavaScript(code)
	case "go":
		return s.formatGo(code)
	default:
		return code, nil
	}
}

func (s *CodeService) formatPython(code string) (string, error) {
	return code, nil
}

func (s *CodeService) formatJavaScript(code string) (string, error) {
	return code, nil
}

func (s *CodeService) formatGo(code string) (string, error) {
	return code, nil
}
