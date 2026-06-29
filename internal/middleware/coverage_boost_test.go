package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMW_AuthRequiredMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", AuthRequired(), func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	req, _ := http.NewRequest("GET", "/x", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized { t.Errorf("expected 401, got %d", w.Code) }
}

func TestMW_AuthRequiredBadToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", AuthRequired(), func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	req, _ := http.NewRequest("GET", "/x", nil)
	req.Header.Set("Authorization", "Bearer bad.jwt.here")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized { t.Errorf("expected 401, got %d", w.Code) }
}

func TestMW_OptionalAuthNoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", OptionalAuth(), func(c *gin.Context) { c.JSON(200, gin.H{"userId": c.GetString("userId")}) })
	req, _ := http.NewRequest("GET", "/x", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK { t.Errorf("expected 200, got %d", w.Code) }
}

func TestMW_RateLimitGlobal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", RateLimitGlobal(), func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	req, _ := http.NewRequest("GET", "/x", nil)
	req.RemoteAddr = "1.1.1.1:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK { t.Errorf("expected 200, got %d", w.Code) }
}

func TestMW_RateLimitAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", RateLimitAuth(), func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	req, _ := http.NewRequest("GET", "/x", nil)
	req.RemoteAddr = "2.2.2.2:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK { t.Errorf("expected 200, got %d", w.Code) }
}
