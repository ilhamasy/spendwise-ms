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

func TestSvc_UpdateBudget(t *testing.T) {
	uid := setupServiceTest(t)
	catID := createSvcCategory(t, uid, "BudgetUpdCat", "expense")

	svc := NewBudgetService(config.DB)
	resp, _ := svc.CreateBudget(uid, dto.CreateBudgetRequest{
		Name: "Old Budget", Amount: 100000, Period: "monthly", CategoryID: catID,
	})

	updated, err := svc.UpdateBudget(uid, resp.ID, dto.UpdateBudgetRequest{
		Name: "Updated Budget", Amount: 200000,
	})
	if err != nil {
		t.Fatalf("UpdateBudget failed: %v", err)
	}
	if updated.Name != "Updated Budget" || updated.Amount != 200000 {
		t.Errorf("unexpected: %+v", updated)
	}
}

func TestSvc_UpdateBudget_NotFound(t *testing.T) {
	uid := setupServiceTest(t)
	svc := NewBudgetService(config.DB)

	_, err := svc.UpdateBudget(uid, "non-existent", dto.UpdateBudgetRequest{Name: "X"})
	if err == nil {
		t.Error("expected error for non-existent budget")
	}
}

func TestSvc_GetBudget(t *testing.T) {
	uid := setupServiceTest(t)
	catID := createSvcCategory(t, uid, "GetBudgetCat", "expense")
	svc := NewBudgetService(config.DB)
	resp, _ := svc.CreateBudget(uid, dto.CreateBudgetRequest{
		Name: "My Budget", Amount: 500000, Period: "weekly", CategoryID: catID,
	})

	found, err := svc.GetBudget(uid, resp.ID)
	if err != nil {
		t.Fatalf("GetBudget failed: %v", err)
	}
	if found.ID != resp.ID {
		t.Errorf("ID mismatch")
	}
}

func TestSvc_GetBudget_NotFound(t *testing.T) {
	uid := setupServiceTest(t)
	svc := NewBudgetService(config.DB)

	_, err := svc.GetBudget(uid, "non-existent")
	if err == nil {
		t.Error("expected error for non-existent budget")
	}
}

func TestSvc_ListBudgets(t *testing.T) {
	uid := setupServiceTest(t)
	svc := NewBudgetService(config.DB)

	resp, err := svc.ListBudgets(uid, "")
	if err != nil {
		t.Fatalf("ListBudgets failed: %v", err)
	}
	if resp == nil {
		t.Log("ListBudgets returned nil (expected for empty)")
	}
}

func TestSvc_DeleteBudget(t *testing.T) {
	uid := setupServiceTest(t)
	catID := createSvcCategory(t, uid, "DelBudgetCat", "expense")
	svc := NewBudgetService(config.DB)
	resp, _ := svc.CreateBudget(uid, dto.CreateBudgetRequest{
		Name: "ToDelete", Amount: 100000, Period: "daily", CategoryID: catID,
	})

	err := svc.DeleteBudget(uid, resp.ID)
	if err != nil {
		t.Errorf("DeleteBudget failed: %v", err)
	}
}

func TestSvc_CreateCategory(t *testing.T) {
	uid := setupServiceTest(t)

	svc := NewCategoryService(config.DB)
	resp, err := svc.CreateCategory(uid, dto.CreateCategoryRequest{
		Name: "NewCat", Type: "expense", Icon: "📁", Color: "#000",
	})
	if err != nil {
		t.Fatalf("CreateCategory failed: %v", err)
	}
	if resp.Name != "NewCat" {
		t.Errorf("unexpected name: %s", resp.Name)
	}
}

func TestSvc_UpdateCategory(t *testing.T) {
	uid := setupServiceTest(t)
	svc := NewCategoryService(config.DB)
	resp, _ := svc.CreateCategory(uid, dto.CreateCategoryRequest{
		Name: "OldCat", Type: "expense",
	})

	updated, err := svc.UpdateCategory(uid, resp.ID, dto.UpdateCategoryRequest{
		Name: "RenamedCat", Color: "#FFF",
	})
	if err != nil {
		t.Fatalf("UpdateCategory failed: %v", err)
	}
	if updated.Name != "RenamedCat" || updated.Color != "#FFF" {
		t.Errorf("unexpected: %+v", updated)
	}
}

func TestSvc_ListCategories(t *testing.T) {
	uid := setupServiceTest(t)
	svc := NewCategoryService(config.DB)

	resp, err := svc.ListCategories(uid, "")
	if err != nil {
		t.Fatalf("ListCategories failed: %v", err)
	}
	if resp == nil {
		t.Log("ListCategories returned nil (expected for empty)")
	}
}

func TestSvc_ListGoals(t *testing.T) {
	uid := setupServiceTest(t)
	svc := NewGoalService(config.DB)

	resp, err := svc.ListGoals(uid, "")
	if err != nil {
		t.Fatalf("ListGoals failed: %v", err)
	}
	if resp == nil {
		t.Log("ListGoals returned nil (expected for empty)")
	}
}

func TestSvc_UpdateGoal(t *testing.T) {
	uid := setupServiceTest(t)
	svc := NewGoalService(config.DB)
	resp, _ := svc.CreateGoal(uid, dto.CreateGoalRequest{
		Name: "Old Goal", TargetAmount: 1000000,
	})

	updated, err := svc.UpdateGoal(uid, resp.ID, dto.UpdateGoalRequest{
		Name: "Updated Goal", TargetAmount: 2000000,
	})
	if err != nil {
		t.Fatalf("UpdateGoal failed: %v", err)
	}
	if updated.Name != "Updated Goal" {
		t.Errorf("unexpected: %+v", updated)
	}
}

func TestSvc_GetGoal(t *testing.T) {
	uid := setupServiceTest(t)
	svc := NewGoalService(config.DB)
	resp, _ := svc.CreateGoal(uid, dto.CreateGoalRequest{
		Name: "FindMe", TargetAmount: 500000,
	})

	found, err := svc.GetGoal(uid, resp.ID)
	if err != nil {
		t.Fatalf("GetGoal failed: %v", err)
	}
	if found.ID != resp.ID {
		t.Error("ID mismatch")
	}
}

func TestSvc_ArchiveUnarchiveGoal(t *testing.T) {
	uid := setupServiceTest(t)
	svc := NewGoalService(config.DB)
	resp, _ := svc.CreateGoal(uid, dto.CreateGoalRequest{
		Name: "ArchiveMe", TargetAmount: 100000,
	})

	_, err := svc.ArchiveGoal(uid, resp.ID)
	if err != nil {
		t.Fatalf("ArchiveGoal failed: %v", err)
	}

	_, err = svc.UnarchiveGoal(uid, resp.ID)
	if err != nil {
		t.Fatalf("UnarchiveGoal failed: %v", err)
	}
}

func TestSvc_DeleteGoal(t *testing.T) {
	uid := setupServiceTest(t)
	svc := NewGoalService(config.DB)
	resp, _ := svc.CreateGoal(uid, dto.CreateGoalRequest{
		Name: "DeleteMe", TargetAmount: 100000,
	})

	err := svc.DeleteGoal(uid, resp.ID)
	if err != nil {
		t.Errorf("DeleteGoal failed: %v", err)
	}
}

func TestSvc_AddContribution(t *testing.T) {
	uid := setupServiceTest(t)
	svc := NewGoalService(config.DB)
	resp, _ := svc.CreateGoal(uid, dto.CreateGoalRequest{
		Name: "ContributeGoal", TargetAmount: 1000000,
	})

	_, _, err := svc.AddContribution(uid, resp.ID, dto.CreateContributionRequest{
		Amount: 100000,
		Note:   "First save",
	})
	if err != nil {
		t.Errorf("AddContribution failed: %v", err)
	}
}

func TestSvc_GetContributions(t *testing.T) {
	uid := setupServiceTest(t)
	svc := NewGoalService(config.DB)
	resp, _ := svc.CreateGoal(uid, dto.CreateGoalRequest{
		Name: "ContribGoal", TargetAmount: 1000000,
	})
	svc.AddContribution(uid, resp.ID, dto.CreateContributionRequest{Amount: 50000, Date: "2026-06-28"})

	resp2, err := svc.GetContributions(uid, resp.ID, 1, 20)
	if err != nil {
		t.Fatalf("GetContributions failed: %v", err)
	}
	if len(resp2.Data) == 0 {
		t.Error("expected at least 1 contribution")
	}
}

func TestSvc_Register(t *testing.T) {
	hash, _ := HashPassword("pass123")
	uid := uuid.New().String()
	email := "directreg" + uid[:8] + "@t.com"

	config.DB.Create(&model.User{ID: uid, Name: "DirReg", Email: email, Password: hash})

	existing, err := Register(config.DB, "NewUser", "newreg"+uid[:8]+"@t.com", "password123")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if existing.Name != "NewUser" {
		t.Errorf("unexpected name: %s", existing.Name)
	}
}

func TestSvc_Register_Duplicate(t *testing.T) {
	uid := setupServiceTest(t)

	_, err := Register(config.DB, "DupUser", "svc"+uid[:8]+"@test.com", "password123")
	if err == nil {
		t.Error("expected duplicate email error")
	}
}

func TestSvc_Login_Success(t *testing.T) {
	hash, _ := HashPassword("pass123")
	uid := uuid.New().String()
	email := "login" + uid[:8] + "@t.com"
	config.DB.Create(&model.User{ID: uid, Name: "LoginTest", Email: email, Password: hash})

	user, err := Login(config.DB, email, "pass123")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if user.Email != email {
		t.Errorf("email mismatch")
	}
}

func TestSvc_Login_Invalid(t *testing.T) {
	uid := setupServiceTest(t)

	_, err := Login(config.DB, "svc"+uid[:8]+"@test.com", "wrongpassword")
	if err == nil {
		t.Error("expected login error")
	}
}

func TestSvc_SyncTransactions(t *testing.T) {
	uid := setupServiceTest(t)
	catID := createSvcCategory(t, uid, "SyncCat", "expense")

	svc := NewTransactionService(config.DB)
	resp, err := svc.SyncTransactions(uid, dto.SyncRequest{
		Transactions: []dto.SyncTransactionRequest{
			{Type: "expense", Amount: 10000, CategoryID: catID, OccurredAt: "2026-06-28"},
			{Type: "expense", Amount: 20000, CategoryID: catID, OccurredAt: "2026-06-28"},
		},
	})
	if err != nil {
		t.Fatalf("SyncTransactions failed: %v", err)
	}
	if resp.Synced != 2 {
		t.Errorf("expected 2 synced, got %d", resp.Synced)
	}
}

func TestSvc_SyncTransactions_Empty(t *testing.T) {
	uid := setupServiceTest(t)
	svc := NewTransactionService(config.DB)

	_, err := svc.SyncTransactions(uid, dto.SyncRequest{
		Transactions: []dto.SyncTransactionRequest{},
	})
	if err == nil {
		t.Error("expected error for empty sync request")
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
