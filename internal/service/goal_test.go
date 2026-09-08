package service

import (
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"
	"spendwise-ms/internal/repository"

	"github.com/google/uuid"
)

func setupGoalTestDB() {
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

func createUserForGoal() string {
	userID := uuid.New().String()
	config.DB.Create(&model.User{
		ID:       userID,
		Name:     "Goal Test User",
		Email:    "goaltest@spendwise.com",
		Password: "hashed",
	})
	return userID
}

func TestCreateGoal_Success(t *testing.T) {
	setupGoalTestDB()
	userID := createUserForGoal()
	svc := NewGoalService(config.DB)

	resp, err := svc.CreateGoal(userID, dto.CreateGoalRequest{
		Name:         "New Car",
		TargetAmount: 100000000,
		CurrentSaved: 10000000,
		TargetDate:   "2026-12-31",
	})
	if err != nil {
		t.Fatalf("CreateGoal failed: %v", err)
	}
	if resp.Name != "New Car" {
		t.Errorf("Expected name 'New Car', got '%s'", resp.Name)
	}
	if resp.TargetAmount != 100000000 {
		t.Errorf("Expected targetAmount 100000000, got %d", resp.TargetAmount)
	}
	if resp.CurrentSaved != 10000000 {
		t.Errorf("Expected currentSaved 10000000, got %d", resp.CurrentSaved)
	}
	if resp.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", resp.Status)
	}
	if resp.Progress != 10 {
		t.Errorf("Expected progress 10%%, got %f", resp.Progress)
	}
}

func TestCreateGoal_ZeroTargetAmount(t *testing.T) {
	setupGoalTestDB()
	userID := createUserForGoal()
	svc := NewGoalService(config.DB)

	_, err := svc.CreateGoal(userID, dto.CreateGoalRequest{
		Name:         "Invalid",
		TargetAmount: 0,
	})
	if err == nil {
		t.Error("Expected validation error for zero targetAmount")
	}
}

func TestListGoals_FilterByStatus(t *testing.T) {
	setupGoalTestDB()
	userID := createUserForGoal()
	svc := NewGoalService(config.DB)

	g1, _ := svc.CreateGoal(userID, dto.CreateGoalRequest{Name: "Active Goal", TargetAmount: 50000})
	g2, _ := svc.CreateGoal(userID, dto.CreateGoalRequest{Name: "Another Active", TargetAmount: 30000})

	svc.ArchiveGoal(userID, g2.ID)

	activeResp, err := svc.ListGoals(userID, "active")
	if err != nil {
		t.Fatalf("ListGoals(active) failed: %v", err)
	}
	if len(activeResp) != 1 {
		t.Errorf("Expected 1 active goal, got %d", len(activeResp))
	}
	if activeResp[0].ID != g1.ID {
		t.Errorf("Expected goal '%s', got '%s'", g1.ID, activeResp[0].ID)
	}

	archivedResp, err := svc.ListGoals(userID, "archived")
	if err != nil {
		t.Fatalf("ListGoals(archived) failed: %v", err)
	}
	if len(archivedResp) != 1 {
		t.Errorf("Expected 1 archived goal, got %d", len(archivedResp))
	}
}

func TestArchiveUnarchiveGoal(t *testing.T) {
	setupGoalTestDB()
	userID := createUserForGoal()
	svc := NewGoalService(config.DB)

	created, _ := svc.CreateGoal(userID, dto.CreateGoalRequest{Name: "To Archive", TargetAmount: 50000})

	archived, err := svc.ArchiveGoal(userID, created.ID)
	if err != nil {
		t.Fatalf("ArchiveGoal failed: %v", err)
	}
	if archived.Status != "archived" {
		t.Errorf("Expected status 'archived', got '%s'", archived.Status)
	}

	unarchived, err := svc.UnarchiveGoal(userID, created.ID)
	if err != nil {
		t.Fatalf("UnarchiveGoal failed: %v", err)
	}
	if unarchived.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", unarchived.Status)
	}
}

func TestDeleteGoal(t *testing.T) {
	setupGoalTestDB()
	userID := createUserForGoal()
	svc := NewGoalService(config.DB)

	created, _ := svc.CreateGoal(userID, dto.CreateGoalRequest{Name: "To Delete", TargetAmount: 50000})

	err := svc.DeleteGoal(userID, created.ID)
	if err != nil {
		t.Fatalf("DeleteGoal failed: %v", err)
	}

	_, err = svc.GetGoal(userID, created.ID)
	if err == nil {
		t.Error("Expected error when retrieving deleted goal")
	}
}

func TestAddContribution(t *testing.T) {
	setupGoalTestDB()
	userID := createUserForGoal()
	svc := NewGoalService(config.DB)

	goal, _ := svc.CreateGoal(userID, dto.CreateGoalRequest{Name: "Savings", TargetAmount: 100000})

	_, updatedGoal, err := svc.AddContribution(userID, goal.ID, dto.CreateContributionRequest{
		Amount: 25000,
		Note:   "Monthly deposit",
	})
	if err != nil {
		t.Fatalf("AddContribution failed: %v", err)
	}
	if updatedGoal.CurrentSaved != 25000 {
		t.Errorf("Expected currentSaved 25000, got %d", updatedGoal.CurrentSaved)
	}
	if updatedGoal.Progress != 25 {
		t.Errorf("Expected progress 25%%, got %f", updatedGoal.Progress)
	}
}

func TestGetContributions(t *testing.T) {
	setupGoalTestDB()
	userID := createUserForGoal()
	svc := NewGoalService(config.DB)

	goal, _ := svc.CreateGoal(userID, dto.CreateGoalRequest{Name: "Savings", TargetAmount: 100000})

	svc.AddContribution(userID, goal.ID, dto.CreateContributionRequest{Amount: 10000, Note: "First"})
	svc.AddContribution(userID, goal.ID, dto.CreateContributionRequest{Amount: 20000, Note: "Second"})
	svc.AddContribution(userID, goal.ID, dto.CreateContributionRequest{Amount: 5000, Note: "Third"})

	resp, err := svc.GetContributions(userID, goal.ID, 1, 2)
	if err != nil {
		t.Fatalf("GetContributions failed: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Errorf("Expected 2 contributions, got %d", len(resp.Data))
	}
	if resp.Meta.Total != 3 {
		t.Errorf("Expected total 3, got %d", resp.Meta.Total)
	}
}

func TestCalculateGoalProgress(t *testing.T) {
	if p := repository.CalculateGoalProgress(50, 100); p != 50 {
		t.Errorf("Expected 50%%, got %f", p)
	}
	if p := repository.CalculateGoalProgress(75, 100); p != 75 {
		t.Errorf("Expected 75%%, got %f", p)
	}
	if p := repository.CalculateGoalProgress(100, 100); p != 100 {
		t.Errorf("Expected 100%%, got %f", p)
	}
	if p := repository.CalculateGoalProgress(0, 100); p != 0 {
		t.Errorf("Expected 0%%, got %f", p)
	}
}
