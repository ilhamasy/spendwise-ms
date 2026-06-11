package service

import (
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"

	"github.com/google/uuid"
)

func setupSvc2(t *testing.T) string {
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

	uid := uuid.New().String()
	config.DB.Create(&model.User{ID: uid, Name: "T", Email: uid[:8] + "@t.com", Password: "hash", Currency: "IDR", Theme: "system"})
	return uid
}

func TestGoalSvc_NotFound(t *testing.T) {
	_ = setupSvc2(t)
	svc := NewGoalService(config.DB)
	_, err := svc.GetGoal("nonexistent", "nonexistent")
	if err == nil {
		t.Error("Expected error")
	}
}

func TestGoalSvc_AddContribution(t *testing.T) {
	uid := setupSvc2(t)
	svc := NewGoalService(config.DB)
	goal, _ := svc.CreateGoal(uid, dto.CreateGoalRequest{Name: "Test", TargetAmount: 100000})

	_, _, err := svc.AddContribution(uid, goal.ID, dto.CreateContributionRequest{Amount: 25000, Note: "Deposit"})
	if err != nil {
		t.Fatalf("AddContribution: %v", err)
	}

	g, _ := svc.GetGoal(uid, goal.ID)
	if g.CurrentSaved != 25000 {
		t.Errorf("Expected 25000, got %d", g.CurrentSaved)
	}
}

func TestGoalSvc_ArchiveOnly(t *testing.T) {
	uid := setupSvc2(t)
	svc := NewGoalService(config.DB)
	goal, _ := svc.CreateGoal(uid, dto.CreateGoalRequest{Name: "Test", TargetAmount: 1000})

	a, _ := svc.ArchiveGoal(uid, goal.ID)
	if a.Status != "archived" {
		t.Errorf("Expected archived, got %s", a.Status)
	}
}

func TestGoalSvc_Contributions(t *testing.T) {
	uid := setupSvc2(t)
	svc := NewGoalService(config.DB)
	goal, _ := svc.CreateGoal(uid, dto.CreateGoalRequest{Name: "Test", TargetAmount: 100000})
	svc.AddContribution(uid, goal.ID, dto.CreateContributionRequest{Amount: 10000})
	svc.AddContribution(uid, goal.ID, dto.CreateContributionRequest{Amount: 20000})

	resp, err := svc.GetContributions(uid, goal.ID, 1, 10)
	if err != nil {
		t.Fatalf("GetContributions: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Errorf("Expected 2, got %d", len(resp.Data))
	}
}

func TestGoalSvc_DeleteWithContribs(t *testing.T) {
	uid := setupSvc2(t)
	svc := NewGoalService(config.DB)
	goal, _ := svc.CreateGoal(uid, dto.CreateGoalRequest{Name: "Test", TargetAmount: 1000})
	svc.AddContribution(uid, goal.ID, dto.CreateContributionRequest{Amount: 500})

	svc.DeleteGoal(uid, goal.ID)
	_, err := svc.GetGoal(uid, goal.ID)
	if err == nil {
		t.Error("Expected error after delete")
	}
}

func TestBudgetSvc_ListAndFilter(t *testing.T) {
	uid := setupSvc2(t)
	svc := NewBudgetService(config.DB)
	cat := &model.Category{ID: uuid.New().String(), Name: "Food", Type: "expense", Icon: "🍔", UserID: uid}
	config.DB.Create(cat)

	svc.CreateBudget(uid, dto.CreateBudgetRequest{Name: "Monthly", Amount: 1000000, Period: "monthly", CategoryID: cat.ID})
	svc.CreateBudget(uid, dto.CreateBudgetRequest{Name: "Daily", Amount: 50000, Period: "daily", CategoryID: cat.ID})

	all, _ := svc.ListBudgets(uid, "")
	if len(all) < 2 {
		t.Errorf("Expected >=2, got %d", len(all))
	}

	monthly, _ := svc.ListBudgets(uid, "monthly")
	if len(monthly) < 1 {
		t.Errorf("Expected >=1 monthly, got %d", len(monthly))
	}
}

func TestBudgetSvc_GetNotFound(t *testing.T) {
	_ = setupSvc2(t)
	svc := NewBudgetService(config.DB)
	_, err := svc.GetBudget("nonexistent", "nonexistent")
	if err == nil {
		t.Error("Expected error")
	}
}

func TestProfileSvc_GetUpdate(t *testing.T) {
	uid := setupSvc2(t)
	svc := NewProfileService(config.DB)

	p, _ := svc.GetProfile(uid)
	if p.Email == "" {
		t.Error("Email should not be empty")
	}

	u, _ := svc.UpdateProfile(uid, dto.UpdateProfileRequest{Name: "NewName", Currency: "USD"})
	if u.Name != "NewName" || u.Currency != "USD" {
		t.Errorf("Update mismatch")
	}
}

func TestProfileSvc_DeleteAccount(t *testing.T) {
	uid := setupSvc2(t)
	svc := NewProfileService(config.DB)
	cat := &model.Category{ID: uuid.New().String(), Name: "Food", Type: "expense", Icon: "🍔", UserID: uid}
	config.DB.Create(cat)
	config.DB.Create(&model.Transaction{ID: uuid.New().String(), UserID: uid, Type: "expense", Amount: 1000, CategoryID: cat.ID, OccurredAt: "2026-06-01"})

	if err := svc.DeleteAccount(uid, "DELETE"); err != nil {
		t.Fatalf("DeleteAccount: %v", err)
	}
	_, err := svc.GetProfile(uid)
	if err == nil {
		t.Error("Expected error after delete")
	}
}

func TestProfileSvc_ExportData(t *testing.T) {
	uid := setupSvc2(t)
	svc := NewProfileService(config.DB)
	resp, err := svc.ExportData(uid)
	if err != nil {
		t.Fatalf("ExportData: %v", err)
	}
	if resp.Profile.Email == "" {
		t.Error("Profile email should not be empty")
	}
}
