package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func setupTE(t *testing.T) (string, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
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

	hash, _ := bcrypt.GenerateFromPassword([]byte("pass123"), 12)
	uid := uuid.New().String()
	email := "u" + uid[:8] + "@t.com"
	config.DB.Create(&model.User{ID: uid, Name: "T", Email: email, Password: string(hash)})

	token := getToken(t, email, "pass123")
	return uid, token
}

func getToken(t *testing.T, email, pass string) string {
	t.Helper()
	r := gin.New()
	r.POST("/login", Login)
	body, _ := json.Marshal(map[string]string{"email": email, "password": pass})
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if t, ok := resp["accessToken"].(string); ok {
		return t
	}
	return ""
}

func TestGetTransactionByID_NotFound(t *testing.T) {
	uid, token := setupTE(t)
	r := gin.New()
	r.GET("/tx/:id", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, GetTransactionByID)

	req, _ := http.NewRequest("GET", "/tx/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", w.Code)
	}
}

func TestUpdateTransaction_NotFound(t *testing.T) {
	uid, token := setupTE(t)
	r := gin.New()
	r.PUT("/tx/:id", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, UpdateTransaction)

	body, _ := json.Marshal(map[string]interface{}{"amount": 50000})
	req, _ := http.NewRequest("PUT", "/tx/nonexistent", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", w.Code)
	}
}

func TestDeleteTransaction_NotFound(t *testing.T) {
	uid, token := setupTE(t)
	r := gin.New()
	r.DELETE("/tx/:id", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, DeleteTransaction)

	req, _ := http.NewRequest("DELETE", "/tx/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", w.Code)
	}
}

func TestCreateTransaction_ValidationError(t *testing.T) {
	uid, token := setupTE(t)
	r := gin.New()
	r.POST("/tx", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, CreateTransaction)

	// Missing required fields
	body, _ := json.Marshal(map[string]interface{}{"type": "expense"})
	req, _ := http.NewRequest("POST", "/tx", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestCreateBudget_ValidationError(t *testing.T) {
	uid, token := setupTE(t)
	r := gin.New()
	r.POST("/budgets", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, CreateBudget)

	body, _ := json.Marshal(map[string]interface{}{"name": "Test"})
	req, _ := http.NewRequest("POST", "/budgets", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestCreateGoal_ValidationError(t *testing.T) {
	uid, token := setupTE(t)
	r := gin.New()
	r.POST("/goals", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, CreateGoal)

	body, _ := json.Marshal(map[string]interface{}{"name": "Test"})
	req, _ := http.NewRequest("POST", "/goals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestCreateCategory_ValidationError(t *testing.T) {
	uid, token := setupTE(t)
	r := gin.New()
	r.POST("/categories", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, CreateCategory)

	body, _ := json.Marshal(map[string]interface{}{"name": "Test"})
	req, _ := http.NewRequest("POST", "/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestGetBudgets_WithFilter(t *testing.T) {
	uid, token := setupTE(t)
	r := gin.New()
	r.GET("/budgets", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, GetBudgets)

	req, _ := http.NewRequest("GET", "/budgets?period=monthly", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestGetBudgetByID_NotFound(t *testing.T) {
	uid, token := setupTE(t)
	r := gin.New()
	r.GET("/budgets/:id", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, GetBudgetByID)

	req, _ := http.NewRequest("GET", "/budgets/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", w.Code)
	}
}

func TestGetGoalByID_NotFound(t *testing.T) {
	uid, token := setupTE(t)
	r := gin.New()
	r.GET("/goals/:id", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, GetGoalByID)

	req, _ := http.NewRequest("GET", "/goals/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", w.Code)
	}
}

func TestArchiveGoal_Handler(t *testing.T) {
	uid, token := setupTE(t)
	goalID := uuid.New().String()
	config.DB.Create(&model.SavingGoal{ID: goalID, UserID: uid, Name: "Test", TargetAmount: 1000, Status: "active"})

	r := gin.New()
	r.PATCH("/goals/:id/archive", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, ArchiveGoal)

	req, _ := http.NewRequest("PATCH", "/goals/"+goalID+"/archive", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestUnarchiveGoal_Handler(t *testing.T) {
	uid, token := setupTE(t)
	goalID := uuid.New().String()
	config.DB.Create(&model.SavingGoal{ID: goalID, UserID: uid, Name: "Test", TargetAmount: 1000, Status: "archived"})

	r := gin.New()
	r.PATCH("/goals/:id/unarchive", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, UnarchiveGoal)

	req, _ := http.NewRequest("PATCH", "/goals/"+goalID+"/unarchive", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestAddContribution_Handler(t *testing.T) {
	uid, token := setupTE(t)
	goalID := uuid.New().String()
	config.DB.Create(&model.SavingGoal{ID: goalID, UserID: uid, Name: "Test", TargetAmount: 100000, Status: "active"})

	r := gin.New()
	r.POST("/goals/:id/contributions", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, AddContribution)

	body, _ := json.Marshal(map[string]interface{}{"amount": 25000, "note": "Test deposit"})
	req, _ := http.NewRequest("POST", "/goals/"+goalID+"/contributions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetContributionHistory_Handler(t *testing.T) {
	uid, token := setupTE(t)
	goalID := uuid.New().String()
	config.DB.Create(&model.SavingGoal{ID: goalID, UserID: uid, Name: "Test", TargetAmount: 100000, Status: "active"})

	r := gin.New()
	r.GET("/goals/:id/contributions", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, GetContributionHistory)

	req, _ := http.NewRequest("GET", "/goals/"+goalID+"/contributions?page=1&limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestExportUserData_Handler(t *testing.T) {
	uid, token := setupTE(t)
	r := gin.New()
	r.GET("/export", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, ExportUserData)

	req, _ := http.NewRequest("GET", "/export", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestSyncTransactions_Handler(t *testing.T) {
	uid, token := setupTE(t)
	r := gin.New()
	r.POST("/tx/sync", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, SyncTransactions)

	body, _ := json.Marshal([]map[string]interface{}{})
	req, _ := http.NewRequest("POST", "/tx/sync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}
