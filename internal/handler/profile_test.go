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

func setupProfileRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	users := r.Group("/api/users/me")
	users.Use(middleware.AuthRequired())
	{
		users.GET("", GetProfile)
		users.PUT("", UpdateProfile)
		users.PUT("/password", ChangePassword)
		users.DELETE("", DeleteAccount)
		users.GET("/export", ExportUserData)
	}
	return r
}

func createUserAndTokenForProfile() (string, string) {
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
	token, _ := createUserAndTokenForProfile()

	req, _ := http.NewRequest("GET", "/api/users/me", nil)
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

	req, _ := http.NewRequest("GET", "/api/users/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestUpdateProfileHandler_Success(t *testing.T) {
	setupProfileTestDB()
	r := setupProfileRouter()
	token, _ := createUserAndTokenForProfile()

	body, _ := json.Marshal(dto.UpdateProfileRequest{Name: "New Name"})
	req, _ := http.NewRequest("PUT", "/api/users/me", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestChangePasswordHandler_Success(t *testing.T) {
	setupProfileTestDB()
	r := setupProfileRouter()
	token, _ := createUserAndTokenForProfile()

	body, _ := json.Marshal(dto.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "newpassword123",
	})
	req, _ := http.NewRequest("PUT", "/api/users/me/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteAccountHandler_Success(t *testing.T) {
	setupProfileTestDB()
	r := setupProfileRouter()
	token, _ := createUserAndTokenForProfile()

	body, _ := json.Marshal(dto.DeleteAccountRequest{Confirmation: "DELETE"})
	req, _ := http.NewRequest("DELETE", "/api/users/me", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestExportHandler_Success(t *testing.T) {
	setupProfileTestDB()
	r := setupProfileRouter()
	token, _ := createUserAndTokenForProfile()

	req, _ := http.NewRequest("GET", "/api/users/me/export", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}
