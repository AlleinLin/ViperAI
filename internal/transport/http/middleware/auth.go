package middleware

import (
	"strings"

	"viperai/internal/pkg/auth"
	"viperai/internal/transport/http/response"

	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			token = c.Query("token")
		}

		if token == "" {
			response.Unauthorized(c, "Token is required")
			c.Abort()
			return
		}

		claims, ok := auth.ParseToken(token)
		if !ok {
			response.Unauthorized(c, "Invalid token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("account", claims.Account)
		c.Next()
	}
}

func GetUserID(c *gin.Context) int64 {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return userID.(int64)
}

func GetAccount(c *gin.Context) string {
	account, exists := c.Get("account")
	if !exists {
		return ""
	}
	return account.(string)
}
