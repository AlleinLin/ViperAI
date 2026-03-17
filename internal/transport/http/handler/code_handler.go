package handler

import (
	"viperai/internal/service"
	"viperai/internal/transport/http/response"

	"github.com/gin-gonic/gin"
)

type CodeHandler struct {
	codeService *service.CodeService
}

func NewCodeHandler(codeService *service.CodeService) *CodeHandler {
	return &CodeHandler{codeService: codeService}
}

type ExecuteCodeRequest struct {
	Language string `json:"language" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Input    string `json:"input,omitempty"`
}

func (h *CodeHandler) Execute(c *gin.Context) {
	var req ExecuteCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid parameters")
		return
	}

	result, err := h.codeService.Execute(c.Request.Context(), &service.CodeExecutionRequest{
		Language: req.Language,
		Code:     req.Code,
		Input:    req.Input,
	})
	if err != nil {
		response.ErrorWithMessage(c, 4001, err.Error())
		return
	}

	response.SuccessWithData(c, result)
}

func (h *CodeHandler) GetLanguages(c *gin.Context) {
	languages := h.codeService.GetSupportedLanguages()
	response.SuccessWithData(c, gin.H{"languages": languages})
}

type AnalyzeCodeRequest struct {
	Language string `json:"language" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

func (h *CodeHandler) Analyze(c *gin.Context) {
	var req AnalyzeCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid parameters")
		return
	}

	result, err := h.codeService.AnalyzeCode(req.Language, req.Code)
	if err != nil {
		response.ErrorWithMessage(c, 4001, err.Error())
		return
	}

	response.SuccessWithData(c, result)
}

type FormatCodeRequest struct {
	Language string `json:"language" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

func (h *CodeHandler) Format(c *gin.Context) {
	var req FormatCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid parameters")
		return
	}

	formatted, err := h.codeService.FormatCode(req.Language, req.Code)
	if err != nil {
		response.ErrorWithMessage(c, 4001, err.Error())
		return
	}

	response.SuccessWithData(c, gin.H{"formatted_code": formatted})
}

type RunTestsRequest struct {
	Language  string             `json:"language" binding:"required"`
	Code      string             `json:"code" binding:"required"`
	TestCases []service.TestCase `json:"test_cases" binding:"required"`
}

func (h *CodeHandler) RunTests(c *gin.Context) {
	var req RunTestsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid parameters")
		return
	}

	result, err := h.codeService.RunTests(c.Request.Context(), &service.CodeExecutionRequest{
		Language: req.Language,
		Code:     req.Code,
	}, req.TestCases)
	if err != nil {
		response.ErrorWithMessage(c, 4001, err.Error())
		return
	}

	response.SuccessWithData(c, result)
}
