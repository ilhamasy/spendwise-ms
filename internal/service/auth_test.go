package service

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("testpassword")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "testpassword" {
		t.Error("Password should not be stored in plaintext")
	}
	if len(hash) == 0 {
		t.Error("Hash should not be empty")
	}
}

func TestCheckPassword(t *testing.T) {
	hash, _ := HashPassword("mypassword")

	if !CheckPassword("mypassword", hash) {
		t.Error("CheckPassword should return true for correct password")
	}
	if CheckPassword("wrongpassword", hash) {
		t.Error("CheckPassword should return false for wrong password")
	}
}

func TestHashPasswordUniqueness(t *testing.T) {
	h1, _ := HashPassword("samepassword")
	h2, _ := HashPassword("samepassword")
	if h1 == h2 {
		t.Error("Same password should produce different hashes (bcrypt salt)")
	}
}

func TestGenerateTokens(t *testing.T) {
	access, refresh, err := GenerateTokens("user-1", "test@test.com")
	if err != nil {
		t.Fatalf("GenerateTokens failed: %v", err)
	}
	if access == "" || refresh == "" {
		t.Error("Tokens should not be empty")
	}
	if access == refresh {
		t.Error("Access and refresh tokens should be different")
	}
}

func TestValidateToken(t *testing.T) {
	access, _, err := GenerateTokens("user-1", "test@test.com")
	if err != nil {
		t.Fatalf("GenerateTokens failed: %v", err)
	}

	claims, err := ValidateToken(access)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.UserID != "user-1" {
		t.Errorf("Expected userID 'user-1', got '%s'", claims.UserID)
	}
	if claims.Email != "test@test.com" {
		t.Errorf("Expected email 'test@test.com', got '%s'", claims.Email)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	_, err := ValidateToken("invalid.token.here")
	if err == nil {
		t.Error("ValidateToken should fail for invalid token")
	}
}
