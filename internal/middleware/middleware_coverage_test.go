package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimit_Exceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	limiter := NewRateLimiter(2, 1*time.Minute)
	r.Use(rateLimit(limiter))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.100")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if i == 2 && w.Code != http.StatusTooManyRequests {
			t.Errorf("Expected 429 Too Many Requests on 3rd attempt, got %d", w.Code)
		}
	}
}

func TestOptionalAuth_NoHeader_Coverage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(OptionalAuth())
	r.GET("/optional", func(c *gin.Context) {
		userID := c.GetString("userId")
		c.String(http.StatusOK, "user="+userID)
	})

	req := httptest.NewRequest("GET", "/optional", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
	if w.Body.String() != "user=" {
		t.Errorf("Expected empty user ID, got %s", w.Body.String())
	}
}
