package handler

import (
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"
	"spendwise-ms/internal/repository"
	"spendwise-ms/internal/service"

	"github.com/google/uuid"
)

func TestInjection_SQLiSearchSafety(t *testing.T) {
	setupTestDB()
	userID := uuid.New().String()
	config.DB.Create(&model.User{ID: userID, Name: "SQLi Test", Email: "sqli@test.com", Password: "hash"})

	catSvc := service.NewCategoryService(config.DB)
	cat, _ := catSvc.CreateCategory(userID, dto.CreateCategoryRequest{Name: "General", Type: "expense"})

	txnSvc := service.NewTransactionService(config.DB)
	txnSvc.CreateTransaction(userID, dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     50000,
		CategoryID: cat.ID,
		OccurredAt: "2026-08-21",
		Note:       "Regular dinner",
	})

	// Attempt SQL injection in search filter
	sqliPayload := "' OR '1'='1' --"
	resp, err := txnSvc.GetTransactions(userID, repository.TransactionFilter{
		Search: sqliPayload,
		Page:   1,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// SQL injection must not return all rows; it should look for literal "' OR '1'='1' --" note
	if len(resp.Data) != 0 {
		t.Errorf("SQL Injection payload returned %d results, expected 0", len(resp.Data))
	}
}

func TestInjection_XSSPayloadSanitization(t *testing.T) {
	setupTestDB()
	userID := uuid.New().String()
	config.DB.Create(&model.User{ID: userID, Name: "XSS Test", Email: "xss@test.com", Password: "hash"})

	catSvc := service.NewCategoryService(config.DB)
	catXSS := "<script>alert('xss')</script>"
	catResp, err := catSvc.CreateCategory(userID, dto.CreateCategoryRequest{
		Name: catXSS,
		Type: "expense",
		Icon: "star",
	})
	if err != nil {
		t.Fatalf("CreateCategory failed: %v", err)
	}

	if catResp.Name == catXSS {
		t.Error("Category name was not sanitized, stored XSS vulnerability present")
	}
	if catResp.Name != "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;" {
		t.Errorf("Unexpected sanitized name: %s", catResp.Name)
	}

	txnSvc := service.NewTransactionService(config.DB)
	txXSS := "<iframe src=\"javascript:alert('tx-xss')\"></iframe>"
	txResp, err := txnSvc.CreateTransaction(userID, dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     10000,
		CategoryID: catResp.ID,
		OccurredAt: "2026-08-21",
		Note:       txXSS,
	})
	if err != nil {
		t.Fatalf("CreateTransaction failed: %v", err)
	}

	if txResp.Note == txXSS {
		t.Error("Transaction note was not sanitized, stored XSS vulnerability present")
	}
	if txResp.Note != "&lt;iframe src=&#34;javascript:alert(&#39;tx-xss&#39;)&#34;&gt;&lt;/iframe&gt;" {
		t.Errorf("Unexpected sanitized note: %s", txResp.Note)
	}
}
