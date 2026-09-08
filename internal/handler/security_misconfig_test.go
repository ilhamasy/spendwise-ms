package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"spendwise-ms/internal/middleware"

	"github.com/gin-gonic/gin"
)

func TestSecurityMisconfiguration_HTTPHeadersPresence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.SecurityHeaders())

	r.GET("/health-test", HealthCheck)

	req, _ := http.NewRequest("GET", "/health-test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", w.Code)
	}

	headers := w.Header()

	if headers.Get("X-Content-Type-Options") != "nosniff" {
		t.Error("Missing or invalid X-Content-Type-Options security header")
	}

	if headers.Get("X-Frame-Options") != "DENY" {
		t.Error("Missing or invalid X-Frame-Options security header")
	}

	if headers.Get("X-XSS-Protection") != "1; mode=block" {
		t.Error("Missing or invalid X-XSS-Protection security header")
	}

	if headers.Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Error("Missing or invalid Referrer-Policy security header")
	}

	if headers.Get("Content-Security-Policy") == "" {
		t.Error("Missing Content-Security-Policy header")
	}

	if headers.Get("Permissions-Policy") == "" {
		t.Error("Missing Permissions-Policy header")
	}
}

func TestSecurityMisconfiguration_HSTSInReleaseMode(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	defer gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(middleware.SecurityHeaders())
	r.GET("/test-hsts", HealthCheck)

	req, _ := http.NewRequest("GET", "/test-hsts", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Strict-Transport-Security") == "" {
		t.Error("Strict-Transport-Security header must be present in Release mode")
	}
}
