package repository

import (
	"testing"
	"time"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/model"

	"github.com/google/uuid"
)

func setupRepoTestDB() string {
	if config.DB == nil {
		config.InitDB(config.Load())
	}
	if !config.DB.Migrator().HasTable("users") {
		config.AutoMigrate(
			&model.User{}, &model.Transaction{}, &model.Category{}, &model.SavingGoal{}, &model.GoalContribution{}, &model.Budget{},
		)
	}
	config.DB.Where("1 = 1").Delete(&model.GoalContribution{})
	config.DB.Where("1 = 1").Delete(&model.Budget{})
	config.DB.Where("1 = 1").Delete(&model.SavingGoal{})
	config.DB.Where("1 = 1").Delete(&model.Transaction{})
	config.DB.Where("1 = 1").Delete(&model.Category{})
	config.DB.Where("1 = 1").Delete(&model.User{})

	userID := uuid.New().String()
	config.DB.Create(&model.User{
		ID:       userID,
		Name:     "Repo Test User",
		Email:    "repo@spendwise.com",
		Password: "hashed",
	})
	return userID
}

func TestCategoryRepository_SyncAndReassign(t *testing.T) {
	userID := setupRepoTestDB()
	catRepo := NewCategoryRepository(config.DB)
	txnRepo := NewTransactionRepository(config.DB)

	cat1 := model.Category{UserID: userID, Name: "Cat 1", Type: "expense"}
	cat2 := model.Category{UserID: userID, Name: "Cat 2", Type: "expense"}
	catRepo.Create(&cat1)
	catRepo.Create(&cat2)

	found, err := catRepo.FindByNameAndType("Cat 1", "expense", userID)
	if err != nil || found == nil {
		t.Fatalf("FindByNameAndType failed: %v", err)
	}

	allSync, err := catRepo.FindAllForSync(userID)
	if err != nil || len(allSync) < 2 {
		t.Fatalf("FindAllForSync failed: %v", err)
	}

	tx := model.Transaction{UserID: userID, Type: "expense", Amount: 5000, CategoryID: cat1.ID, OccurredAt: "2026-08-21"}
	txnRepo.Create(&tx)

	count, err := catRepo.CountTransactionsByCategoryID(cat1.ID, userID)
	if err != nil || count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	err = catRepo.ReassignTransactions(cat1.ID, cat2.ID, userID)
	if err != nil {
		t.Fatalf("ReassignTransactions failed: %v", err)
	}

	count1After, _ := catRepo.CountTransactionsByCategoryID(cat1.ID, userID)
	count2After, _ := catRepo.CountTransactionsByCategoryID(cat2.ID, userID)
	if count1After != 0 || count2After != 1 {
		t.Errorf("Expected 0 and 1 after reassign, got %d and %d", count1After, count2After)
	}
}

func TestBudgetRepository_GetSpentAmountAndPeriods(t *testing.T) {
	userID := setupRepoTestDB()
	budgetRepo := NewBudgetRepository(config.DB)
	catRepo := NewCategoryRepository(config.DB)
	txnRepo := NewTransactionRepository(config.DB)

	cat := model.Category{UserID: userID, Name: "Food", Type: "expense"}
	catRepo.Create(&cat)

	today := time.Now().Format("2006-01-02")
	txnRepo.Create(&model.Transaction{UserID: userID, Type: "expense", Amount: 10000, CategoryID: cat.ID, OccurredAt: today})

	b := model.Budget{UserID: userID, Name: "Daily Budget", Amount: 50000, Period: "daily", CategoryID: cat.ID}
	budgetRepo.Create(&b)

	bWeekly := model.Budget{UserID: userID, Name: "Weekly Budget", Amount: 200000, Period: "weekly", CategoryID: cat.ID}
	budgetRepo.Create(&bWeekly)

	bYearly := model.Budget{UserID: userID, Name: "Yearly Budget", Amount: 10000000, Period: "yearly", CategoryID: cat.ID}
	budgetRepo.Create(&bYearly)

	spentDaily, _ := budgetRepo.GetSpentAmount(userID, cat.ID, "daily")
	if spentDaily != 10000 {
		t.Errorf("Expected spentDaily 10000, got %d", spentDaily)
	}

	spentWeekly, _ := budgetRepo.GetSpentAmount(userID, cat.ID, "weekly")
	if spentWeekly != 10000 {
		t.Errorf("Expected spentWeekly 10000, got %d", spentWeekly)
	}

	spentYearly, _ := budgetRepo.GetSpentAmount(userID, cat.ID, "yearly")
	if spentYearly != 10000 {
		t.Errorf("Expected spentYearly 10000, got %d", spentYearly)
	}
}

func TestGoalRepository_UpdateStatusAndProgress(t *testing.T) {
	userID := setupRepoTestDB()
	goalRepo := NewGoalRepository(config.DB)

	g := model.SavingGoal{UserID: userID, Name: "Goal 1", TargetAmount: 100000, CurrentSaved: 50000, Status: "active"}
	goalRepo.Create(&g)

	err := goalRepo.UpdateStatus(g.ID, userID, "archived")
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	found, _ := goalRepo.FindByID(g.ID, userID)
	if found.Status != "archived" {
		t.Errorf("Expected status archived, got %s", found.Status)
	}

	progress := CalculateGoalProgress(50000, 100000)
	if progress != 50.0 {
		t.Errorf("Expected progress 50.0, got %f", progress)
	}

	progressZero := CalculateGoalProgress(50000, 0)
	if progressZero != 0.0 {
		t.Errorf("Expected progress 0.0 for zero target, got %f", progressZero)
	}
}

func TestTransactionRepository_BatchCreateAndCalculations(t *testing.T) {
	userID := setupRepoTestDB()
	txnRepo := NewTransactionRepository(config.DB)
	catRepo := NewCategoryRepository(config.DB)

	cat := model.Category{UserID: userID, Name: "Shopping", Type: "expense"}
	catRepo.Create(&cat)

	txs := []model.Transaction{
		{UserID: userID, Type: "expense", Amount: 5000, CategoryID: cat.ID, OccurredAt: "2026-08-21"},
		{UserID: userID, Type: "expense", Amount: 15000, CategoryID: cat.ID, OccurredAt: "2026-08-21"},
	}

	_, _, err := txnRepo.BatchCreate(txs)
	if err != nil {
		t.Fatalf("BatchCreate failed: %v", err)
	}

	pages := CalculateTotalPages(100, 20)
	if pages != 5 {
		t.Errorf("Expected 5 pages, got %d", pages)
	}

	pagesZero := CalculateTotalPages(0, 20)
	if pagesZero != 0 {
		t.Errorf("Expected 0 pages, got %d", pagesZero)
	}
}

func TestRepository_FindAllAndUpdates(t *testing.T) {
	userID := setupRepoTestDB()
	catRepo := NewCategoryRepository(config.DB)
	txnRepo := NewTransactionRepository(config.DB)
	goalRepo := NewGoalRepository(config.DB)
	budgetRepo := NewBudgetRepository(config.DB)

	cat := model.Category{UserID: userID, Name: "Cat All", Type: "expense", Icon: "icon", Color: "#fff"}
	catRepo.Create(&cat)

	allCats, err := catRepo.FindAll(userID, "expense")
	if err != nil || len(allCats) == 0 {
		t.Fatalf("Cat FindAll failed: %v", err)
	}

	cat.Name = "Cat All Updated"
	catRepo.Update(&cat)
	catRepo.Delete(cat.ID, userID)

	tx := model.Transaction{UserID: userID, Type: "expense", Amount: 20000, CategoryID: cat.ID, OccurredAt: "2026-08-21", Note: "SearchNote"}
	txnRepo.Create(&tx)

	allTxns, total, err := txnRepo.FindAll(TransactionFilter{UserID: userID, Search: "SearchNote", Sort: "date_asc"})
	if err != nil || total == 0 || len(allTxns) == 0 {
		t.Fatalf("Txn FindAll failed: %v", err)
	}

	txnRepo.FindAll(TransactionFilter{UserID: userID, Type: "expense", CategoryID: cat.ID, StartDate: "2026-08-01", EndDate: "2026-08-31", Sort: "amount_desc"})
	txnRepo.FindAll(TransactionFilter{UserID: userID, Sort: "amount_asc"})

	goal := model.SavingGoal{UserID: userID, Name: "Goal All", TargetAmount: 50000, CurrentSaved: 10000, TargetDate: "2027-01-01", Status: "active"}
	goalRepo.Create(&goal)

	allGoals, err := goalRepo.FindAll(userID, "active")
	if err != nil || len(allGoals) == 0 {
		t.Fatalf("Goal FindAll failed: %v", err)
	}

	goal.Name = "Goal All Updated"
	goalRepo.Update(&goal)

	contrib := model.GoalContribution{GoalID: goal.ID, Amount: 5000, Date: "2026-08-21", Note: "Contrib"}
	goalRepo.CreateContribution(&contrib)

	contribs, count, err := goalRepo.FindContributionsByGoalID(goal.ID, 1, 10)
	if err != nil || count == 0 || len(contribs) == 0 {
		t.Fatalf("FindContributionsByGoalID failed: %v", err)
	}

	goalRepo.Delete(goal.ID, userID)

	b := model.Budget{UserID: userID, Name: "Budget All", Amount: 100000, Period: "monthly", CategoryID: cat.ID}
	budgetRepo.Create(&b)

	allBudgets, err := budgetRepo.FindAll(userID, "monthly")
	if err != nil || len(allBudgets) == 0 {
		t.Fatalf("Budget FindAll failed: %v", err)
	}

	b.Name = "Budget All Updated"
	budgetRepo.Update(&b)
	budgetRepo.Delete(b.ID, userID)
}
