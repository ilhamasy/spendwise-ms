package service

import (
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"
	"spendwise-ms/internal/repository"

	"github.com/google/uuid"
)

func TestBoost_CreateTransaction(t *testing.T) {
	uid := boostUser(t)
	cat := boostCat(t, uid)
	svc := NewTransactionService(config.DB)
	resp, err := svc.CreateTransaction(uid, dto.CreateTransactionRequest{Type: "expense", Amount: 75000, CategoryID: cat, OccurredAt: "2026-06-28", Note: "Test"})
	if err != nil { t.Fatalf("failed: %v", err) }
	if resp.Amount != 75000 { t.Errorf("amount: %d", resp.Amount) }
}

func TestBoost_CreateTransactionBadCategory(t *testing.T) {
	uid := boostUser(t)
	svc := NewTransactionService(config.DB)
	_, err := svc.CreateTransaction(uid, dto.CreateTransactionRequest{Type: "expense", Amount: 50000, CategoryID: "no-such", OccurredAt: "2026-06-28"})
	if err == nil { t.Error("expected error") }
}

func TestBoost_GetTransactionsEmpty(t *testing.T) {
	uid := boostUser(t)
	svc := NewTransactionService(config.DB)
	resp, err := svc.GetTransactions(uid, repository.TransactionFilter{Page: 1, Limit: 20})
	if err != nil { t.Fatalf("failed: %v", err) }
	if resp.Meta.Total != 0 { t.Errorf("expected 0, got %d", resp.Meta.Total) }
}

func TestBoost_UpdateTransaction(t *testing.T) {
	uid := boostUser(t)
	cat := boostCat(t, uid)
	svc := NewTransactionService(config.DB)
	r, _ := svc.CreateTransaction(uid, dto.CreateTransactionRequest{Type: "expense", Amount: 10000, CategoryID: cat, OccurredAt: "2026-06-28"})
	u, err := svc.UpdateTransaction(uid, r.ID, dto.UpdateTransactionRequest{Amount: 25000, Note: "updated"})
	if err != nil { t.Fatalf("failed: %v", err) }
	if u.Amount != 25000 { t.Errorf("amount: %d", u.Amount) }
}

func TestBoost_UpdateTransactionNotFound(t *testing.T) {
	svc := NewTransactionService(config.DB)
	_, err := svc.UpdateTransaction(boostUser(t), "nope", dto.UpdateTransactionRequest{Amount: 10000})
	if err == nil { t.Error("expected error") }
}

func TestBoost_GetTransactionByID(t *testing.T) {
	uid := boostUser(t)
	cat := boostCat(t, uid)
	svc := NewTransactionService(config.DB)
	r, _ := svc.CreateTransaction(uid, dto.CreateTransactionRequest{Type: "expense", Amount: 50000, CategoryID: cat, OccurredAt: "2026-06-28"})
	f, err := svc.GetTransactionByID(uid, r.ID)
	if err != nil { t.Fatalf("failed: %v", err) }
	if f.ID != r.ID { t.Error("mismatch") }
}

func TestBoost_GetTransactionNotFound(t *testing.T) {
	svc := NewTransactionService(config.DB)
	_, err := svc.GetTransactionByID(boostUser(t), "nope")
	if err == nil { t.Error("expected error") }
}

func TestBoost_DeleteTransaction(t *testing.T) {
	uid := boostUser(t)
	cat := boostCat(t, uid)
	svc := NewTransactionService(config.DB)
	r, _ := svc.CreateTransaction(uid, dto.CreateTransactionRequest{Type: "expense", Amount: 10000, CategoryID: cat, OccurredAt: "2026-06-28"})
	if err := svc.DeleteTransaction(uid, r.ID); err != nil { t.Errorf("failed: %v", err) }
}

func TestBoost_DeleteTransactionNotFound(t *testing.T) {
	svc := NewTransactionService(config.DB)
	if err := svc.DeleteTransaction(boostUser(t), "nope"); err == nil { t.Error("expected error") }
}

func TestBoost_CreateGoal(t *testing.T) {
	svc := NewGoalService(config.DB)
	r, err := svc.CreateGoal(boostUser(t), dto.CreateGoalRequest{Name: "Laptop", TargetAmount: 15000000, TargetDate: "2026-12-31"})
	if err != nil { t.Fatalf("failed: %v", err) }
	if r.Name != "Laptop" { t.Errorf("name: %s", r.Name) }
}

func TestBoost_CreateGoalBadAmount(t *testing.T) {
	svc := NewGoalService(config.DB)
	_, err := svc.CreateGoal(boostUser(t), dto.CreateGoalRequest{Name: "Bad", TargetAmount: -100})
	if err == nil { t.Error("expected error") }
}

func TestBoost_UpdateGoal(t *testing.T) {
	svc := NewGoalService(config.DB)
	uid := boostUser(t)
	r, _ := svc.CreateGoal(uid, dto.CreateGoalRequest{Name: "Old", TargetAmount: 1000000})
	u, err := svc.UpdateGoal(uid, r.ID, dto.UpdateGoalRequest{Name: "New", TargetAmount: 2000000})
	if err != nil { t.Fatalf("failed: %v", err) }
	if u.Name != "New" { t.Errorf("name: %s", u.Name) }
}

func TestBoost_GetGoal(t *testing.T) {
	svc := NewGoalService(config.DB)
	uid := boostUser(t)
	r, _ := svc.CreateGoal(uid, dto.CreateGoalRequest{Name: "Find", TargetAmount: 500000})
	f, err := svc.GetGoal(uid, r.ID)
	if err != nil { t.Fatalf("failed: %v", err) }
	if f.ID != r.ID { t.Error("mismatch") }
}

func TestBoost_ArchiveUnarchiveGoal(t *testing.T) {
	svc := NewGoalService(config.DB)
	uid := boostUser(t)
	r, _ := svc.CreateGoal(uid, dto.CreateGoalRequest{Name: "Arc", TargetAmount: 100000})
	if _, err := svc.ArchiveGoal(uid, r.ID); err != nil { t.Fatalf("archive: %v", err) }
	if _, err := svc.UnarchiveGoal(uid, r.ID); err != nil { t.Fatalf("unarchive: %v", err) }
}

func TestBoost_DeleteGoal(t *testing.T) {
	svc := NewGoalService(config.DB)
	uid := boostUser(t)
	r, _ := svc.CreateGoal(uid, dto.CreateGoalRequest{Name: "Del", TargetAmount: 100000})
	if err := svc.DeleteGoal(uid, r.ID); err != nil { t.Errorf("failed: %v", err) }
}

func TestBoost_AddContribution(t *testing.T) {
	svc := NewGoalService(config.DB)
	uid := boostUser(t)
	r, _ := svc.CreateGoal(uid, dto.CreateGoalRequest{Name: "C", TargetAmount: 1000000})
	_, _, err := svc.AddContribution(uid, r.ID, dto.CreateContributionRequest{Amount: 100000, Date: "2026-06-28"})
	if err != nil { t.Errorf("failed: %v", err) }
}

func TestBoost_GetContributions(t *testing.T) {
	svc := NewGoalService(config.DB)
	uid := boostUser(t)
	r, _ := svc.CreateGoal(uid, dto.CreateGoalRequest{Name: "CG", TargetAmount: 1000000})
	svc.AddContribution(uid, r.ID, dto.CreateContributionRequest{Amount: 50000, Date: "2026-06-28"})
	resp, err := svc.GetContributions(uid, r.ID, 1, 20)
	if err != nil { t.Fatalf("failed: %v", err) }
	if len(resp.Data) == 0 { t.Error("empty") }
}

func TestBoost_ListGoals(t *testing.T) {
	svc := NewGoalService(config.DB)
	_, err := svc.ListGoals(boostUser(t), "")
	if err != nil { t.Fatalf("failed: %v", err) }
}

func TestBoost_CreateBudget(t *testing.T) {
	uid := boostUser(t)
	cat := boostCat(t, uid)
	svc := NewBudgetService(config.DB)
	r, err := svc.CreateBudget(uid, dto.CreateBudgetRequest{Name: "Food", Amount: 2000000, Period: "monthly", CategoryID: cat})
	if err != nil { t.Fatalf("failed: %v", err) }
	if r.Name != "Food" { t.Errorf("name: %s", r.Name) }
}

func TestBoost_CreateBudgetBadCategory(t *testing.T) {
	svc := NewBudgetService(config.DB)
	_, err := svc.CreateBudget(boostUser(t), dto.CreateBudgetRequest{Name: "Bad", Amount: 100000, Period: "monthly", CategoryID: "nope"})
	if err == nil { t.Error("expected error") }
}

func TestBoost_UpdateBudget(t *testing.T) {
	uid := boostUser(t)
	cat := boostCat(t, uid)
	svc := NewBudgetService(config.DB)
	r, _ := svc.CreateBudget(uid, dto.CreateBudgetRequest{Name: "Old", Amount: 100000, Period: "monthly", CategoryID: cat})
	u, err := svc.UpdateBudget(uid, r.ID, dto.UpdateBudgetRequest{Name: "New", Amount: 200000})
	if err != nil { t.Fatalf("failed: %v", err) }
	if u.Name != "New" { t.Errorf("name: %s", u.Name) }
}

func TestBoost_UpdateBudgetNotFound(t *testing.T) {
	svc := NewBudgetService(config.DB)
	_, err := svc.UpdateBudget(boostUser(t), "nope", dto.UpdateBudgetRequest{Name: "X"})
	if err == nil { t.Error("expected error") }
}

func TestBoost_GetBudget(t *testing.T) {
	uid := boostUser(t)
	cat := boostCat(t, uid)
	svc := NewBudgetService(config.DB)
	r, _ := svc.CreateBudget(uid, dto.CreateBudgetRequest{Name: "G", Amount: 500000, Period: "weekly", CategoryID: cat})
	f, err := svc.GetBudget(uid, r.ID)
	if err != nil { t.Fatalf("failed: %v", err) }
	if f.ID != r.ID { t.Error("mismatch") }
}

func TestBoost_GetBudgetNotFound(t *testing.T) {
	svc := NewBudgetService(config.DB)
	_, err := svc.GetBudget(boostUser(t), "nope")
	if err == nil { t.Error("expected error") }
}

func TestBoost_ListBudgets(t *testing.T) {
	svc := NewBudgetService(config.DB)
	_, err := svc.ListBudgets(boostUser(t), "")
	if err != nil { t.Fatalf("failed: %v", err) }
}

func TestBoost_DeleteBudget(t *testing.T) {
	uid := boostUser(t)
	cat := boostCat(t, uid)
	svc := NewBudgetService(config.DB)
	r, _ := svc.CreateBudget(uid, dto.CreateBudgetRequest{Name: "Del", Amount: 100000, Period: "daily", CategoryID: cat})
	if err := svc.DeleteBudget(uid, r.ID); err != nil { t.Errorf("failed: %v", err) }
}

func TestBoost_CreateCategory(t *testing.T) {
	svc := NewCategoryService(config.DB)
	r, err := svc.CreateCategory(boostUser(t), dto.CreateCategoryRequest{Name: "NewCat", Type: "expense", Icon: "📁", Color: "#000"})
	if err != nil { t.Fatalf("failed: %v", err) }
	if r.Name != "NewCat" { t.Errorf("name: %s", r.Name) }
}

func TestBoost_UpdateCategory(t *testing.T) {
	uid := boostUser(t)
	svc := NewCategoryService(config.DB)
	r, _ := svc.CreateCategory(uid, dto.CreateCategoryRequest{Name: "Old", Type: "expense"})
	u, err := svc.UpdateCategory(uid, r.ID, dto.UpdateCategoryRequest{Name: "Renamed", Color: "#FFF"})
	if err != nil { t.Fatalf("failed: %v", err) }
	if u.Name != "Renamed" { t.Errorf("name: %s", u.Name) }
}

func TestBoost_ListCategories(t *testing.T) {
	svc := NewCategoryService(config.DB)
	_, err := svc.ListCategories(boostUser(t), "")
	if err != nil { t.Fatalf("failed: %v", err) }
}

func TestBoost_DeleteCategory(t *testing.T) {
	uid := boostUser(t)
	svc := NewCategoryService(config.DB)
	r, _ := svc.CreateCategory(uid, dto.CreateCategoryRequest{Name: "DelCat", Type: "expense"})
	if err := svc.DeleteCategory(uid, r.ID, ""); err != nil { t.Errorf("failed: %v", err) }
}

func TestBoost_SyncTransactions(t *testing.T) {
	uid := boostUser(t)
	cat := boostCat(t, uid)
	svc := NewTransactionService(config.DB)
	resp, err := svc.SyncTransactions(uid, dto.SyncRequest{Transactions: []dto.SyncTransactionRequest{
		{Type: "expense", Amount: 10000, CategoryID: cat, OccurredAt: "2026-06-28"},
		{Type: "expense", Amount: 20000, CategoryID: cat, OccurredAt: "2026-06-28"},
	}})
	if err != nil { t.Fatalf("failed: %v", err) }
	if resp.Synced != 2 { t.Errorf("synced: %d", resp.Synced) }
}

func TestBoost_SyncTransactionsEmpty(t *testing.T) {
	svc := NewTransactionService(config.DB)
	_, err := svc.SyncTransactions(boostUser(t), dto.SyncRequest{Transactions: []dto.SyncTransactionRequest{}})
	if err == nil { t.Error("expected error") }
}

func TestBoost_SyncData(t *testing.T) {
	uid := boostUser(t)
	cat := boostCat(t, uid)
	svc := NewSyncService(config.DB)
	resp, err := svc.Sync(uid, dto.DataSyncRequest{LastSyncTimestamp: "", Changes: []dto.SyncChange{
		{EntityType: "transaction", EntityID: uuid.New().String(), Operation: "CREATE", Payload: map[string]interface{}{"type": "expense", "amount": 10000, "categoryId": cat, "occurredAt": "2026-06-28"}},
	}})
	if err != nil { t.Fatalf("failed: %v", err) }
	if resp == nil { t.Error("nil") }
}

func TestBoost_SyncDataEmpty(t *testing.T) {
	svc := NewSyncService(config.DB)
	resp, err := svc.Sync(boostUser(t), dto.DataSyncRequest{LastSyncTimestamp: "", Changes: []dto.SyncChange{}})
	if err != nil { t.Fatalf("failed: %v", err) }
	if resp == nil { t.Error("nil") }
}

func boostUser(t *testing.T) string {
	t.Helper()
	hash, _ := HashPassword("pass123")
	uid := uuid.New().String()
	config.DB.Create(&model.User{ID: uid, Name: "B", Email: "b"+uid[:8]+"@t.com", Password: hash})
	return uid
}

func boostCat(t *testing.T, uid string) string {
	t.Helper()
	id := uuid.New().String()
	config.DB.Create(&model.Category{ID: id, UserID: uid, Name: "BCat", Type: "expense", IsDefault: true})
	return id
}
