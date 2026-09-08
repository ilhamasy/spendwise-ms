package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/middleware"
	"spendwise-ms/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func setupTest(t *testing.T) (*gin.Engine, string, string) {
	t.Helper()
	middleware.ResetLimiters()
	gin.SetMode(gin.TestMode)
	if config.DB == nil {
		cfg := config.Load()
		config.InitDB(cfg)
		config.AutoMigrate(&model.User{}, &model.Transaction{}, &model.Category{}, &model.SavingGoal{}, &model.GoalContribution{}, &model.Budget{})
	}
	config.DB.Where("1 = 1").Delete(&model.Transaction{})
	config.DB.Where("1 = 1").Delete(&model.Category{})
	config.DB.Where("1 = 1").Delete(&model.Budget{})
	config.DB.Where("1 = 1").Delete(&model.GoalContribution{})
	config.DB.Where("1 = 1").Delete(&model.SavingGoal{})
	config.DB.Where("1 = 1").Delete(&model.User{})

	hash, _ := bcrypt.GenerateFromPassword([]byte("pass123"), 12)
	uid := uuid.New().String()
	email := "u" + uid[:8] + "@test.com"
	config.DB.Create(&model.User{ID: uid, Name: "Tester", Email: email, Password: string(hash)})

	token := loginAndGetToken(t, email, "pass123")
	return nil, uid, token
}

func loginAndGetToken(t *testing.T, email, pass string) string {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/login", Login)
	body, _ := json.Marshal(dto.LoginRequest{Email: email, Password: pass})
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
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

func createTestCategory(t *testing.T, uid, name, catType string) string {
	t.Helper()
	cat := &model.Category{ID: uuid.New().String(), Name: name, Type: catType, Icon: "📁", UserID: uid}
	config.DB.Create(cat)
	return cat.ID
}

func TestUpdateCategory_Handler(t *testing.T) {
	_, uid, token := setupTest(t)
	catID := createTestCategory(t, uid, "UpdateMe", "expense")

	r := gin.New()
	r.PUT("/api/categories/:id", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, UpdateCategory)

	body, _ := json.Marshal(map[string]string{"name": "UpdatedCat", "icon": "🔄", "color": "#fff"})
	req, _ := http.NewRequest("PUT", "/api/categories/"+catID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteCategory_Handler(t *testing.T) {
	_, uid, token := setupTest(t)
	catID := createTestCategory(t, uid, "DeleteMe", "expense")

	r := gin.New()
	r.DELETE("/api/categories/:id", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, DeleteCategory)

	req, _ := http.NewRequest("DELETE", "/api/categories/"+catID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdateTransaction_Handler(t *testing.T) {
	_, uid, token := setupTest(t)
	catID := createTestCategory(t, uid, "Food", "expense")
	txID := uuid.New().String()
	config.DB.Create(&model.Transaction{ID: txID, UserID: uid, Type: "expense", Amount: 10000, CategoryID: catID, OccurredAt: "2026-06-01"})

	r := gin.New()
	r.PUT("/api/transactions/:id", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, UpdateTransaction)

	body, _ := json.Marshal(map[string]interface{}{"amount": 25000, "note": "Updated"})
	req, _ := http.NewRequest("PUT", "/api/transactions/"+txID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteTransaction_Handler(t *testing.T) {
	_, uid, token := setupTest(t)
	catID := createTestCategory(t, uid, "Food", "expense")
	txID := uuid.New().String()
	config.DB.Create(&model.Transaction{ID: txID, UserID: uid, Type: "expense", Amount: 10000, CategoryID: catID, OccurredAt: "2026-06-01"})

	r := gin.New()
	r.DELETE("/api/transactions/:id", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, DeleteTransaction)

	req, _ := http.NewRequest("DELETE", "/api/transactions/"+txID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdateGoal_Handler(t *testing.T) {
	_, uid, token := setupTest(t)
	goalID := uuid.New().String()
	config.DB.Create(&model.SavingGoal{ID: goalID, UserID: uid, Name: "Old Goal", TargetAmount: 1000, Status: "active"})

	r := gin.New()
	r.PUT("/api/goals/:id", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, UpdateGoal)

	body, _ := json.Marshal(map[string]interface{}{"name": "Updated Goal", "targetAmount": 2000})
	req, _ := http.NewRequest("PUT", "/api/goals/"+goalID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteGoal_Handler(t *testing.T) {
	_, uid, token := setupTest(t)
	goalID := uuid.New().String()
	config.DB.Create(&model.SavingGoal{ID: goalID, UserID: uid, Name: "ToDelete", TargetAmount: 1000, Status: "active"})

	r := gin.New()
	r.DELETE("/api/goals/:id", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, DeleteGoal)

	req, _ := http.NewRequest("DELETE", "/api/goals/"+goalID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdateBudget_Handler(t *testing.T) {
	_, uid, token := setupTest(t)
	catID := createTestCategory(t, uid, "Food", "expense")
	budID := uuid.New().String()
	config.DB.Create(&model.Budget{ID: budID, UserID: uid, Name: "Old Budget", Amount: 1000000, Period: "monthly", CategoryID: catID})

	r := gin.New()
	r.PUT("/api/budgets/:id", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, UpdateBudget)

	body, _ := json.Marshal(map[string]interface{}{"name": "Updated Budget", "amount": 3000000})
	req, _ := http.NewRequest("PUT", "/api/budgets/"+budID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteBudget_Handler(t *testing.T) {
	_, uid, token := setupTest(t)
	catID := createTestCategory(t, uid, "Food", "expense")
	budID := uuid.New().String()
	config.DB.Create(&model.Budget{ID: budID, UserID: uid, Name: "ToDelete", Amount: 1000000, Period: "monthly", CategoryID: catID})

	r := gin.New()
	r.DELETE("/api/budgets/:id", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, DeleteBudget)

	req, _ := http.NewRequest("DELETE", "/api/budgets/"+budID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestChangePassword_Handler(t *testing.T) {
	_, uid, token := setupTest(t)

	r := gin.New()
	r.PUT("/api/password", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, ChangePassword)

	body, _ := json.Marshal(map[string]string{
		"currentPassword": "pass123",
		"newPassword":     "newpass456",
	})
	req, _ := http.NewRequest("PUT", "/api/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteAccount_Handler(t *testing.T) {
	_, uid, token := setupTest(t)

	r := gin.New()
	r.DELETE("/api/account", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, DeleteAccount)

	body, _ := json.Marshal(map[string]string{"confirmation": "DELETE"})
	req, _ := http.NewRequest("DELETE", "/api/account", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected 204, got %d: %s", w.Code, w.Body.String())
	}
}
