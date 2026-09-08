package repository

import (
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/model"

	"github.com/google/uuid"
)

func TestRepo_TxnCRUD(t *testing.T) {
	uid := repoBoostUser(t)
	cat := repoBoostCat(t, uid)
	repo := NewTransactionRepository(config.DB)

	txn := &model.Transaction{ID: uuid.New().String(), UserID: uid, Type: "expense", Amount: 50000, CategoryID: cat, OccurredAt: "2026-06-28"}
	if err := repo.Create(txn); err != nil { t.Fatalf("create: %v", err) }
	found, err := repo.FindByID(txn.ID, uid)
	if err != nil { t.Fatalf("find: %v", err) }
	if found.Amount != 50000 { t.Errorf("amount: %d", found.Amount) }

	txn.Amount = 25000
	if err := repo.Update(txn); err != nil { t.Fatalf("update: %v", err) }

	if err := repo.Delete(txn.ID, uid); err != nil { t.Errorf("delete: %v", err) }
}

func TestRepo_TxnNotFound(t *testing.T) {
	uid := repoBoostUser(t)
	repo := NewTransactionRepository(config.DB)
	if _, err := repo.FindByID("nope", uid); err == nil { t.Error("expected error") }
	if err := repo.Delete("nope", uid); err == nil { t.Error("expected error") }
}

func TestRepo_BatchCreate(t *testing.T) {
	uid := repoBoostUser(t)
	cat := repoBoostCat(t, uid)
	repo := NewTransactionRepository(config.DB)
	txns := []model.Transaction{
		{ID: uuid.New().String(), UserID: uid, Type: "expense", Amount: 1000, CategoryID: cat, OccurredAt: "2026-06-28"},
		{ID: uuid.New().String(), UserID: uid, Type: "income", Amount: 5000, CategoryID: cat, OccurredAt: "2026-06-28"},
	}
	s, f, _ := repo.BatchCreate(txns)
	if s != 2 { t.Errorf("synced: %d", s) }
	if len(f) != 0 { t.Errorf("failed: %d", len(f)) }
}

func TestRepo_CatCRUD(t *testing.T) {
	uid := repoBoostUser(t)
	repo := NewCategoryRepository(config.DB)
	cat := &model.Category{ID: uuid.New().String(), UserID: uid, Name: "RCat", Type: "expense", Icon: "📁", Color: "#000"}
	if err := repo.Create(cat); err != nil { t.Fatalf("create: %v", err) }
	found, err := repo.FindByID(cat.ID, uid)
	if err != nil { t.Fatalf("find: %v", err) }
	if found.Name != "RCat" { t.Errorf("name: %s", found.Name) }
}

func repoBoostUser(t *testing.T) string {
	t.Helper()
	uid := uuid.New().String()
	config.DB.Create(&model.User{ID: uid, Name: "RepoB", Email: "rb"+uid[:8]+"@t.com", Password: "x"})
	return uid
}

func repoBoostCat(t *testing.T, uid string) string {
	t.Helper()
	id := uuid.New().String()
	config.DB.Create(&model.Category{ID: id, UserID: uid, Name: "RBCat", Type: "expense", IsDefault: true})
	return id
}
