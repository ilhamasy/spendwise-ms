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

func setupTestDB() {
	if config.DB == nil {
		config.InitDB(config.Load())
	}
	if !config.DB.Migrator().HasTable("users") {
		config.AutoMigrate(
			&model.User{}, &model.Transaction{}, &model.Category{}, &model.SavingGoal{}, &model.GoalContribution{}, &model.Budget{},
		)
	}
	config.DB.Where("1 = 1").Delete(&model.GoalContribution{})
	config.DB.Where("1 = 1").Delete(&model.Budget{})
	config.DB.Where("1 = 1").Delete(&model.SavingGoal{})
	config.DB.Where("1 = 1").Delete(&model.Transaction{})
	config.DB.Where("1 = 1").Delete(&model.Category{})
	config.DB.Where("1 = 1").Delete(&model.User{})
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/api/auth/register", Register)
	r.POST("/api/auth/login", Login)
	r.POST("/api/auth/refresh", RefreshToken)
	return r
}

func TestRegisterHandler(t *testing.T) {
	setupTestDB()
	r := setupRouter()

	body, _ := json.Marshal(dto.RegisterRequest{
		Name:     "Test User",
		Email:    "unittest@spendwise.com",
		Password: "password123",
	})

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp dto.AuthResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.AccessToken == "" {
		t.Error("Access token should not be empty")
	}
	if resp.RefreshToken == "" {
		t.Error("Refresh token should not be empty")
	}
	if resp.User.Email != "unittest@spendwise.com" {
		t.Errorf("Expected email 'unittest@spendwise.com', got '%s'", resp.User.Email)
	}
}

func TestRegisterHandler_Duplicate(t *testing.T) {
	setupTestDB()
	r := setupRouter()

	body, _ := json.Marshal(dto.RegisterRequest{
		Name:     "Dup User",
		Email:    "dup@spendwise.com",
		Password: "password123",
	})

	// First registration
	req1, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	// Second registration with same email
	req2, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("Expected status 409 for duplicate, got %d", w2.Code)
	}
}

func TestLoginHandler_Success(t *testing.T) {
	setupTestDB()
	r := setupRouter()

	// First register
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	config.DB.Create(&model.User{
		ID:       uuid.New().String(),
		Name:     "Login Test",
		Email:    "login@spendwise.com",
		Password: string(hash),
	})

	body, _ := json.Marshal(dto.LoginRequest{
		Email:    "login@spendwise.com",
		Password: "password123",
	})

	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestLoginHandler_Invalid(t *testing.T) {
	setupTestDB()
	r := setupRouter()

	body, _ := json.Marshal(dto.LoginRequest{
		Email:    "nonexistent@spendwise.com",
		Password: "wrong",
	})

	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestRefreshHandler(t *testing.T) {
	setupTestDB()
	r := setupRouter()

	// Register first
	regBody, _ := json.Marshal(dto.RegisterRequest{
		Name:     "Refresh Test",
		Email:    "refresh@spendwise.com",
		Password: "password123",
	})
	req1, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(regBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	var regResp dto.AuthResponse
	json.Unmarshal(w1.Body.Bytes(), &regResp)

	// Refresh
	body, _ := json.Marshal(dto.RefreshRequest{RefreshToken: regResp.RefreshToken})
	req2, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var resp dto.AuthResponse
	json.Unmarshal(w2.Body.Bytes(), &resp)
	if resp.AccessToken == "" {
		t.Error("Refreshed access token should not be empty")
	}
}

func TestValidation_EmptyFields(t *testing.T) {
	setupTestDB()
	r := setupRouter()

	body, _ := json.Marshal(dto.RegisterRequest{
		Name:     "",
		Email:    "",
		Password: "",
	})

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for empty fields, got %d", w.Code)
	}
}
