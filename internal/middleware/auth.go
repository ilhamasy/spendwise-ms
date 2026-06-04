package middleware

import (
	"net/http"
	"strings"

	"spendwise-ms/internal/service"

	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Missing authorization header"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := service.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("userId", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Next()
	}
}

func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.Next()
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := service.ValidateToken(token)
		if err != nil {
			c.Next()
			return
		}

		c.Set("userId", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Next()
	}
}
