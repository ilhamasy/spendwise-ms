package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/middleware"
	"spendwise-ms/internal/model"
	"spendwise-ms/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestInsecureDesign_NumericBoundValidation(t *testing.T) {
	setupTestDB()
	userID := uuid.New().String()
	config.DB.Create(&model.User{ID: userID, Name: "Bound Test", Email: "bound@test.com", Password: "hash"})

	catSvc := service.NewCategoryService(config.DB)
	cat, _ := catSvc.CreateCategory(userID, dto.CreateCategoryRequest{Name: "General", Type: "expense"})

	txnSvc := service.NewTransactionService(config.DB)

	// Attempt transaction amount exceeding 1 Trillion IDR
	hugeAmount := int64(1000000000001)
	_, err := txnSvc.CreateTransaction(userID, dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     hugeAmount,
		CategoryID: cat.ID,
		OccurredAt: "2026-08-23",
	})
	if err == nil {
		t.Error("Expected error for transaction amount exceeding max bound, got nil")
	}

	goalSvc := service.NewGoalService(config.DB)
	_, err = goalSvc.CreateGoal(userID, dto.CreateGoalRequest{
		Name:         "Overbound Goal",
		TargetAmount: hugeAmount,
	})
	if err == nil {
		t.Error("Expected error for goal target amount exceeding max bound, got nil")
	}
}

func TestInsecureDesign_PastTargetDateRejection(t *testing.T) {
	setupTestDB()
	userID := uuid.New().String()
	config.DB.Create(&model.User{ID: userID, Name: "Past Date Test", Email: "pastdate@test.com", Password: "hash"})

	goalSvc := service.NewGoalService(config.DB)
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	_, err := goalSvc.CreateGoal(userID, dto.CreateGoalRequest{
		Name:         "Past Goal",
		TargetAmount: 1000000,
		TargetDate:   yesterday,
	})
	if err == nil {
		t.Error("Expected error when creating a goal with target date in the past, got nil")
	}
}

func TestInsecureDesign_AuthBruteForceRateLimit(t *testing.T) {
	middleware.ResetLimiters()
	defer middleware.ResetLimiters()

	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/api/auth/login", middleware.RateLimitAuth(), func(c *gin.Context) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
	})

	body, _ := json.Marshal(map[string]string{"email": "brute@test.com", "password": "wrong"})

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", "203.0.113.195")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("Iteration %d expected 401 Unauthorized, got %d", i+1, w.Code)
		}
	}

	// 11th request must be rate limited (429 Too Many Requests)
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "203.0.113.195")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected 429 Too Many Requests on 11th attempt, got %d", w.Code)
	}

	if w.Header().Get("Retry-After") == "" {
		t.Error("Rate-limited response missing Retry-After header")
	}
}
