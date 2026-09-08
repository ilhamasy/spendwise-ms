package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"spendwise-ms/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func TestCryptographicFailures_CookieSecurityFlags(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/test-cookies", func(c *gin.Context) {
		setTokenCookies(c, "access.jwt.token", "refresh.jwt.token")
		c.Status(http.StatusOK)
	})

	r.GET("/test-clear-cookies", func(c *gin.Context) {
		clearTokenCookies(c)
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test-cookies", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", w.Code)
	}

	cookies := w.Result().Cookies()
	var accessCookie, refreshCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "spendwise-access-token" {
			accessCookie = c
		}
		if c.Name == "spendwise-refresh-token" {
			refreshCookie = c
		}
	}

	if accessCookie == nil || refreshCookie == nil {
		t.Fatal("Expected both access and refresh cookies to be set")
	}

	if !accessCookie.HttpOnly || !refreshCookie.HttpOnly {
		t.Error("Cookies must have HttpOnly set to true to prevent XSS token theft")
	}

	if accessCookie.SameSite != http.SameSiteLaxMode || refreshCookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("Cookies must have SameSite=Lax mode, got access=%v, refresh=%v", accessCookie.SameSite, refreshCookie.SameSite)
	}
}

func TestCryptographicFailures_ReleaseModeSecureCookie(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	defer gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/test-release-cookies", func(c *gin.Context) {
		setTokenCookies(c, "access.jwt.token", "refresh.jwt.token")
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test-release-cookies", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "spendwise-access-token" {
			if !c.Secure {
				t.Error("Cookie Secure flag must be true in Release mode")
			}
		}
	}
}

func TestCryptographicFailures_BcryptCostFactor(t *testing.T) {
	hash, err := service.HashPassword("securePassword123!")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		t.Fatalf("Failed to extract bcrypt cost: %v", err)
	}

	if cost < 12 {
		t.Errorf("Bcrypt cost factor must be at least 12 for strong password hashing, got %d", cost)
	}
}
