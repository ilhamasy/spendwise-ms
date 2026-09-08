package service

import (
	"testing"
	"time"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
)

func TestTransactionService_ValidationAndErrors(t *testing.T) {
	setupCategoryTestDB()
	userID := createUserForCategory()
	svc := NewTransactionService(config.DB)
	catSvc := NewCategoryService(config.DB)

	cat, _ := catSvc.CreateCategory(userID, dto.CreateCategoryRequest{Name: "Food", Type: "expense"})

	// 1. Invalid category ID
	_, err := svc.CreateTransaction(userID, dto.CreateTransactionRequest{
		Type: "expense", Amount: 100, CategoryID: "non-existent", OccurredAt: "2026-08-21",
	})
	if err == nil {
		t.Error("Expected error for non-existent category")
	}

	// 2. Invalid date format
	_, err = svc.CreateTransaction(userID, dto.CreateTransactionRequest{
		Type: "expense", Amount: 100, CategoryID: cat.ID, OccurredAt: "invalid-date",
	})
	if err == nil {
		t.Error("Expected error for invalid date format")
	}

	// 3. Future date
	futureDate := time.Now().Add(48 * time.Hour).Format("2006-01-02")
	_, err = svc.CreateTransaction(userID, dto.CreateTransactionRequest{
		Type: "expense", Amount: 100, CategoryID: cat.ID, OccurredAt: futureDate,
	})
	if err == nil {
		t.Error("Expected error for future date")
	}

	// 4. Update Transaction errors & validation
	created, _ := svc.CreateTransaction(userID, dto.CreateTransactionRequest{
		Type: "expense", Amount: 1000, CategoryID: cat.ID, OccurredAt: "2026-08-21", Note: "Original",
	})

	// Update non-existent
	_, err = svc.UpdateTransaction(userID, "non-existent", dto.UpdateTransactionRequest{Amount: 2000})
	if err == nil {
		t.Error("Expected error for non-existent transaction update")
	}

	// Update with invalid date
	_, err = svc.UpdateTransaction(userID, created.ID, dto.UpdateTransactionRequest{OccurredAt: "invalid-date"})
	if err == nil {
		t.Error("Expected error for invalid date format in update")
	}

	// Update with future date
	_, err = svc.UpdateTransaction(userID, created.ID, dto.UpdateTransactionRequest{OccurredAt: futureDate})
	if err == nil {
		t.Error("Expected error for future date in update")
	}

	// Update valid fields
	updated, err := svc.UpdateTransaction(userID, created.ID, dto.UpdateTransactionRequest{
		Type: "income", Amount: 5000, CategoryID: cat.ID, OccurredAt: "2026-08-21", Note: "Updated",
	})
	if err != nil || updated.Note != "Updated" {
		t.Fatalf("UpdateTransaction failed: %v", err)
	}

	// 5. Delete non-existent
	err = svc.DeleteTransaction(userID, "non-existent")
	if err == nil {
		t.Error("Expected error for non-existent transaction delete")
	}

	// 6. Get by ID non-existent
	_, err = svc.GetTransactionByID(userID, "non-existent")
	if err == nil {
		t.Error("Expected error for non-existent transaction get")
	}

	// 7. SyncTransactions
	_, err = svc.SyncTransactions(userID, dto.SyncRequest{Transactions: []dto.SyncTransactionRequest{}})
	if err == nil {
		t.Error("Expected error for empty sync transactions")
	}

	resSync, err := svc.SyncTransactions(userID, dto.SyncRequest{
		Transactions: []dto.SyncTransactionRequest{
			{Type: "expense", Amount: 100, CategoryID: cat.ID, OccurredAt: "2026-08-21", Note: "valid"},
			{Type: "invalid-type", Amount: 0, CategoryID: cat.ID, OccurredAt: "2026-08-21"},
			{Type: "expense", Amount: 50, CategoryID: "non-existent-cat", OccurredAt: "2026-08-21"},
			{Type: "expense", Amount: 50, CategoryID: cat.ID, OccurredAt: "bad-date"},
			{Type: "expense", Amount: 50, CategoryID: cat.ID, OccurredAt: "2099-12-31"},
		},
	})
	if err != nil || resSync == nil {
		t.Fatalf("SyncTransactions failed: %v", err)
	}
	if resSync.Synced != 1 || len(resSync.Failed) != 4 {
		t.Errorf("Expected 1 synced and 4 failed, got %d synced and %d failed", resSync.Synced, len(resSync.Failed))
	}
}

func TestBudgetService_ValidationAndErrors(t *testing.T) {
	setupCategoryTestDB()
	userID := createUserForCategory()
	svc := NewBudgetService(config.DB)
	catSvc := NewCategoryService(config.DB)

	cat, _ := catSvc.CreateCategory(userID, dto.CreateCategoryRequest{Name: "Bills", Type: "expense"})

	// 1. Create budget non-existent category
	_, err := svc.CreateBudget(userID, dto.CreateBudgetRequest{
		Name: "Test", Amount: 100, Period: "monthly", CategoryID: "non-existent",
	})
	if err == nil {
		t.Error("Expected error for non-existent category in CreateBudget")
	}

	// 2. Update non-existent budget
	_, err = svc.UpdateBudget(userID, "non-existent", dto.UpdateBudgetRequest{Name: "Update"})
	if err == nil {
		t.Error("Expected error for non-existent budget update")
	}

	// 3. Update budget non-existent category
	created, _ := svc.CreateBudget(userID, dto.CreateBudgetRequest{
		Name: "Test", Amount: 100000, Period: "monthly", CategoryID: cat.ID,
	})
	_, err = svc.UpdateBudget(userID, created.ID, dto.UpdateBudgetRequest{CategoryID: "non-existent"})
	if err == nil {
		t.Error("Expected error for non-existent category in UpdateBudget")
	}
}
