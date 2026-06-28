package service

import (
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"
	"spendwise-ms/internal/repository"

	"github.com/google/uuid"
)

func TestSvc_CreateTransaction(t *testing.T) {
	uid := setupServiceTest(t)
	catID := createSvcCategory(t, uid, "TestCat", "expense")

	svc := NewTransactionService(config.DB)
	resp, err := svc.CreateTransaction(uid, dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     75000,
		CategoryID: catID,
		OccurredAt: "2026-06-28",
		Note:       "Test",
	})
	if err != nil {
		t.Fatalf("CreateTransaction failed: %v", err)
	}
	if resp.Type != "expense" || resp.Amount != 75000 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestSvc_CreateTransaction_InvalidCategory(t *testing.T) {
	uid := setupServiceTest(t)

	svc := NewTransactionService(config.DB)
	_, err := svc.CreateTransaction(uid, dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     50000,
		CategoryID: "non-existent-category",
		OccurredAt: "2026-06-28",
	})
	if err == nil {
		t.Error("expected error for invalid category")
	}
}

func TestSvc_GetTransactions_Empty(t *testing.T) {
	uid := setupServiceTest(t)

	svc := NewTransactionService(config.DB)
	resp, err := svc.GetTransactions(uid, repository.TransactionFilter{
		Page:  1,
		Limit: 20,
	})
	if err != nil {
		t.Fatalf("GetTransactions failed: %v", err)
	}
	if resp.Meta.Total != 0 {
		t.Errorf("expected 0 transactions, got %d", resp.Meta.Total)
	}
}

func TestSvc_UpdateTransaction_NotFound(t *testing.T) {
	uid := setupServiceTest(t)

	svc := NewTransactionService(config.DB)
	_, err := svc.UpdateTransaction(uid, "non-existent-id", dto.UpdateTransactionRequest{
		Amount: 10000,
	})
	if err == nil {
		t.Error("expected error for non-existent transaction")
	}
}

func TestSvc_DeleteTransaction_NotFound(t *testing.T) {
	uid := setupServiceTest(t)

	svc := NewTransactionService(config.DB)
	err := svc.DeleteTransaction(uid, "non-existent-id")
	if err == nil {
		t.Error("expected error for non-existent transaction")
	}
}

func TestSvc_DeleteTransaction_Success(t *testing.T) {
	uid := setupServiceTest(t)
	catID := createSvcCategory(t, uid, "DelCat", "expense")
	svc := NewTransactionService(config.DB)
	resp, _ := svc.CreateTransaction(uid, dto.CreateTransactionRequest{
		Type: "expense", Amount: 10000, CategoryID: catID, OccurredAt: "2026-06-28",
	})

	err := svc.DeleteTransaction(uid, resp.ID)
	if err != nil {
		t.Errorf("DeleteTransaction failed: %v", err)
	}
}

func TestSvc_UpdateTransaction_Success(t *testing.T) {
	uid := setupServiceTest(t)
	catID := createSvcCategory(t, uid, "UpdCat", "expense")
	svc := NewTransactionService(config.DB)
	resp, _ := svc.CreateTransaction(uid, dto.CreateTransactionRequest{
		Type: "expense", Amount: 10000, CategoryID: catID, OccurredAt: "2026-06-28",
	})

	updated, err := svc.UpdateTransaction(uid, resp.ID, dto.UpdateTransactionRequest{
		Amount: 25000, Note: "updated",
	})
	if err != nil {
		t.Fatalf("UpdateTransaction failed: %v", err)
	}
	if updated.Amount != 25000 || updated.Note != "updated" {
		t.Errorf("unexpected update result: %+v", updated)
	}
}

func TestSvc_GetTransactionByID_NotFound(t *testing.T) {
	uid := setupServiceTest(t)
	svc := NewTransactionService(config.DB)

	_, err := svc.GetTransactionByID(uid, "non-existent")
	if err == nil {
		t.Error("expected error for non-existent transaction")
	}
}

func TestSvc_GetTransactionByID_Success(t *testing.T) {
	uid := setupServiceTest(t)
	catID := createSvcCategory(t, uid, "GetCat", "expense")
	svc := NewTransactionService(config.DB)
	created, _ := svc.CreateTransaction(uid, dto.CreateTransactionRequest{
		Type: "expense", Amount: 50000, CategoryID: catID, OccurredAt: "2026-06-28",
	})

	found, err := svc.GetTransactionByID(uid, created.ID)
	if err != nil {
		t.Fatalf("GetTransactionByID failed: %v", err)
	}
	if found.ID != created.ID {
		t.Errorf("ID mismatch: %s vs %s", found.ID, created.ID)
	}
}

func TestSvc_CreateGoal(t *testing.T) {
	uid := setupServiceTest(t)

	svc := NewGoalService(config.DB)
	resp, err := svc.CreateGoal(uid, dto.CreateGoalRequest{
		Name:         "New Laptop",
		TargetAmount: 15000000,
		TargetDate:   "2026-12-31",
	})
	if err != nil {
		t.Fatalf("CreateGoal failed: %v", err)
	}
	if resp.Name != "New Laptop" || resp.TargetAmount != 15000000 {
		t.Errorf("unexpected goal: %+v", resp)
	}
}

func TestSvc_CreateGoal_InvalidAmount(t *testing.T) {
	uid := setupServiceTest(t)

	svc := NewGoalService(config.DB)
	_, err := svc.CreateGoal(uid, dto.CreateGoalRequest{
		Name:         "Bad Goal",
		TargetAmount: -100,
	})
	if err == nil {
		t.Error("expected error for negative target amount")
	}
}

func TestSvc_CreateBudget(t *testing.T) {
	uid := setupServiceTest(t)
	catID := createSvcCategory(t, uid, "BudgetCat", "expense")

	svc := NewBudgetService(config.DB)
	resp, err := svc.CreateBudget(uid, dto.CreateBudgetRequest{
		Name:       "Monthly Food",
		Amount:     2000000,
		Period:     "monthly",
		CategoryID: catID,
	})
	if err != nil {
		t.Fatalf("CreateBudget failed: %v", err)
	}
	if resp.Name != "Monthly Food" {
		t.Errorf("unexpected budget name: %s", resp.Name)
	}
}

func TestSvc_CreateBudget_InvalidCategory(t *testing.T) {
	uid := setupServiceTest(t)

	svc := NewBudgetService(config.DB)
	_, err := svc.CreateBudget(uid, dto.CreateBudgetRequest{
		Name:       "Bad Budget",
		Amount:     100000,
		Period:     "monthly",
		CategoryID: "non-existent",
	})
	if err == nil {
		t.Error("expected error for invalid category")
	}
}

func setupServiceTest(t *testing.T) string {
	t.Helper()
	hash, _ := HashPassword("pass123")
	uid := uuid.New().String()
	email := "svc" + uid[:8] + "@test.com"
	config.DB.Create(&model.User{ID: uid, Name: "SvcTest", Email: email, Password: hash})
	return uid
}

func createSvcCategory(t *testing.T, uid, name, catType string) string {
	t.Helper()
	cat := &model.Category{
		ID:        uuid.New().String(),
		UserID:    uid,
		Name:      name,
		Type:      catType,
		IsDefault: true,
	}
	config.DB.Create(cat)
	return cat.ID
}
