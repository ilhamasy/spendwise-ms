package service

import (
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"

	"github.com/google/uuid"
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

func TestUpdateProfile_Theme(t *testing.T) {
	setupProfileTestDB()
	userID, _ := createTestUser()
	svc := NewProfileService(config.DB)

	resp, err := svc.UpdateProfile(userID, dto.UpdateProfileRequest{
		Theme: "dark",
	})
	if err != nil {
		t.Fatalf("UpdateProfile failed: %v", err)
	}
	if resp.Theme != "dark" {
		t.Errorf("Expected theme 'dark', got '%s'", resp.Theme)
	}
}

func TestChangePassword_Success(t *testing.T) {
	setupProfileTestDB()
	userID, _ := createTestUser()
	hash, _ := HashPassword("oldpassword")
	config.DB.Model(&model.User{}).Where("id = ?", userID).Update("password", hash)
	svc := NewProfileService(config.DB)

	err := svc.ChangePassword(userID, "oldpassword", "newpassword123")
	if err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}
}

func TestChangePassword_WrongCurrent(t *testing.T) {
	setupProfileTestDB()
	userID, _ := createTestUser()
	hash, _ := HashPassword("correctpassword")
	config.DB.Model(&model.User{}).Where("id = ?", userID).Update("password", hash)
	svc := NewProfileService(config.DB)

	err := svc.ChangePassword(userID, "wrongpassword", "newpassword")
	if err == nil {
		t.Error("Expected error for wrong current password")
	}
}

func TestDeleteAccount_Success(t *testing.T) {
	setupProfileTestDB()
	userID, _ := createTestUser()
	svc := NewProfileService(config.DB)

	err := svc.DeleteAccount(userID, "DELETE")
	if err != nil {
		t.Fatalf("DeleteAccount failed: %v", err)
	}

	_, err = svc.GetProfile(userID)
	if err == nil {
		t.Error("Expected user to be deleted")
	}
}

func TestDeleteAccount_WrongConfirmation(t *testing.T) {
	setupProfileTestDB()
	userID, _ := createTestUser()
	svc := NewProfileService(config.DB)

	err := svc.DeleteAccount(userID, "delete")
	if err == nil {
		t.Error("Expected error for incorrect confirmation")
	}
}

func TestExportData(t *testing.T) {
	setupProfileTestDB()
	userID, _ := createTestUser()
	svc := NewProfileService(config.DB)

	resp, err := svc.ExportData(userID)
	if err != nil {
		t.Fatalf("ExportData failed: %v", err)
	}
	if resp.Profile.Name != "Profile User" {
		t.Errorf("Expected profile name 'Profile User', got '%s'", resp.Profile.Name)
	}
}
