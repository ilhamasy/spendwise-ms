package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func setupHandlerTest(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	if config.DB == nil {
		cfg := config.Load()
		config.InitDB(cfg)
		config.AutoMigrate(
			&model.User{}, &model.Transaction{}, &model.Category{},
			&model.SavingGoal{}, &model.GoalContribution{}, &model.Budget{},
		)
	}
	config.DB.Where("1 = 1").Delete(&model.Transaction{})
	config.DB.Where("1 = 1").Delete(&model.Category{})
	config.DB.Where("1 = 1").Delete(&model.Budget{})
	config.DB.Where("1 = 1").Delete(&model.GoalContribution{})
	config.DB.Where("1 = 1").Delete(&model.SavingGoal{})
	config.DB.Where("1 = 1").Delete(&model.User{})
}

func createTestUser(t *testing.T, email string) (string, string) {
	t.Helper()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	userID := uuid.New().String()
	config.DB.Create(&model.User{
		ID:       userID,
		Name:     "Test User",
		Email:    email,
		Password: string(hash),
	})
	return userID, email
}

func getAuthToken(t *testing.T, email, password string) string {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/auth/login", Login)

	body, _ := json.Marshal(dto.LoginRequest{Email: email, Password: password})
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "spendwise-access-token" {
			return cookie.Value
		}
	}
	t.Fatal("spendwise-access-token cookie not found in login response")
	return ""
}

func TestGetCategories_Empty(t *testing.T) {
	setupHandlerTest(t)
	uid, email := createTestUser(t, "cattest@test.com")
	token := getAuthToken(t, email, "password123")

	r := gin.New()
	r.GET("/api/categories", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, GetCategories)

	req, _ := http.NewRequest("GET", "/api/categories", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestCreateCategory_Success(t *testing.T) {
	setupHandlerTest(t)
	uid, email := createTestUser(t, "catcreate@test.com")
	token := getAuthToken(t, email, "password123")

	r := gin.New()
	r.POST("/api/categories", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, CreateCategory)

	body, _ := json.Marshal(map[string]string{
		"name": "NewCat", "type": "expense", "icon": "📁", "color": "#000",
	})
	req, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetTransactions_Empty(t *testing.T) {
	setupHandlerTest(t)
	uid, email := createTestUser(t, "txempty@test.com")
	token := getAuthToken(t, email, "password123")

	r := gin.New()
	r.GET("/api/transactions", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, GetTransactions)

	req, _ := http.NewRequest("GET", "/api/transactions?page=1&limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestCreateTransaction_Success(t *testing.T) {
	setupHandlerTest(t)
	uid, email := createTestUser(t, "txcreate@test.com")
	token := getAuthToken(t, email, "password123")

	cat := &model.Category{ID: uuid.New().String(), Name: "Food", Type: "expense", Icon: "🍔", UserID: uid}
	config.DB.Create(cat)

	r := gin.New()
	r.POST("/api/transactions", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, CreateTransaction)

	body, _ := json.Marshal(dto.CreateTransactionRequest{
		Type:       "expense",
		Amount:     50000,
		CategoryID: cat.ID,
		OccurredAt: "2026-06-09",
	})
	req, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetProfile_Success(t *testing.T) {
	setupHandlerTest(t)
	uid, email := createTestUser(t, "profile@test.com")
	token := getAuthToken(t, email, "password123")

	r := gin.New()
	r.GET("/api/profile", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, GetProfile)

	req, _ := http.NewRequest("GET", "/api/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestUpdateProfile_Success(t *testing.T) {
	setupHandlerTest(t)
	uid, email := createTestUser(t, "updateprofile@test.com")
	token := getAuthToken(t, email, "password123")

	r := gin.New()
	r.PUT("/api/profile", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, UpdateProfile)

	body, _ := json.Marshal(map[string]string{"name": "Updated Name"})
	req, _ := http.NewRequest("PUT", "/api/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestGetBudgets_Empty(t *testing.T) {
	setupHandlerTest(t)
	uid, email := createTestUser(t, "budempty@test.com")
	token := getAuthToken(t, email, "password123")

	r := gin.New()
	r.GET("/api/budgets", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, GetBudgets)

	req, _ := http.NewRequest("GET", "/api/budgets", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestCreateBudget_Success(t *testing.T) {
	setupHandlerTest(t)
	uid, email := createTestUser(t, "budcreate@test.com")
	token := getAuthToken(t, email, "password123")

	cat := &model.Category{ID: uuid.New().String(), Name: "Food", Type: "expense", Icon: "🍔", UserID: uid}
	config.DB.Create(cat)

	r := gin.New()
	r.POST("/api/budgets", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, CreateBudget)

	body, _ := json.Marshal(map[string]interface{}{
		"name":       "Food Budget",
		"amount":     2000000,
		"period":     "monthly",
		"categoryId": cat.ID,
	})
	req, _ := http.NewRequest("POST", "/api/budgets", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetGoals_Empty(t *testing.T) {
	setupHandlerTest(t)
	uid, email := createTestUser(t, "goalempty@test.com")
	token := getAuthToken(t, email, "password123")

	r := gin.New()
	r.GET("/api/goals", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, GetGoals)

	req, _ := http.NewRequest("GET", "/api/goals?status=active", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestCreateGoal_Success(t *testing.T) {
	setupHandlerTest(t)
	uid, email := createTestUser(t, "goalcreate@test.com")
	token := getAuthToken(t, email, "password123")

	r := gin.New()
	r.POST("/api/goals", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, CreateGoal)

	body, _ := json.Marshal(map[string]interface{}{
		"name":         "New Car",
		"targetAmount": 100000000,
	})
	req, _ := http.NewRequest("POST", "/api/goals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSyncData_Success(t *testing.T) {
	setupHandlerTest(t)
	uid, email := createTestUser(t, "sync@test.com")
	token := getAuthToken(t, email, "password123")

	r := gin.New()
	r.POST("/api/sync", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, SyncData)

	body, _ := json.Marshal(dto.DataSyncRequest{
		LastSyncTimestamp: "",
		Changes:           []dto.SyncChange{},
	})
	req, _ := http.NewRequest("POST", "/api/sync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
