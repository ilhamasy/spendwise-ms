package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/middleware"
	"spendwise-ms/internal/model"
	"spendwise-ms/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func setupTransactionTestDB() {
	config.DB.Where("1 = 1").Delete(&model.Transaction{})
	config.DB.Where("1 = 1").Delete(&model.Category{})
	config.DB.Where("1 = 1").Delete(&model.User{})
}

func setupTransactionRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	auth := r.Group("/api/auth")
	{
		auth.POST("/register", Register)
		auth.POST("/login", Login)
	}

	transactions := r.Group("/api/transactions")
	transactions.Use(middleware.AuthRequired())
	{
		transactions.GET("", GetTransactions)
		transactions.POST("", CreateTransaction)
		transactions.GET("/:id", GetTransactionByID)
		transactions.PUT("/:id", UpdateTransaction)
		transactions.DELETE("/:id", DeleteTransaction)
		transactions.POST("/sync", SyncTransactions)
	}

	return r
}

func createTestUserAndToken(t *testing.T) (string, string) {
	t.Helper()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	userID := uuid.New().String()
	config.DB.Create(&model.User{
		ID:       userID,
		Name:     "Txn Test User",
		Email:    "txntest@spendwise.com",
		Password: string(hash),
	})

	token, _, _ := service.GenerateTokens(userID, "txntest@spendwise.com")
	return token, userID
}

func createTestCategoryForHandler(userID string) *model.Category {
	cat := &model.Category{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      "Test Category",
		Type:      "expense",
		Icon:      "test",
		Color:     "#000000",
		IsDefault: false,
	}
	config.DB.Create(cat)
	return cat
}

func TestCreateTransactionHandler_Success(t *testing.T) {
	setupTransactionTestDB()
	r := setupTransactionRouter()
	token, userID := createTestUserAndToken(t)
	cat := createTestCategoryForHandler(userID)

	body, _ := json.Marshal(dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     50000,
		CategoryID: cat.ID,
		OccurredAt: time.Now().Format("2006-01-02"),
		Note:       "Test transaction",
	})

	req, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp dto.TransactionResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.ID == "" {
		t.Error("Transaction ID should not be empty")
	}
	if resp.Type != "expense" {
		t.Errorf("Expected type 'expense', got '%s'", resp.Type)
	}
	if resp.Amount != 50000 {
		t.Errorf("Expected amount 50000, got %d", resp.Amount)
	}
}

func TestCreateTransactionHandler_Unauthenticated(t *testing.T) {
	setupTransactionTestDB()
	r := setupTransactionRouter()

	body, _ := json.Marshal(dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     50000,
		CategoryID: "some-category",
		OccurredAt: time.Now().Format("2006-01-02"),
	})

	req, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestCreateTransactionHandler_ValidationError(t *testing.T) {
	setupTransactionTestDB()
	r := setupTransactionRouter()
	token, _ := createTestUserAndToken(t)

	body, _ := json.Marshal(dto.CreateTransactionRequest{
		Type:       "invalid",
		Amount:     0,
		CategoryID: "",
		OccurredAt: "",
	})

	req, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetTransactionsHandler_Success(t *testing.T) {
	setupTransactionTestDB()
	r := setupTransactionRouter()
	token, userID := createTestUserAndToken(t)
	cat := createTestCategoryForHandler(userID)

	createReq, _ := json.Marshal(dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     30000,
		CategoryID: cat.ID,
		OccurredAt: time.Now().Format("2006-01-02"),
	})

	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBuffer(createReq))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	req, _ := http.NewRequest("GET", "/api/transactions?page=1&limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp dto.TransactionListResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if len(resp.Data) != 3 {
		t.Errorf("Expected 3 transactions, got %d", len(resp.Data))
	}
	if resp.Meta.Total != 3 {
		t.Errorf("Expected total 3, got %d", resp.Meta.Total)
	}
}

func TestGetTransactionsHandler_FilterByType(t *testing.T) {
	setupTransactionTestDB()
	r := setupTransactionRouter()
	token, userID := createTestUserAndToken(t)
	cat := createTestCategoryForHandler(userID)

	expenseBody, _ := json.Marshal(dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     20000,
		CategoryID: cat.ID,
		OccurredAt: time.Now().Format("2006-01-02"),
	})
	incomeBody, _ := json.Marshal(dto.CreateTransactionRequest{
		Type:       "income",
		Amount:     100000,
		CategoryID: cat.ID,
		OccurredAt: time.Now().Format("2006-01-02"),
	})

	for _, b := range [][]byte{expenseBody, expenseBody, incomeBody} {
		req, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBuffer(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	req, _ := http.NewRequest("GET", "/api/transactions?type=income", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp dto.TransactionListResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Data) != 1 {
		t.Errorf("Expected 1 income transaction, got %d", len(resp.Data))
	}
	if resp.Data[0].Type != "income" {
		t.Errorf("Expected type 'income', got '%s'", resp.Data[0].Type)
	}
}

func TestGetTransactionByIDHandler_Success(t *testing.T) {
	setupTransactionTestDB()
	r := setupTransactionRouter()
	token, userID := createTestUserAndToken(t)
	cat := createTestCategoryForHandler(userID)

	body, _ := json.Marshal(dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     40000,
		CategoryID: cat.ID,
		OccurredAt: time.Now().Format("2006-01-02"),
	})

	req, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var created dto.TransactionResponse
	json.Unmarshal(w.Body.Bytes(), &created)

	req2, _ := http.NewRequest("GET", "/api/transactions/"+created.ID, nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestGetTransactionByIDHandler_NotFound(t *testing.T) {
	setupTransactionTestDB()
	r := setupTransactionRouter()
	token, _ := createTestUserAndToken(t)

	req, _ := http.NewRequest("GET", "/api/transactions/non-existent-id", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestUpdateTransactionHandler_Success(t *testing.T) {
	setupTransactionTestDB()
	r := setupTransactionRouter()
	token, userID := createTestUserAndToken(t)
	cat := createTestCategoryForHandler(userID)

	body, _ := json.Marshal(dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     10000,
		CategoryID: cat.ID,
		OccurredAt: time.Now().Format("2006-01-02"),
		Note:       "Original",
	})

	req, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var created dto.TransactionResponse
	json.Unmarshal(w.Body.Bytes(), &created)

	updateBody, _ := json.Marshal(dto.UpdateTransactionRequest{
		Amount: 25000,
		Note:   "Updated",
	})
	req2, _ := http.NewRequest("PUT", "/api/transactions/"+created.ID, bytes.NewBuffer(updateBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var updated dto.TransactionResponse
	json.Unmarshal(w2.Body.Bytes(), &updated)
	if updated.Amount != 25000 {
		t.Errorf("Expected amount 25000, got %d", updated.Amount)
	}
	if updated.Note != "Updated" {
		t.Errorf("Expected note 'Updated', got '%s'", updated.Note)
	}
}

func TestDeleteTransactionHandler_Success(t *testing.T) {
	setupTransactionTestDB()
	r := setupTransactionRouter()
	token, userID := createTestUserAndToken(t)
	cat := createTestCategoryForHandler(userID)

	body, _ := json.Marshal(dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     10000,
		CategoryID: cat.ID,
		OccurredAt: time.Now().Format("2006-01-02"),
	})

	req, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var created dto.TransactionResponse
	json.Unmarshal(w.Body.Bytes(), &created)

	req2, _ := http.NewRequest("DELETE", "/api/transactions/"+created.ID, nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w2.Code)
	}
}

func TestSyncTransactionsHandler_Success(t *testing.T) {
	setupTransactionTestDB()
	r := setupTransactionRouter()
	token, userID := createTestUserAndToken(t)
	cat := createTestCategoryForHandler(userID)

	body, _ := json.Marshal(dto.SyncRequest{
		Transactions: []dto.SyncTransactionRequest{
			{
				Type:       "expense",
				Amount:     15000,
				CategoryID: cat.ID,
				OccurredAt: time.Now().Format("2006-01-02"),
				Note:       "Sync test 1",
			},
			{
				Type:       "income",
				Amount:     75000,
				CategoryID: cat.ID,
				OccurredAt: time.Now().Format("2006-01-02"),
				Note:       "Sync test 2",
			},
		},
	})

	req, _ := http.NewRequest("POST", "/api/transactions/sync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp dto.SyncResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Synced != 2 {
		t.Errorf("Expected 2 synced, got %d", resp.Synced)
	}
	if len(resp.Failed) != 0 {
		t.Errorf("Expected 0 failed, got %d", len(resp.Failed))
	}
}

func TestSyncTransactionsHandler_PartialFailure(t *testing.T) {
	setupTransactionTestDB()
	r := setupTransactionRouter()
	token, userID := createTestUserAndToken(t)
	cat := createTestCategoryForHandler(userID)

	body, _ := json.Marshal(dto.SyncRequest{
		Transactions: []dto.SyncTransactionRequest{
			{
				Type:       "expense",
				Amount:     10000,
				CategoryID: cat.ID,
				OccurredAt: time.Now().Format("2006-01-02"),
			},
			{
				Type:       "invalid_type",
				Amount:     50000,
				CategoryID: cat.ID,
				OccurredAt: time.Now().Format("2006-01-02"),
			},
		},
	})

	req, _ := http.NewRequest("POST", "/api/transactions/sync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp dto.SyncResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Synced != 1 {
		t.Errorf("Expected 1 synced, got %d", resp.Synced)
	}
	if len(resp.Failed) != 1 {
		t.Errorf("Expected 1 failed, got %d", len(resp.Failed))
	}
}

func TestSearchTransactionsHandler(t *testing.T) {
	setupTransactionTestDB()
	r := setupTransactionRouter()
	token, userID := createTestUserAndToken(t)
	cat := createTestCategoryForHandler(userID)

	s1, _ := json.Marshal(dto.CreateTransactionRequest{
		Type: "expense", Amount: 50000, CategoryID: cat.ID,
		OccurredAt: time.Now().Format("2006-01-02"), Note: "Coffee at cafe",
	})
	s2, _ := json.Marshal(dto.CreateTransactionRequest{
		Type: "expense", Amount: 30000, CategoryID: cat.ID,
		OccurredAt: time.Now().Format("2006-01-02"), Note: "Buy groceries",
	})

	for _, b := range [][]byte{s1, s2} {
		req, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBuffer(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	req, _ := http.NewRequest("GET", "/api/transactions?search=coffee", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp dto.TransactionListResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Data) != 1 {
		t.Errorf("Expected 1 search result for 'coffee', got %d", len(resp.Data))
	}
}
