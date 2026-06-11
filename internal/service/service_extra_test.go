package service

import (
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func setupSvc(t *testing.T) string {
	t.Helper()
	if config.DB == nil {
		cfg := config.Load()
		config.InitDB(cfg)
		config.AutoMigrate(&model.User{}, &model.Transaction{}, &model.Category{}, &model.SavingGoal{}, &model.GoalContribution{}, &model.Budget{})
	}
	config.DB.Where("1=1").Delete(&model.Transaction{})
	config.DB.Where("1=1").Delete(&model.Category{})
	config.DB.Where("1=1").Delete(&model.Budget{})
	config.DB.Where("1=1").Delete(&model.GoalContribution{})
	config.DB.Where("1=1").Delete(&model.SavingGoal{})
	config.DB.Where("1=1").Delete(&model.User{})

	hash, _ := bcrypt.GenerateFromPassword([]byte("pass123"), 12)
	uid := uuid.New().String()
	config.DB.Create(&model.User{ID: uid, Name: "T", Email: uid[:8] + "@t.com", Password: string(hash)})
	return uid
}

func makeCat(t *testing.T, uid, name string) string {
	t.Helper()
	id := uuid.New().String()
	config.DB.Create(&model.Category{ID: id, Name: name, Type: "expense", Icon: "📁", UserID: uid})
	return id
}

func TestTxService_CreateGetUpdateDelete(t *testing.T) {
	uid := setupSvc(t)
	catID := makeCat(t, uid, "Food")
	svc := NewTransactionService(config.DB)

	// Create
	tx, err := svc.CreateTransaction(uid, dto.CreateTransactionRequest{Type: "expense", Amount: 50000, CategoryID: catID, OccurredAt: "2026-06-09"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Get
	got, err := svc.GetTransactionByID(uid, tx.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got == nil || got.Amount != 50000 {
		t.Errorf("GetByID mismatch: %+v", got)
	}

	// Update
	upd, err := svc.UpdateTransaction(uid, tx.ID, dto.UpdateTransactionRequest{Amount: 75000, Note: "new note"})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if upd == nil {
		t.Fatal("Update returned nil")
	}
	if upd.Amount != 75000 || upd.Note != "new note" {
		t.Errorf("Update mismatch: amount=%d note=%s", upd.Amount, upd.Note)
	}

	// Delete
	if err := svc.DeleteTransaction(uid, tx.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := svc.GetTransactionByID(uid, tx.ID); err == nil {
		t.Error("Should not find deleted transaction")
	}
}

func TestCatService_CreateUpdateDelete(t *testing.T) {
	uid := setupSvc(t)
	svc := NewCategoryService(config.DB)

	cat, err := svc.CreateCategory(uid, dto.CreateCategoryRequest{Name: "Test", Type: "expense", Icon: "📁", Color: "#000"})
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}

	upd, err := svc.UpdateCategory(uid, cat.ID, dto.UpdateCategoryRequest{Name: "Renamed", Icon: "🔄"})
	if err != nil {
		t.Fatalf("UpdateCategory: %v", err)
	}
	if upd == nil || upd.Name != "Renamed" {
		t.Errorf("UpdateCategory mismatch: %+v", upd)
	}

	if err := svc.DeleteCategory(uid, cat.ID, ""); err != nil {
		t.Fatalf("DeleteCategory: %v", err)
	}
}

func TestBudgetService_CreateUpdateDelete(t *testing.T) {
	uid := setupSvc(t)
	catID := makeCat(t, uid, "Food")
	svc := NewBudgetService(config.DB)

	bud, err := svc.CreateBudget(uid, dto.CreateBudgetRequest{Name: "Food Budget", Amount: 2000000, Period: "monthly", CategoryID: catID})
	if err != nil {
		t.Fatalf("CreateBudget: %v", err)
	}

	upd, err := svc.UpdateBudget(uid, bud.ID, dto.UpdateBudgetRequest{Name: "Updated", Amount: 3000000})
	if err != nil {
		t.Fatalf("UpdateBudget: %v", err)
	}
	if upd == nil || upd.Name != "Updated" {
		t.Errorf("UpdateBudget mismatch: %+v", upd)
	}

	if err := svc.DeleteBudget(uid, bud.ID); err != nil {
		t.Fatalf("DeleteBudget: %v", err)
	}
}

func TestSyncService_SyncEmpty(t *testing.T) {
	setupSvc(t)
	svc := NewSyncService(config.DB)
	resp, err := svc.Sync("nonexistent", dto.DataSyncRequest{Changes: []dto.SyncChange{}, LastSyncTimestamp: ""})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if resp.NewSyncTimestamp == "" {
		t.Error("Expected non-empty timestamp")
	}
}
