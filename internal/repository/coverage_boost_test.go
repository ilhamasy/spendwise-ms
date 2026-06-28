package repository

import (
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/model"

	"github.com/google/uuid"
)

func TestRepo_CreateAndFindTransaction(t *testing.T) {
	uid := setupRepoTest(t)
	catID := createRepoCategory(t, uid, "RepoCat")

	repo := NewTransactionRepository(config.DB)

	txn := &model.Transaction{
		ID:         uuid.New().String(),
		UserID:     uid,
		Type:       "expense",
		Amount:     50000,
		CategoryID: catID,
		OccurredAt: "2026-06-28",
	}
	err := repo.Create(txn)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(txn.ID, uid)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Amount != 50000 {
		t.Errorf("amount mismatch: %d", found.Amount)
	}
}

func TestRepo_FindTransaction_NotFound(t *testing.T) {
	uid := setupRepoTest(t)
	repo := NewTransactionRepository(config.DB)

	_, err := repo.FindByID("non-existent", uid)
	if err == nil {
		t.Error("expected error for non-existent transaction")
	}
}

func TestRepo_UpdateTransaction(t *testing.T) {
	uid := setupRepoTest(t)
	catID := createRepoCategory(t, uid, "UpdRepo")

	repo := NewTransactionRepository(config.DB)
	txn := &model.Transaction{
		ID: uuid.New().String(), UserID: uid, Type: "expense",
		Amount: 10000, CategoryID: catID, OccurredAt: "2026-06-28",
	}
	repo.Create(txn)

	txn.Amount = 25000
	err := repo.Update(txn)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.FindByID(txn.ID, uid)
	if found.Amount != 25000 {
		t.Errorf("update not persisted: %d", found.Amount)
	}
}

func TestRepo_DeleteTransaction(t *testing.T) {
	uid := setupRepoTest(t)
	catID := createRepoCategory(t, uid, "DelRepo")

	repo := NewTransactionRepository(config.DB)
	txn := &model.Transaction{
		ID: uuid.New().String(), UserID: uid, Type: "expense",
		Amount: 10000, CategoryID: catID, OccurredAt: "2026-06-28",
	}
	repo.Create(txn)

	err := repo.Delete(txn.ID, uid)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
}

func TestRepo_DeleteTransaction_NotFound(t *testing.T) {
	uid := setupRepoTest(t)
	repo := NewTransactionRepository(config.DB)

	err := repo.Delete("non-existent", uid)
	if err == nil {
		t.Error("expected error for non-existent delete")
	}
}

func TestRepo_BatchCreate(t *testing.T) {
	uid := setupRepoTest(t)
	catID := createRepoCategory(t, uid, "BatchCat")

	repo := NewTransactionRepository(config.DB)
	txns := []model.Transaction{
		{ID: uuid.New().String(), UserID: uid, Type: "expense", Amount: 1000, CategoryID: catID, OccurredAt: "2026-06-28"},
		{ID: uuid.New().String(), UserID: uid, Type: "income", Amount: 5000, CategoryID: catID, OccurredAt: "2026-06-28"},
	}
	synced, failed, _ := repo.BatchCreate(txns)
	if synced != 2 {
		t.Errorf("expected 2 synced, got %d", synced)
	}
	if len(failed) != 0 {
		t.Errorf("expected 0 failed, got %d", len(failed))
	}
}

func TestRepo_CreateAndFindCategory(t *testing.T) {
	uid := setupRepoTest(t)
	repo := NewCategoryRepository(config.DB)

	cat := &model.Category{
		ID: uuid.New().String(), UserID: uid, Name: "TestCategory",
		Type: "expense", Icon: "📁", Color: "#000",
	}
	err := repo.Create(cat)
	if err != nil {
		t.Fatalf("Create category failed: %v", err)
	}

	found, err := repo.FindByID(cat.ID, uid)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Name != "TestCategory" {
		t.Errorf("name mismatch: %s", found.Name)
	}
}

func setupRepoTest(t *testing.T) string {
	t.Helper()
	if config.DB == nil {
		t.Fatal("config.DB is nil")
	}
	uid := uuid.New().String()
	email := "repo" + uid[:8] + "@t.com"
	config.DB.Create(&model.User{ID: uid, Name: "RepoTest", Email: email, Password: "hash"})
	return uid
}

func createRepoCategory(t *testing.T, uid, name string) string {
	t.Helper()
	cat := &model.Category{ID: uuid.New().String(), UserID: uid, Name: name, Type: "expense", IsDefault: true}
	config.DB.Create(cat)
	return cat.ID
}
