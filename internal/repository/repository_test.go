package repository

import (
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/model"

	"github.com/google/uuid"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	cfg := config.Load()
	if config.DB == nil {
		config.InitDB(cfg)
		config.AutoMigrate(
			&model.User{},
			&model.Transaction{},
			&model.Category{},
			&model.SavingGoal{},
			&model.GoalContribution{},
			&model.Budget{},
		)
	}
	config.DB.Where("1 = 1").Delete(&model.Transaction{})
	config.DB.Where("1 = 1").Delete(&model.Category{})
	config.DB.Where("1 = 1").Delete(&model.Budget{})
	config.DB.Where("1 = 1").Delete(&model.GoalContribution{})
	config.DB.Where("1 = 1").Delete(&model.SavingGoal{})
	config.DB.Where("1 = 1").Delete(&model.User{})
}

func TestTransactionRepository_Create(t *testing.T) {
	setupTestDB(t)
	repo := NewTransactionRepository(config.DB)

	tx := &model.Transaction{
		ID:         uuid.New().String(),
		UserID:     "user-1",
		Type:       "expense",
		Amount:     50000,
		CategoryID: "cat-1",
		OccurredAt: "2026-06-01",
	}
	if err := repo.Create(tx); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(tx.ID, "user-1")
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Amount != 50000 {
		t.Errorf("Expected 50000, got %d", found.Amount)
	}
}

func TestTransactionRepository_FindAll(t *testing.T) {
	setupTestDB(t)
	repo := NewTransactionRepository(config.DB)

	tx1 := &model.Transaction{ID: uuid.New().String(), UserID: "u1", Type: "expense", Amount: 100, CategoryID: "c1", OccurredAt: "2026-01-01"}
	tx2 := &model.Transaction{ID: uuid.New().String(), UserID: "u1", Type: "income", Amount: 200, CategoryID: "c1", OccurredAt: "2026-06-01"}
	repo.Create(tx1)
	repo.Create(tx2)

	txs, _, err := repo.FindAll(TransactionFilter{UserID: "u1", Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(txs) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(txs))
	}
}

func TestTransactionRepository_Update(t *testing.T) {
	setupTestDB(t)
	repo := NewTransactionRepository(config.DB)

	tx := &model.Transaction{ID: uuid.New().String(), UserID: "u1", Type: "expense", Amount: 100, CategoryID: "c1", OccurredAt: "2026-01-01"}
	repo.Create(tx)

	tx.Amount = 500
	if err := repo.Update(tx); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.FindByID(tx.ID, "u1")
	if found.Amount != 500 {
		t.Errorf("Expected 500, got %d", found.Amount)
	}
}

func TestTransactionRepository_Delete(t *testing.T) {
	setupTestDB(t)
	repo := NewTransactionRepository(config.DB)

	tx := &model.Transaction{ID: uuid.New().String(), UserID: "u1", Type: "expense", Amount: 100, CategoryID: "c1", OccurredAt: "2026-01-01"}
	repo.Create(tx)

	if err := repo.Delete(tx.ID, "u1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := repo.FindByID(tx.ID, "u1")
	if err == nil {
		t.Error("Expected error after delete, got nil")
	}
}

func TestCategoryRepository_CreateAndFind(t *testing.T) {
	setupTestDB(t)
	repo := NewCategoryRepository(config.DB)

	cat := &model.Category{
		ID:   uuid.New().String(),
		Name: "TestCat",
		Type: "expense",
		Icon: "📁",
		UserID: "u1",
	}
	if err := repo.Create(cat); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(cat.ID, "u1")
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Name != "TestCat" {
		t.Errorf("Expected 'TestCat', got '%s'", found.Name)
	}
}

func TestCategoryRepository_FindAll(t *testing.T) {
	setupTestDB(t)
	repo := NewCategoryRepository(config.DB)

	repo.Create(&model.Category{ID: uuid.New().String(), Name: "Food", Type: "expense", Icon: "🍔", UserID: "u1"})
	repo.Create(&model.Category{ID: uuid.New().String(), Name: "Salary", Type: "income", Icon: "💰", UserID: "u1"})

	cats, err := repo.FindAll("u1", "")
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(cats) < 2 {
		t.Errorf("Expected at least 2 categories, got %d", len(cats))
	}

	expenses, _ := repo.FindAll("u1", "expense")
	if len(expenses) < 1 {
		t.Errorf("Expected at least 1 expense category")
	}
}

func TestCategoryRepository_Update(t *testing.T) {
	setupTestDB(t)
	repo := NewCategoryRepository(config.DB)

	cat := &model.Category{ID: uuid.New().String(), Name: "Old", Type: "expense", Icon: "📁", UserID: "u1"}
	repo.Create(cat)

	cat.Name = "Updated"
	if err := repo.Update(cat); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.FindByID(cat.ID, "u1")
	if found.Name != "Updated" {
		t.Errorf("Expected 'Updated', got '%s'", found.Name)
	}
}

func TestCategoryRepository_Delete(t *testing.T) {
	setupTestDB(t)
	repo := NewCategoryRepository(config.DB)

	cat := &model.Category{ID: uuid.New().String(), Name: "ToDelete", Type: "expense", Icon: "🗑️", UserID: "u1"}
	repo.Create(cat)

	if err := repo.Delete(cat.ID, "u1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	found, _ := repo.FindByID(cat.ID, "u1")
	if found != nil && found.DeletedAt == nil {
		t.Error("Expected deleted_at to be set after soft delete")
	}
}

func TestGoalRepository_CRUD(t *testing.T) {
	setupTestDB(t)
	repo := NewGoalRepository(config.DB)

	goal := &model.SavingGoal{
		ID:           uuid.New().String(),
		UserID:       "u1",
		Name:         "New Car",
		TargetAmount: 100000000,
		Status:       "active",
	}
	if err := repo.Create(goal); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(goal.ID, "u1")
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Name != "New Car" {
		t.Errorf("Expected 'New Car', got '%s'", found.Name)
	}

	goals, _ := repo.FindAll("u1", "")
	if len(goals) < 1 {
		t.Error("Expected at least 1 goal")
	}

	goal.Status = "archived"
	repo.Update(goal)
	updated, _ := repo.FindByID(goal.ID, "u1")
	if updated.Status != "archived" {
		t.Errorf("Expected 'archived', got '%s'", updated.Status)
	}

	repo.Delete(goal.ID, "u1")
	_, err = repo.FindByID(goal.ID, "u1")
	if err == nil {
		t.Error("Expected error after delete")
	}
}

func TestGoalRepository_Contributions(t *testing.T) {
	setupTestDB(t)
	repo := NewGoalRepository(config.DB)
	goal := &model.SavingGoal{ID: uuid.New().String(), UserID: "u1", Name: "Goal", TargetAmount: 1000, Status: "active"}
	repo.Create(goal)

	contrib := &model.GoalContribution{
		ID:     uuid.New().String(),
		GoalID: goal.ID,
		Amount: 500,
		Date:   "2026-06-01",
	}
	if err := repo.CreateContribution(contrib); err != nil {
		t.Fatalf("CreateContribution failed: %v", err)
	}

	contribs, _, _ := repo.FindContributionsByGoalID(goal.ID, 1, 10)
	if len(contribs) != 1 {
		t.Errorf("Expected 1 contribution, got %d", len(contribs))
	}
}

func TestBudgetRepository_CRUD(t *testing.T) {
	setupTestDB(t)
	repo := NewBudgetRepository(config.DB)

	budget := &model.Budget{
		ID:         uuid.New().String(),
		UserID:     "u1",
		Name:       "Food Budget",
		Amount:     2000000,
		Period:     "monthly",
		CategoryID: "cat-1",
	}
	if err := repo.Create(budget); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(budget.ID, "u1")
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Name != "Food Budget" {
		t.Errorf("Expected 'Food Budget', got '%s'", found.Name)
	}

	all, _ := repo.FindAll("u1", "")
	if len(all) < 1 {
		t.Error("Expected at least 1 budget")
	}

	budget.Name = "Updated Budget"
	repo.Update(budget)
	updated, _ := repo.FindByID(budget.ID, "u1")
	if updated.Name != "Updated Budget" {
		t.Errorf("Expected 'Updated Budget', got '%s'", updated.Name)
	}

	repo.Delete(budget.ID, "u1")
	_, err = repo.FindByID(budget.ID, "u1")
	if err == nil {
		t.Error("Expected error after delete")
	}
}
