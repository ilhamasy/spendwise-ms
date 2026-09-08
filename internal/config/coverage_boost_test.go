package config

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestConfig_LoadDefaults(t *testing.T) {
	cfg := Load()
	if cfg.Port != "8080" { t.Errorf("port: %s", cfg.Port) }
	if cfg.DBHost != "localhost" { t.Errorf("host: %s", cfg.DBHost) }
}

func TestConfig_LoadEnvOverrides(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DB_NAME", "testdb")
	cfg := Load()
	if cfg.Port != "9090" { t.Errorf("port: %s", cfg.Port) }
	if cfg.DBName != "testdb" { t.Errorf("db: %s", cfg.DBName) }
}

func TestConfig_CORS_Preflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 204 { t.Errorf("Expected 204, got %d", w.Code) }
}

func TestConfig_CORS_SpendwiseOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://spendwise.vercel.app")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Header().Get("Access-Control-Allow-Origin") != "https://spendwise.vercel.app" {
		t.Errorf("origin: %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("missing credentials")
	}
}

func TestConfig_CORS_UnknownOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("fallback origin: %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestConfig_getEnv_Default(t *testing.T) {
	val := getEnv("NONEXISTENT_KEY", "fallback")
	if val != "fallback" { t.Errorf("expected fallback, got %s", val) }
}

func TestConfig_getEnv_Set(t *testing.T) {
	t.Setenv("MY_KEY", "myvalue")
	val := getEnv("MY_KEY", "default")
	if val != "myvalue" { t.Errorf("expected myvalue, got %s", val) }
}
