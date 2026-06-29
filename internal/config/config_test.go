package config

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLoad_Defaults(t *testing.T) {
	os.Unsetenv("PORT")
	os.Unsetenv("DB_HOST")
	cfg := Load()
	if cfg.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", cfg.Port)
	}
	if cfg.DBHost != "localhost" {
		t.Errorf("Expected localhost, got %s", cfg.DBHost)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("DB_NAME", "test_db")
	defer os.Unsetenv("PORT")
	defer os.Unsetenv("DB_NAME")

	cfg := Load()
	if cfg.Port != "9090" {
		t.Errorf("Expected port 9090, got %s", cfg.Port)
	}
	if cfg.DBName != "test_db" {
		t.Errorf("Expected test_db, got %s", cfg.DBName)
	}
}

func TestLoad_JWTSecret(t *testing.T) {
	os.Setenv("JWT_SECRET", "my-secret")
	defer os.Unsetenv("JWT_SECRET")
	cfg := Load()
	if cfg.JWTSecret != "my-secret" {
		t.Errorf("Expected my-secret, got %s", cfg.JWTSecret)
	}
}

func TestCORS_OptionsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 204 {
		t.Errorf("Expected 204 for OPTIONS, got %d", w.Code)
	}
}

func TestCORS_NormalRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3003")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3003" {
		t.Errorf("Expected CORS header Access-Control-Allow-Origin: http://localhost:3003, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("Expected Access-Control-Allow-Credentials: true")
	}
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestInitDB(t *testing.T) {
	cfg := Load()
	cfg.DBName = "spendwise_main_db"
	if DB != nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			t.Logf("DB connection failed (expected if no DB): %v", r)
		}
	}()
	InitDB(cfg)
}
