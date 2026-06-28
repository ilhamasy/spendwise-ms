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


func TestBoostSvc_GetProfile(t *testing.T) {
	hash, _ := HashPassword("x")
	uid := uuid.New().String()
	config.DB.Create(&model.User{ID: uid, Name: "GP", Email: "gp"+uid[:8]+"@t.com", Password: hash})
	svc := NewProfileService(config.DB)
	p, err := svc.GetProfile(uid)
	if err != nil { t.Fatalf("GetProfile: %v", err) }
	if p.ID != uid { t.Error("id mismatch") }
}

func TestBoostSvc_UpdateProfile(t *testing.T) {
	hash, _ := HashPassword("x")
	uid := uuid.New().String()
	config.DB.Create(&model.User{ID: uid, Name: "UP", Email: "up"+uid[:8]+"@t.com", Password: hash})
	svc := NewProfileService(config.DB)
	p, err := svc.UpdateProfile(uid, dto.UpdateProfileRequest{Name: "NewName", Currency: "USD"})
	if err != nil { t.Fatalf("UpdateProfile: %v", err) }
	if p.Name != "NewName" { t.Errorf("name: %s", p.Name) }
}

func TestBoostSvc_ChangePassword(t *testing.T) {
	hash, _ := HashPassword("old")
	uid := uuid.New().String()
	config.DB.Create(&model.User{ID: uid, Name: "CP", Email: "cp"+uid[:8]+"@t.com", Password: hash})
	svc := NewProfileService(config.DB)
	err := svc.ChangePassword(uid, "old", "newpass456")
	if err != nil { t.Fatalf("ChangePassword: %v", err) }
}

func TestBoostSvc_ExportData(t *testing.T) {
	hash, _ := HashPassword("x")
	uid := uuid.New().String()
	config.DB.Create(&model.User{ID: uid, Name: "ED", Email: "ed"+uid[:8]+"@t.com", Password: hash})
	svc := NewProfileService(config.DB)
	data, err := svc.ExportData(uid)
	if err != nil { t.Fatalf("ExportData: %v", err) }
	if data == nil { t.Error("nil data") }
}

func TestBoostSvc_DeleteAccount(t *testing.T) {
	hash, _ := HashPassword("x")
	uid := uuid.New().String()
	config.DB.Create(&model.User{ID: uid, Name: "DA", Email: "da"+uid[:8]+"@t.com", Password: hash})
	svc := NewProfileService(config.DB)
	err := svc.DeleteAccount(uid, "DELETE")
	if err != nil { t.Errorf("DeleteAccount: %v", err) }
}
