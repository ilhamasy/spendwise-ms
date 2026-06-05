package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/middleware"
	"spendwise-ms/internal/model"
	"spendwise-ms/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func setupProfileTestDB() {
	if config.DB != nil {
		config.DB.Where("1 = 1").Delete(&model.User{})
		return
	}
	config.InitDB(config.Load())
	config.AutoMigrate(&model.User{})
	config.DB.Where("1 = 1").Delete(&model.User{})
}

func setupProfileRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	profile := r.Group("/api/profile")
	profile.Use(middleware.AuthRequired())
	{
		profile.GET("", GetProfile)
		profile.PUT("", UpdateProfile)
	}
	return r
}

func createProfileUserAndToken() (string, string) {
	userID := uuid.New().String()
	email := "pftest@spendwise.com"
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	config.DB.Create(&model.User{
		ID: userID, Name: "PF User", Email: email, Password: string(hash),
	})
	token, _, _ := service.GenerateTokens(userID, email)
	return token, userID
}

func TestGetProfileHandler_Success(t *testing.T) {
	setupProfileTestDB()
	r := setupProfileRouter()
	token, _ := createProfileUserAndToken()

	req, _ := http.NewRequest("GET", "/api/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetProfileHandler_Unauthenticated(t *testing.T) {
	setupProfileTestDB()
	r := setupProfileRouter()

	req, _ := http.NewRequest("GET", "/api/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestUpdateProfileHandler_Success(t *testing.T) {
	setupProfileTestDB()
	r := setupProfileRouter()
	token, _ := createProfileUserAndToken()

	body, _ := json.Marshal(dto.UpdateProfileRequest{Name: "New Name"})
	req, _ := http.NewRequest("PUT", "/api/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}
