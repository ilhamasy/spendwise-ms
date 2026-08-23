package middleware

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

type SecurityLogEntry struct {
	Timestamp  string `json:"timestamp"`
	ClientIP   string `json:"clientIp"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	StatusCode int    `json:"statusCode"`
	LatencyMs  int64  `json:"latencyMs"`
	UserID     string `json:"userId,omitempty"`
	EventType  string `json:"eventType,omitempty"`
	Message    string `json:"message,omitempty"`
}

func SecurityLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		status := c.Writer.Status()
		// Log security-sensitive HTTP events (errors, auth failures, rate limits)
		if status >= 400 {
			latency := time.Since(start).Milliseconds()
			userID := c.GetString("userId")

			eventType := "HTTP_ERROR"
			if status == 401 {
				eventType = "AUTH_FAILURE"
			} else if status == 403 {
				eventType = "ACCESS_DENIED"
			} else if status == 429 {
				eventType = "RATE_LIMIT_EXCEEDED"
			}

			entry := SecurityLogEntry{
				Timestamp:  time.Now().Format(time.RFC3339),
				ClientIP:   c.ClientIP(),
				Method:     c.Request.Method,
				Path:       c.Request.URL.Path,
				StatusCode: status,
				LatencyMs:  latency,
				UserID:     userID,
				EventType:  eventType,
			}

			data, err := json.Marshal(entry)
			if err == nil {
				log.Println(string(data))
			}
		}
	}
}

func LogSecurityEvent(c *gin.Context, eventType, message string) {
	entry := SecurityLogEntry{
		Timestamp:  time.Now().Format(time.RFC3339),
		ClientIP:   c.ClientIP(),
		Method:     c.Request.Method,
		Path:       c.Request.URL.Path,
		StatusCode: c.Writer.Status(),
		UserID:     c.GetString("userId"),
		EventType:  eventType,
		Message:    message,
	}

	data, err := json.Marshal(entry)
	if err == nil {
		log.Println(string(data))
	}
}
