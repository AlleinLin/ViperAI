package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

type DataResponse struct {
	Response
	Data interface{} `json:"data,omitempty"`
}

const (
	CodeSuccess         = 1000
	CodeInvalidParams   = 2001
	CodeUserExists      = 2002
	CodeUserNotFound    = 2003
	CodeInvalidPassword = 2004
	CodeInvalidToken    = 2005
	CodeNotLoggedIn     = 2006
	CodeInvalidCaptcha  = 2007
	CodeRecordNotFound  = 2008
	CodeForbidden       = 3001
	CodeServerError     = 4001
	CodeAIModelError    = 5001
	CodeTTSError        = 6001
)

var codeMessages = map[int]string{
	CodeSuccess:         "Success",
	CodeInvalidParams:   "Invalid parameters",
	CodeUserExists:      "User already exists",
	CodeUserNotFound:    "User not found",
	CodeInvalidPassword: "Invalid username or password",
	CodeInvalidToken:    "Invalid token",
	CodeNotLoggedIn:     "Not logged in",
	CodeInvalidCaptcha:  "Invalid captcha",
	CodeRecordNotFound:  "Record not found",
	CodeForbidden:       "Permission denied",
	CodeServerError:     "Server error",
	CodeAIModelError:    "AI model error",
	CodeTTSError:        "TTS service error",
}

func getMessage(code int) string {
	if msg, ok := codeMessages[code]; ok {
		return msg
	}
	return "Unknown error"
}

func Success(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: getMessage(CodeSuccess),
	})
}

func SuccessWithData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, DataResponse{
		Response: Response{
			Code:    CodeSuccess,
			Message: getMessage(CodeSuccess),
		},
		Data: data,
	})
}

func Error(c *gin.Context, code int) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: getMessage(code),
	})
}

func ErrorWithMessage(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeInvalidToken,
		Message: message,
	})
}

func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeInvalidParams,
		Message: message,
	})
}

func ServerError(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeServerError,
		Message: message,
	})
}
