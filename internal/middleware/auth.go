package middleware

import (
	"net/http"
	"strings"

	"spendwise-ms/internal/service"

	"github.com/gin-gonic/gin"
)

func extractToken(c *gin.Context) (string, error) {
	if token, err := c.Cookie("spendwise-access-token"); err == nil && token != "" {
		return token, nil
	}

	header := c.GetHeader("Authorization")
	if header != "" {
		return strings.TrimPrefix(header, "Bearer "), nil
	}

	return "", http.ErrNoCookie
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := extractToken(c)
		if err != nil || token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Missing authentication token"})
			c.Abort()
			return
		}

		claims, err := service.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Invalid or expired token"})
			c.Abort()
			return
		}

		if claims.TokenType != service.AccessToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Token is not an access token"})
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
		token, err := extractToken(c)
		if err != nil || token == "" {
			c.Next()
			return
		}

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
