package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestLogoutHandler(t *testing.T) {
	setupTestDB()
	r := gin.New()
	r.POST("/logout", Logout)

	req, _ := http.NewRequest("POST", "/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestLoginHandler_ValidationError(t *testing.T) {
	setupTestDB()
	r := gin.New()
	r.POST("/login", Login)

	body, _ := json.Marshal(map[string]string{})
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for empty body, got %d", w.Code)
	}
}

func TestRefreshToken_MissingCookie(t *testing.T) {
	setupTestDB()
	r := gin.New()
	r.POST("/refresh", RefreshToken)

	req, _ := http.NewRequest("POST", "/refresh", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for missing cookie, got %d", w.Code)
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	setupTestDB()
	r := gin.New()
	r.POST("/refresh", RefreshToken)

	req, _ := http.NewRequest("POST", "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "spendwise-refresh-token", Value: "invalid.token.here"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid token, got %d", w.Code)
	}
}

func TestGetTransactionsHandler_Empty(t *testing.T) {
	setupTestDB()
	uid, token := createTestUserForCoverage(t)

	r := gin.New()
	r.GET("/api/transactions", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, GetTransactions)

	req, _ := http.NewRequest("GET", "/api/transactions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestGetCategoriesHandler_Empty(t *testing.T) {
	setupTestDB()
	uid, token := createTestUserForCoverage(t)

	r := gin.New()
	r.GET("/api/categories", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, GetCategories)

	req, _ := http.NewRequest("GET", "/api/categories", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestCreateTransaction_Handler(t *testing.T) {
	setupTestDB()
	uid, token := createTestUserForCoverage(t)
	catID := createTestCategoryForCoverage(t, uid, "Food")

	r := gin.New()
	r.POST("/api/transactions", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, CreateTransaction)

	body, _ := json.Marshal(dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     50000,
		CategoryID: catID,
		OccurredAt: "2026-06-28",
		Note:       "Lunch",
	})
	req, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetGoalsHandler_Empty(t *testing.T) {
	setupTestDB()
	uid, token := createTestUserForCoverage(t)

	r := gin.New()
	r.GET("/api/goals", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, GetGoals)

	req, _ := http.NewRequest("GET", "/api/goals", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func createTestUserForCoverage(t *testing.T) (string, string) {
	t.Helper()
	hash, _ := bcrypt.GenerateFromPassword([]byte("pass123"), 12)
	uid := uuid.New().String()
	email := "cov" + uid[:8] + "@t.com"
	config.DB.Create(&model.User{ID: uid, Name: "CovTest", Email: email, Password: string(hash)})

	r := gin.New()
	r.POST("/login", Login)
	body, _ := json.Marshal(dto.LoginRequest{Email: email, Password: "pass123"})
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	for _, c := range w.Result().Cookies() {
		if c.Name == "spendwise-access-token" {
			return uid, c.Value
		}
	}
	t.Fatal("login failed — no access token cookie")
	return "", ""
}

func createTestCategoryForCoverage(t *testing.T, uid, name string) string {
	t.Helper()
	cat := &model.Category{ID: uuid.New().String(), UserID: uid, Name: name, Type: "expense", IsDefault: true}
	config.DB.Create(cat)
	return cat.ID
}
