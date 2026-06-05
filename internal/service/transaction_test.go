package service

import (
	"testing"
	"time"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"
	"spendwise-ms/internal/repository"

	"github.com/google/uuid"
)

func setupTransactionTestDB() {
	if config.DB == nil {
		config.InitDB(config.Load())
	}
	config.AutoMigrate(&model.Transaction{}, &model.Category{}, &model.User{})
	config.DB.Where("1 = 1").Delete(&model.Transaction{})
	config.DB.Where("1 = 1").Delete(&model.Category{})
}

func createTestCategory() *model.Category {
	cat := &model.Category{
		ID:        uuid.New().String(),
		UserID:    "test-user-1",
		Name:      "Food",
		Type:      "expense",
		Icon:      "food",
		Color:     "#FF0000",
		IsDefault: false,
	}
	config.DB.Create(cat)
	return cat
}

func createDefaultCategory() *model.Category {
	cat := &model.Category{
		ID:        uuid.New().String(),
		UserID:    "",
		Name:      "Default Category",
		Type:      "expense",
		Icon:      "default",
		Color:     "#CCCCCC",
		IsDefault: true,
	}
	config.DB.Create(cat)
	return cat
}

func TestCreateTransaction_Success(t *testing.T) {
	setupTransactionTestDB()
	cat := createTestCategory()
	svc := NewTransactionService(config.DB)

	resp, err := svc.CreateTransaction("test-user-1", dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     50000,
		CategoryID: cat.ID,
		OccurredAt: time.Now().Format("2006-01-02"),
		Note:       "Lunch",
	})
	if err != nil {
		t.Fatalf("CreateTransaction failed: %v", err)
	}
	if resp.ID == "" {
		t.Error("Transaction ID should not be empty")
	}
	if resp.Type != "expense" {
		t.Errorf("Expected type 'expense', got '%s'", resp.Type)
	}
	if resp.Amount != 50000 {
		t.Errorf("Expected amount 50000, got %d", resp.Amount)
	}
	if resp.CategoryID != cat.ID {
		t.Errorf("Expected categoryId '%s', got '%s'", cat.ID, resp.CategoryID)
	}
	if resp.Note != "Lunch" {
		t.Errorf("Expected note 'Lunch', got '%s'", resp.Note)
	}
}

func TestCreateTransaction_InvalidType(t *testing.T) {
	setupTransactionTestDB()
	cat := createTestCategory()
	svc := NewTransactionService(config.DB)

	_, err := svc.CreateTransaction("test-user-1", dto.CreateTransactionRequest{
		Type:       "invalid",
		Amount:     50000,
		CategoryID: cat.ID,
		OccurredAt: time.Now().Format("2006-01-02"),
	})
	if err == nil {
		t.Error("Expected validation error for invalid type")
	}
}

func TestCreateTransaction_ZeroAmount(t *testing.T) {
	setupTransactionTestDB()
	cat := createTestCategory()
	svc := NewTransactionService(config.DB)

	_, err := svc.CreateTransaction("test-user-1", dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     0,
		CategoryID: cat.ID,
		OccurredAt: time.Now().Format("2006-01-02"),
	})
	if err == nil {
		t.Error("Expected validation error for zero amount")
	}
}

func TestCreateTransaction_InvalidCategory(t *testing.T) {
	setupTransactionTestDB()
	svc := NewTransactionService(config.DB)

	_, err := svc.CreateTransaction("test-user-1", dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     50000,
		CategoryID: "non-existent-category-id",
		OccurredAt: time.Now().Format("2006-01-02"),
	})
	if err == nil {
		t.Error("Expected error for invalid category")
	}
}

func TestCreateTransaction_FutureDate(t *testing.T) {
	setupTransactionTestDB()
	cat := createTestCategory()
	svc := NewTransactionService(config.DB)

	futureDate := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	_, err := svc.CreateTransaction("test-user-1", dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     50000,
		CategoryID: cat.ID,
		OccurredAt: futureDate,
	})
	if err == nil {
		t.Error("Expected error for future date")
	}
}

func TestGetTransactions_Pagination(t *testing.T) {
	setupTransactionTestDB()
	cat := createTestCategory()
	svc := NewTransactionService(config.DB)

	for i := 0; i < 5; i++ {
		svc.CreateTransaction("test-user-1", dto.CreateTransactionRequest{
			Type:       "expense",
			Amount:     int64(10000 * (i + 1)),
			CategoryID: cat.ID,
			OccurredAt: time.Now().Format("2006-01-02"),
			Note:       "Note",
		})
	}

	resp, err := svc.GetTransactions("test-user-1", repository.TransactionFilter{
		Page:  1,
		Limit: 3,
	})
	if err != nil {
		t.Fatalf("GetTransactions failed: %v", err)
	}
	if len(resp.Data) != 3 {
		t.Errorf("Expected 3 items on first page, got %d", len(resp.Data))
	}
	if resp.Meta.Total != 5 {
		t.Errorf("Expected total 5, got %d", resp.Meta.Total)
	}
	if resp.Meta.Page != 1 {
		t.Errorf("Expected page 1, got %d", resp.Meta.Page)
	}
	if resp.Meta.Limit != 3 {
		t.Errorf("Expected limit 3, got %d", resp.Meta.Limit)
	}
	if resp.Meta.TotalPages != 2 {
		t.Errorf("Expected totalPages 2, got %d", resp.Meta.TotalPages)
	}
}

func TestGetTransactions_FilterByType(t *testing.T) {
	setupTransactionTestDB()
	cat := createTestCategory()
	svc := NewTransactionService(config.DB)

	svc.CreateTransaction("test-user-1", dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     10000,
		CategoryID: cat.ID,
		OccurredAt: time.Now().Format("2006-01-02"),
	})
	svc.CreateTransaction("test-user-1", dto.CreateTransactionRequest{
		Type:       "income",
		Amount:     50000,
		CategoryID: cat.ID,
		OccurredAt: time.Now().Format("2006-01-02"),
	})

	resp, err := svc.GetTransactions("test-user-1", repository.TransactionFilter{
		Type:  "income",
		Page:  1,
		Limit: 20,
	})
	if err != nil {
		t.Fatalf("GetTransactions failed: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Errorf("Expected 1 income transaction, got %d", len(resp.Data))
	}
	if resp.Data[0].Type != "income" {
		t.Errorf("Expected type 'income', got '%s'", resp.Data[0].Type)
	}
}

func TestGetTransactions_Search(t *testing.T) {
	setupTransactionTestDB()
	cat := createTestCategory()
	svc := NewTransactionService(config.DB)

	svc.CreateTransaction("test-user-1", dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     12000,
		CategoryID: cat.ID,
		OccurredAt: time.Now().Format("2006-01-02"),
		Note:       "Starbucks coffee",
	})
	svc.CreateTransaction("test-user-1", dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     25000,
		CategoryID: cat.ID,
		OccurredAt: time.Now().Format("2006-01-02"),
		Note:       "Grocery shopping",
	})

	resp, err := svc.GetTransactions("test-user-1", repository.TransactionFilter{
		Search: "coffee",
		Page:   1,
		Limit:  20,
	})
	if err != nil {
		t.Fatalf("GetTransactions failed: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Errorf("Expected 1 search result for 'coffee', got %d", len(resp.Data))
	}
}

func TestGetTransactionByID_Success(t *testing.T) {
	setupTransactionTestDB()
	cat := createTestCategory()
	svc := NewTransactionService(config.DB)

	created, _ := svc.CreateTransaction("test-user-1", dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     30000,
		CategoryID: cat.ID,
		OccurredAt: time.Now().Format("2006-01-02"),
	})

	resp, err := svc.GetTransactionByID("test-user-1", created.ID)
	if err != nil {
		t.Fatalf("GetTransactionByID failed: %v", err)
	}
	if resp.ID != created.ID {
		t.Errorf("Expected ID '%s', got '%s'", created.ID, resp.ID)
	}
}

func TestGetTransactionByID_NotFound(t *testing.T) {
	setupTransactionTestDB()
	svc := NewTransactionService(config.DB)

	_, err := svc.GetTransactionByID("test-user-1", "non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent transaction")
	}
}

func TestUpdateTransaction_Success(t *testing.T) {
	setupTransactionTestDB()
	cat := createTestCategory()
	svc := NewTransactionService(config.DB)

	created, _ := svc.CreateTransaction("test-user-1", dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     10000,
		CategoryID: cat.ID,
		OccurredAt: time.Now().Format("2006-01-02"),
		Note:       "Old note",
	})

	resp, err := svc.UpdateTransaction("test-user-1", created.ID, dto.UpdateTransactionRequest{
		Amount: 20000,
		Note:   "Updated note",
	})
	if err != nil {
		t.Fatalf("UpdateTransaction failed: %v", err)
	}
	if resp.Amount != 20000 {
		t.Errorf("Expected amount 20000, got %d", resp.Amount)
	}
	if resp.Note != "Updated note" {
		t.Errorf("Expected note 'Updated note', got '%s'", resp.Note)
	}
}

func TestDeleteTransaction_Success(t *testing.T) {
	setupTransactionTestDB()
	cat := createTestCategory()
	svc := NewTransactionService(config.DB)

	created, _ := svc.CreateTransaction("test-user-1", dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     10000,
		CategoryID: cat.ID,
		OccurredAt: time.Now().Format("2006-01-02"),
	})

	err := svc.DeleteTransaction("test-user-1", created.ID)
	if err != nil {
		t.Fatalf("DeleteTransaction failed: %v", err)
	}

	_, err = svc.GetTransactionByID("test-user-1", created.ID)
	if err == nil {
		t.Error("Expected transaction to be deleted")
	}
}

func TestSyncTransactions_Success(t *testing.T) {
	setupTransactionTestDB()
	cat := createTestCategory()
	svc := NewTransactionService(config.DB)

	resp, err := svc.SyncTransactions("test-user-1", dto.SyncRequest{
		Transactions: []dto.SyncTransactionRequest{
			{
				Type:       "expense",
				Amount:     10000,
				CategoryID: cat.ID,
				OccurredAt: time.Now().Format("2006-01-02"),
				Note:       "Synced txn 1",
			},
			{
				Type:       "income",
				Amount:     50000,
				CategoryID: cat.ID,
				OccurredAt: time.Now().Format("2006-01-02"),
				Note:       "Synced txn 2",
			},
		},
	})
	if err != nil {
		t.Fatalf("SyncTransactions failed: %v", err)
	}
	if resp.Synced != 2 {
		t.Errorf("Expected 2 synced, got %d", resp.Synced)
	}
	if len(resp.Failed) != 0 {
		t.Errorf("Expected 0 failed, got %d", len(resp.Failed))
	}
}

func TestSyncTransactions_PartialFailure(t *testing.T) {
	setupTransactionTestDB()
	cat := createTestCategory()
	svc := NewTransactionService(config.DB)

	resp, err := svc.SyncTransactions("test-user-1", dto.SyncRequest{
		Transactions: []dto.SyncTransactionRequest{
			{
				Type:       "expense",
				Amount:     10000,
				CategoryID: cat.ID,
				OccurredAt: time.Now().Format("2006-01-02"),
			},
			{
				Type:       "invalid_type",
				Amount:     20000,
				CategoryID: cat.ID,
				OccurredAt: time.Now().Format("2006-01-02"),
			},
		},
	})
	if err != nil {
		t.Fatalf("SyncTransactions failed: %v", err)
	}
	if resp.Synced != 1 {
		t.Errorf("Expected 1 synced, got %d", resp.Synced)
	}
	if len(resp.Failed) != 1 {
		t.Errorf("Expected 1 failed, got %d", len(resp.Failed))
	}
}

func TestDefaultCategoryAccessible(t *testing.T) {
	setupTransactionTestDB()
	cat := createDefaultCategory()
	svc := NewTransactionService(config.DB)

	resp, err := svc.CreateTransaction("other-user", dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     50000,
		CategoryID: cat.ID,
		OccurredAt: time.Now().Format("2006-01-02"),
	})
	if err != nil {
		t.Fatalf("Should allow default category for any user: %v", err)
	}
	if resp.ID == "" {
		t.Error("Transaction ID should not be empty")
	}
}

func TestCalculateTotalPages(t *testing.T) {
	if tp := repository.CalculateTotalPages(20, 10); tp != 2 {
		t.Errorf("Expected 2 total pages for 20 items with limit 10, got %d", tp)
	}
	if tp := repository.CalculateTotalPages(21, 10); tp != 3 {
		t.Errorf("Expected 3 total pages for 21 items with limit 10, got %d", tp)
	}
	if tp := repository.CalculateTotalPages(0, 10); tp != 0 {
		t.Errorf("Expected 0 total pages for 0 items, got %d", tp)
	}
}
