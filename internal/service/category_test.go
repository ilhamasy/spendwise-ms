package service

import (
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"
	"spendwise-ms/internal/repository"

	"github.com/google/uuid"
)

func setupCategoryTestDB() {
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
}

func createUserForCategory() string {
	userID := uuid.New().String()
	config.DB.Create(&model.User{
		ID:       userID,
		Name:     "Cat Test User",
		Email:    "cattest@spendwise.com",
		Password: "hashed",
	})
	return userID
}

func TestListCategories_Empty(t *testing.T) {
	setupCategoryTestDB()
	userID := createUserForCategory()
	svc := NewCategoryService(config.DB)

	resp, err := svc.ListCategories(userID, "")
	if err != nil {
		t.Fatalf("ListCategories failed: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("Expected 0 categories, got %d", len(resp))
	}
}

func TestCreateCategory_Success(t *testing.T) {
	setupCategoryTestDB()
	userID := createUserForCategory()
	svc := NewCategoryService(config.DB)

	resp, err := svc.CreateCategory(userID, dto.CreateCategoryRequest{
		Name:  "Food",
		Type:  "expense",
		Icon:  "food",
		Color: "#FF0000",
	})
	if err != nil {
		t.Fatalf("CreateCategory failed: %v", err)
	}
	if resp.Name != "Food" {
		t.Errorf("Expected name 'Food', got '%s'", resp.Name)
	}
	if resp.Type != "expense" {
		t.Errorf("Expected type 'expense', got '%s'", resp.Type)
	}
	if resp.IsDefault {
		t.Error("User-created category should not be default")
	}
}

func TestCreateCategory_DuplicateName(t *testing.T) {
	setupCategoryTestDB()
	userID := createUserForCategory()
	svc := NewCategoryService(config.DB)

	svc.CreateCategory(userID, dto.CreateCategoryRequest{
		Name: "Food", Type: "expense", Icon: "food",
	})

	_, err := svc.CreateCategory(userID, dto.CreateCategoryRequest{
		Name: "Food", Type: "expense", Icon: "restaurant",
	})
	if err == nil {
		t.Error("Expected duplicate name error")
	}
}

func TestCreateCategory_DifferentTypeSameName(t *testing.T) {
	setupCategoryTestDB()
	userID := createUserForCategory()
	svc := NewCategoryService(config.DB)

	_, err := svc.CreateCategory(userID, dto.CreateCategoryRequest{
		Name: "Salary", Type: "income", Icon: "money",
	})
	if err != nil {
		t.Fatalf("CreateCategory failed: %v", err)
	}

	resp, err := svc.CreateCategory(userID, dto.CreateCategoryRequest{
		Name: "Salary", Type: "expense", Icon: "money",
	})
	if err != nil {
		t.Fatalf("Should allow same name with different type: %v", err)
	}
	if resp.Type != "expense" {
		t.Errorf("Expected type 'expense', got '%s'", resp.Type)
	}
}

func TestListCategories_FilterByType(t *testing.T) {
	setupCategoryTestDB()
	userID := createUserForCategory()
	svc := NewCategoryService(config.DB)

	svc.CreateCategory(userID, dto.CreateCategoryRequest{Name: "Food", Type: "expense"})
	svc.CreateCategory(userID, dto.CreateCategoryRequest{Name: "Transport", Type: "expense"})
	svc.CreateCategory(userID, dto.CreateCategoryRequest{Name: "Salary", Type: "income"})

	resp, err := svc.ListCategories(userID, "income")
	if err != nil {
		t.Fatalf("ListCategories failed: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("Expected 1 income category, got %d", len(resp))
	}
	if resp[0].Type != "income" {
		t.Errorf("Expected type 'income', got '%s'", resp[0].Type)
	}
}

func TestUpdateCategory_Success(t *testing.T) {
	setupCategoryTestDB()
	userID := createUserForCategory()
	svc := NewCategoryService(config.DB)

	created, _ := svc.CreateCategory(userID, dto.CreateCategoryRequest{
		Name: "Food", Type: "expense", Icon: "food", Color: "#FF0000",
	})

	resp, err := svc.UpdateCategory(userID, created.ID, dto.UpdateCategoryRequest{
		Name:  "Groceries",
		Icon:  "cart",
		Color: "#00FF00",
	})
	if err != nil {
		t.Fatalf("UpdateCategory failed: %v", err)
	}
	if resp.Name != "Groceries" {
		t.Errorf("Expected name 'Groceries', got '%s'", resp.Name)
	}
	if resp.Icon != "cart" {
		t.Errorf("Expected icon 'cart', got '%s'", resp.Icon)
	}
	if resp.Type != "expense" {
		t.Errorf("Type should be preserved as 'expense', got '%s'", resp.Type)
	}
}

func TestDeleteCategory_WithoutTransactions(t *testing.T) {
	setupCategoryTestDB()
	userID := createUserForCategory()
	svc := NewCategoryService(config.DB)

	created, _ := svc.CreateCategory(userID, dto.CreateCategoryRequest{
		Name: "DeleteMe", Type: "expense",
	})

	err := svc.DeleteCategory(userID, created.ID, "")
	if err != nil {
		t.Fatalf("DeleteCategory failed: %v", err)
	}

	all, _ := svc.ListCategories(userID, "")
	for _, cat := range all {
		if cat.ID == created.ID {
			t.Error("Category should have been deleted")
		}
	}
}

func TestDeleteCategory_WithTransactionsNoReassign(t *testing.T) {
	setupCategoryTestDB()
	userID := createUserForCategory()
	catSvc := NewCategoryService(config.DB)

	cat, _ := catSvc.CreateCategory(userID, dto.CreateCategoryRequest{
		Name: "InUse", Type: "expense",
	})

	txnRepo := repository.NewTransactionRepository(config.DB)
	txnRepo.Create(&model.Transaction{
		UserID:     userID,
		Type:       "expense",
		Amount:     10000,
		CategoryID: cat.ID,
		OccurredAt: "2026-06-01",
	})

	err := catSvc.DeleteCategory(userID, cat.ID, "")
	if err == nil {
		t.Error("Expected error when deleting category with transactions")
	}
}

func TestDeleteCategory_WithTransactionsReassign(t *testing.T) {
	setupCategoryTestDB()
	userID := createUserForCategory()
	catSvc := NewCategoryService(config.DB)

	inUseCat, _ := catSvc.CreateCategory(userID, dto.CreateCategoryRequest{
		Name: "OldCategory", Type: "expense",
	})
	reassignCat, _ := catSvc.CreateCategory(userID, dto.CreateCategoryRequest{
		Name: "NewCategory", Type: "expense",
	})

	txnRepo := repository.NewTransactionRepository(config.DB)
	txnRepo.Create(&model.Transaction{
		UserID:     userID,
		Type:       "expense",
		Amount:     10000,
		CategoryID: inUseCat.ID,
		OccurredAt: "2026-06-01",
	})

	err := catSvc.DeleteCategory(userID, inUseCat.ID, reassignCat.ID)
	if err != nil {
		t.Fatalf("Expected successful reassign+delete: %v", err)
	}
}
