package service

import (
	"testing"
	"time"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"
	"spendwise-ms/internal/repository"

	"github.com/google/uuid"
)

func setupBudgetTestDB() {
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

func createUserAndCatForBudget() (string, string) {
	userID := uuid.New().String()
	config.DB.Create(&model.User{
		ID:       userID,
		Name:     "Budget Test User",
		Email:    "budgettest@spendwise.com",
		Password: "hashed",
	})
	catID := uuid.New().String()
	config.DB.Create(&model.Category{
		ID:        catID,
		UserID:    userID,
		Name:      "Food",
		Type:      "expense",
		IsDefault: false,
	})
	return userID, catID
}

func TestCreateBudget_Success(t *testing.T) {
	setupBudgetTestDB()
	userID, catID := createUserAndCatForBudget()
	svc := NewBudgetService(config.DB)

	resp, err := svc.CreateBudget(userID, dto.CreateBudgetRequest{
		Name:       "Monthly Food",
		Amount:     2000000,
		Period:     "monthly",
		CategoryID: catID,
	})
	if err != nil {
		t.Fatalf("CreateBudget failed: %v", err)
	}
	if resp.Name != "Monthly Food" {
		t.Errorf("Expected name 'Monthly Food', got '%s'", resp.Name)
	}
	if resp.Amount != 2000000 {
		t.Errorf("Expected amount 2000000, got %d", resp.Amount)
	}
	if resp.Period != "monthly" {
		t.Errorf("Expected period 'monthly', got '%s'", resp.Period)
	}
	if resp.Spent != 0 {
		t.Errorf("Expected spent 0, got %d", resp.Spent)
	}
	if resp.Remaining != 2000000 {
		t.Errorf("Expected remaining 2000000, got %d", resp.Remaining)
	}
}

func TestCreateBudget_InvalidPeriod(t *testing.T) {
	setupBudgetTestDB()
	userID, catID := createUserAndCatForBudget()
	svc := NewBudgetService(config.DB)

	_, err := svc.CreateBudget(userID, dto.CreateBudgetRequest{
		Name:       "Bad Budget",
		Amount:     1000,
		Period:     "yearlyyy",
		CategoryID: catID,
	})
	if err == nil {
		t.Error("Expected validation error for invalid period")
	}
}

func TestCreateBudget_ZeroAmount(t *testing.T) {
	setupBudgetTestDB()
	userID, catID := createUserAndCatForBudget()
	svc := NewBudgetService(config.DB)

	_, err := svc.CreateBudget(userID, dto.CreateBudgetRequest{
		Name:       "Zero Budget",
		Amount:     0,
		Period:     "monthly",
		CategoryID: catID,
	})
	if err == nil {
		t.Error("Expected validation error for zero amount")
	}
}

func TestCreateBudget_InvalidCategory(t *testing.T) {
	setupBudgetTestDB()
	userID, _ := createUserAndCatForBudget()
	svc := NewBudgetService(config.DB)

	_, err := svc.CreateBudget(userID, dto.CreateBudgetRequest{
		Name:       "Bad Cat",
		Amount:     1000,
		Period:     "monthly",
		CategoryID: "non-existent",
	})
	if err == nil {
		t.Error("Expected error for invalid category")
	}
}

func TestListBudgets(t *testing.T) {
	setupBudgetTestDB()
	userID, catID := createUserAndCatForBudget()
	svc := NewBudgetService(config.DB)

	svc.CreateBudget(userID, dto.CreateBudgetRequest{
		Name: "Food Budget", Amount: 1000000, Period: "monthly", CategoryID: catID,
	})
	svc.CreateBudget(userID, dto.CreateBudgetRequest{
		Name: "Transport Budget", Amount: 500000, Period: "weekly", CategoryID: catID,
	})

	resp, err := svc.ListBudgets(userID, "")
	if err != nil {
		t.Fatalf("ListBudgets failed: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("Expected 2 budgets, got %d", len(resp))
	}
}

func TestListBudgets_FilterByPeriod(t *testing.T) {
	setupBudgetTestDB()
	userID, catID := createUserAndCatForBudget()
	svc := NewBudgetService(config.DB)

	svc.CreateBudget(userID, dto.CreateBudgetRequest{
		Name: "Monthly", Amount: 500000, Period: "monthly", CategoryID: catID,
	})
	svc.CreateBudget(userID, dto.CreateBudgetRequest{
		Name: "Weekly", Amount: 100000, Period: "weekly", CategoryID: catID,
	})

	resp, err := svc.ListBudgets(userID, "weekly")
	if err != nil {
		t.Fatalf("ListBudgets(weekly) failed: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("Expected 1 weekly budget, got %d", len(resp))
	}
}

func TestUpdateBudget(t *testing.T) {
	setupBudgetTestDB()
	userID, catID := createUserAndCatForBudget()
	svc := NewBudgetService(config.DB)

	created, _ := svc.CreateBudget(userID, dto.CreateBudgetRequest{
		Name: "Old", Amount: 100000, Period: "monthly", CategoryID: catID,
	})

	resp, err := svc.UpdateBudget(userID, created.ID, dto.UpdateBudgetRequest{
		Name:   "Updated",
		Amount: 250000,
	})
	if err != nil {
		t.Fatalf("UpdateBudget failed: %v", err)
	}
	if resp.Name != "Updated" {
		t.Errorf("Expected name 'Updated', got '%s'", resp.Name)
	}
	if resp.Amount != 250000 {
		t.Errorf("Expected amount 250000, got %d", resp.Amount)
	}
}

func TestDeleteBudget(t *testing.T) {
	setupBudgetTestDB()
	userID, catID := createUserAndCatForBudget()
	svc := NewBudgetService(config.DB)

	created, _ := svc.CreateBudget(userID, dto.CreateBudgetRequest{
		Name: "DeleteMe", Amount: 10000, Period: "daily", CategoryID: catID,
	})

	err := svc.DeleteBudget(userID, created.ID)
	if err != nil {
		t.Fatalf("DeleteBudget failed: %v", err)
	}

	_, err = svc.GetBudget(userID, created.ID)
	if err == nil {
		t.Error("Expected error when retrieving deleted budget")
	}
}

func TestBudgetSpentCalculation(t *testing.T) {
	setupBudgetTestDB()
	userID, catID := createUserAndCatForBudget()
	svc := NewBudgetService(config.DB)
	txnRepo := repository.NewTransactionRepository(config.DB)

	today := time.Now().Format("2006-01-02")
	txnRepo.Create(&model.Transaction{
		UserID: userID, Type: "expense", Amount: 30000,
		CategoryID: catID, OccurredAt: today,
	})
	txnRepo.Create(&model.Transaction{
		UserID: userID, Type: "expense", Amount: 20000,
		CategoryID: catID, OccurredAt: today,
	})

	resp, err := svc.CreateBudget(userID, dto.CreateBudgetRequest{
		Name: "Tracked", Amount: 100000, Period: "monthly", CategoryID: catID,
	})
	if err != nil {
		t.Fatalf("CreateBudget failed: %v", err)
	}
	if resp.Spent != 50000 {
		t.Errorf("Expected spent 50000, got %d", resp.Spent)
	}
	if resp.Remaining != 50000 {
		t.Errorf("Expected remaining 50000, got %d", resp.Remaining)
	}
}
