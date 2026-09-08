package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/service"

	"github.com/gin-gonic/gin"
)

func TestAuthFailures_WeakPasswordRejection(t *testing.T) {
	err := service.ValidatePasswordStrength("12345678")
	if err == nil {
		t.Error("Expected error for weak password '12345678', got nil")
	}

	err = service.ValidatePasswordStrength("password")
	if err == nil {
		t.Error("Expected error for weak password 'password', got nil")
	}

	err = service.ValidatePasswordStrength("Short1")
	if err == nil {
		t.Error("Expected error for password shorter than 8 characters, got nil")
	}

	err = service.ValidatePasswordStrength("StrongAuthPassword123!")
	if err != nil {
		t.Errorf("Unexpected error for strong password: %v", err)
	}
}

func TestAuthFailures_GenericLoginErrorMessage(t *testing.T) {
	setupTestDB()

	// 1. Login with non-existent email
	_, errNonExistent := service.Login(config.DB, "nonexistent@test.com", "anyPassword123!")
	if errNonExistent == nil {
		t.Fatal("Expected login error for non-existent user")
	}

	// 2. Login with registered email but wrong password
	service.Register(config.DB, "User Fail", "userfail@test.com", "CorrectPassword123!")
	_, errWrongPass := service.Login(config.DB, "userfail@test.com", "WrongPassword123!")
	if errWrongPass == nil {
		t.Fatal("Expected login error for wrong password")
	}

	// Both failure messages must be identical ("invalid email or password") to prevent user enumeration
	if errNonExistent.Error() != errWrongPass.Error() {
		t.Errorf("Mismatch in login failure messages: non-existent='%s', wrong-pass='%s'", errNonExistent.Error(), errWrongPass.Error())
	}
	if errNonExistent.Error() != "invalid email or password" {
		t.Errorf("Expected generic error 'invalid email or password', got '%s'", errNonExistent.Error())
	}
}

func TestAuthFailures_RefreshTokenTokenTypeValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/auth/refresh", RefreshToken)

	// Generate an Access token and attempt to use it as Refresh token
	accessToken, _, err := service.GenerateTokens("user-101", "token@test.com")
	if err != nil {
		t.Fatalf("GenerateTokens failed: %v", err)
	}

	body, _ := json.Marshal(dto.RefreshRequest{RefreshToken: accessToken})
	req, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized when using Access token as Refresh token, got %d", w.Code)
	}
}
