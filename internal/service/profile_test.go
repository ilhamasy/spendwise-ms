package service

import (
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"

	"github.com/google/uuid"
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

func createTestUser() (string, string) {
	userID := uuid.New().String()
	email := "profiletest@spendwise.com"
	config.DB.Create(&model.User{
		ID: userID, Name: "Profile User", Email: email, Password: "hash",
	})
	return userID, email
}

func TestGetProfile_Success(t *testing.T) {
	setupProfileTestDB()
	userID, email := createTestUser()
	svc := NewProfileService(config.DB)

	resp, err := svc.GetProfile(userID)
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}
	if resp.Name != "Profile User" {
		t.Errorf("Expected name 'Profile User', got '%s'", resp.Name)
	}
	if resp.Email != email {
		t.Errorf("Expected email '%s', got '%s'", email, resp.Email)
	}
	if resp.CreatedAt == "" {
		t.Error("CreatedAt should not be empty")
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	setupProfileTestDB()
	svc := NewProfileService(config.DB)

	_, err := svc.GetProfile("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

func TestUpdateProfile_Name(t *testing.T) {
	setupProfileTestDB()
	userID, _ := createTestUser()
	svc := NewProfileService(config.DB)

	resp, err := svc.UpdateProfile(userID, dto.UpdateProfileRequest{
		Name: "Updated Name",
	})
	if err != nil {
		t.Fatalf("UpdateProfile failed: %v", err)
	}
	if resp.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", resp.Name)
	}
}

func TestUpdateProfile_EmailAlreadyTaken(t *testing.T) {
	setupProfileTestDB()
	userID, _ := createTestUser()
	otherID := uuid.New().String()
	config.DB.Create(&model.User{
		ID: otherID, Name: "Other", Email: "taken@test.com", Password: "hash",
	})
	svc := NewProfileService(config.DB)

	_, err := svc.UpdateProfile(userID, dto.UpdateProfileRequest{
		Email: "taken@test.com",
	})
	if err == nil {
		t.Error("Expected error for duplicate email")
	}
}
