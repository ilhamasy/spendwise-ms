package handler

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"spendwise-ms/internal/middleware"

	"github.com/gin-gonic/gin"
)

func TestSecurityLogging_LogsUnauthorizedEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.SecurityLogger())
	r.GET("/api/test-auth", func(c *gin.Context) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	})

	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(nil)

	req, _ := http.NewRequest("GET", "/api/test-auth", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "AUTH_FAILURE") {
		t.Errorf("Expected log to contain 'AUTH_FAILURE', got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "/api/test-auth") {
		t.Errorf("Expected log to contain path '/api/test-auth', got: %s", logOutput)
	}
}

func TestSecurityLogging_ExplicitEventLogging(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/test-event", func(c *gin.Context) {
		middleware.LogSecurityEvent(c, "SUSPICIOUS_ACTIVITY", "Multiple failed attempts")
		c.Status(http.StatusOK)
	})

	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(nil)

	req, _ := http.NewRequest("POST", "/api/test-event", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "SUSPICIOUS_ACTIVITY") {
		t.Errorf("Expected log to contain 'SUSPICIOUS_ACTIVITY', got: %s", logOutput)
	}
}
