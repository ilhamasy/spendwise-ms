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

func TestHandler_Logout(t *testing.T) {
	setupTestDB()
	r := gin.New()
	r.POST("/logout", Logout)
	req, _ := http.NewRequest("POST", "/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestHandler_LoginValidationError(t *testing.T) {
	setupTestDB()
	r := gin.New()
	r.POST("/login", Login)
	body, _ := json.Marshal(map[string]string{})
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestHandler_RefreshMissingCookie(t *testing.T) {
	setupTestDB()
	r := gin.New()
	r.POST("/refresh", RefreshToken)
	req, _ := http.NewRequest("POST", "/refresh", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

func TestHandler_RefreshInvalidToken(t *testing.T) {
	setupTestDB()
	r := gin.New()
	r.POST("/refresh", RefreshToken)
	req, _ := http.NewRequest("POST", "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "spendwise-refresh-token", Value: "bad.token"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

func TestHandler_GetBudgetsEmpty(t *testing.T) {
	setupTestDB()
	uid, token := createBoostUser(t)
	r := gin.New()
	r.GET("/budgets", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, GetBudgets)
	req, _ := http.NewRequest("GET", "/budgets", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestHandler_GetGoalsEmpty(t *testing.T) {
	setupTestDB()
	uid, token := createBoostUser(t)
	r := gin.New()
	r.GET("/goals", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, GetGoals)
	req, _ := http.NewRequest("GET", "/goals", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestHandler_GetCategoriesEmpty(t *testing.T) {
	setupTestDB()
	uid, token := createBoostUser(t)
	r := gin.New()
	r.GET("/categories", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, GetCategories)
	req, _ := http.NewRequest("GET", "/categories", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestHandler_GetTransactionsEmpty(t *testing.T) {
	setupTestDB()
	uid, token := createBoostUser(t)
	r := gin.New()
	r.GET("/tx", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, GetTransactions)
	req, _ := http.NewRequest("GET", "/tx", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestHandler_CreateTransaction(t *testing.T) {
	setupTestDB()
	uid, token := createBoostUser(t)
	cat := createBoostCategory(t, uid)

	r := gin.New()
	r.POST("/tx", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, CreateTransaction)
	body, _ := json.Marshal(dto.CreateTransactionRequest{Type: "expense", Amount: 50000, CategoryID: cat, OccurredAt: "2026-06-28", Note: "Lunch"})
	req, _ := http.NewRequest("POST", "/tx", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_SyncData(t *testing.T) {
	setupTestDB()
	uid, token := createBoostUser(t)
	cat := createBoostCategory(t, uid)

	r := gin.New()
	r.POST("/sync", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, SyncData)
	body, _ := json.Marshal(dto.DataSyncRequest{
		LastSyncTimestamp: "",
		Changes: []dto.SyncChange{
			{EntityType: "transaction", EntityID: uuid.New().String(), Operation: "CREATE", Payload: map[string]interface{}{"type": "expense", "amount": 20000, "categoryId": cat, "occurredAt": "2026-06-28"}},
		},
	})
	req, _ := http.NewRequest("POST", "/sync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_SyncTransactions(t *testing.T) {
	setupTestDB()
	uid, token := createBoostUser(t)
	cat := createBoostCategory(t, uid)

	r := gin.New()
	r.POST("/txsync", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, SyncTransactions)
	body, _ := json.Marshal(dto.SyncRequest{Transactions: []dto.SyncTransactionRequest{{Type: "expense", Amount: 30000, CategoryID: cat, OccurredAt: "2026-06-28"}}})
	req, _ := http.NewRequest("POST", "/txsync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_GetBudgetByID(t *testing.T) {
	setupTestDB()
	uid, token := createBoostUser(t)
	cat := createBoostCategory(t, uid)

	r := gin.New()
	r.POST("/budgets", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, CreateBudget)
	r.GET("/budgets/:id", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, GetBudgetByID)

	body, _ := json.Marshal(dto.CreateBudgetRequest{Name: "B", Amount: 100000, Period: "monthly", CategoryID: cat})
	req, _ := http.NewRequest("POST", "/budgets", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var resp dto.BudgetResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	req2, _ := http.NewRequest("GET", "/budgets/"+resp.ID, nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w2.Code)
	}
}

func createBoostUser(t *testing.T) (string, string) {
	t.Helper()
	hash, _ := bcrypt.GenerateFromPassword([]byte("pass123"), 12)
	uid := uuid.New().String()
	email := "bst" + uid[:8] + "@t.com"
	config.DB.Create(&model.User{ID: uid, Name: "Boost", Email: email, Password: string(hash)})

	r := gin.New()
	r.POST("/login", Login)
	body, _ := json.Marshal(dto.LoginRequest{Email: email, Password: "pass123"})
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	for _, c := range w.Result().Cookies() {
		if c.Name == "spendwise-access-token" { return uid, c.Value }
	}
	t.Fatal("no access token")
	return "", ""
}

func createBoostCategory(t *testing.T, uid string) string {
	t.Helper()
	id := uuid.New().String()
	config.DB.Create(&model.Category{ID: id, UserID: uid, Name: "BoostCat", Type: "expense", IsDefault: true})
	return id
}

func TestHandler_UpdateGoal(t *testing.T) {
	setupTestDB()
	uid, token := createBoostUser(t)

	r := gin.New()
	r.POST("/goals", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, CreateGoal)
	r.PUT("/goals/:id", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, UpdateGoal)

	body, _ := json.Marshal(dto.CreateGoalRequest{Name: "OldGoal", TargetAmount: 1000000})
	req, _ := http.NewRequest("POST", "/goals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var resp dto.GoalResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	body2, _ := json.Marshal(dto.UpdateGoalRequest{Name: "NewGoal", TargetAmount: 2000000})
	req2, _ := http.NewRequest("PUT", "/goals/"+resp.ID, bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK { t.Errorf("Expected 200, got %d", w2.Code) }
}

func TestHandler_AddContribution(t *testing.T) {
	setupTestDB()
	uid, token := createBoostUser(t)

	r := gin.New()
	r.POST("/goals", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, CreateGoal)
	r.POST("/goals/:id/contributions", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, AddContribution)

	body, _ := json.Marshal(dto.CreateGoalRequest{Name: "GoalWithContrib", TargetAmount: 1000000})
	req, _ := http.NewRequest("POST", "/goals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var resp dto.GoalResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	body2, _ := json.Marshal(dto.CreateContributionRequest{Amount: 100000, Date: "2026-06-28"})
	req2, _ := http.NewRequest("POST", "/goals/"+resp.ID+"/contributions", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusCreated { t.Errorf("Expected 201, got %d", w2.Code) }
}



func TestHandler_DeleteGoal(t *testing.T) {
	setupTestDB()
	uid, token := createBoostUser(t)

	r := gin.New()
	r.POST("/goals", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, CreateGoal)
	r.DELETE("/goals/:id", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, DeleteGoal)

	body, _ := json.Marshal(dto.CreateGoalRequest{Name: "DelMe", TargetAmount: 500000})
	req, _ := http.NewRequest("POST", "/goals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var resp dto.GoalResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	req2, _ := http.NewRequest("DELETE", "/goals/"+resp.ID, nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusNoContent { t.Errorf("Expected 204, got %d", w2.Code) }
}

func TestHandler_UpdateCategory(t *testing.T) {
	setupTestDB()
	uid, token := createBoostUser(t)

	r := gin.New()
	r.POST("/cats", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, CreateCategory)
	r.PUT("/cats/:id", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, UpdateCategory)

	body, _ := json.Marshal(dto.CreateCategoryRequest{Name: "OldCat", Type: "expense"})
	req, _ := http.NewRequest("POST", "/cats", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var resp dto.CategoryResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	body2, _ := json.Marshal(dto.UpdateCategoryRequest{Name: "UpdatedCat", Color: "#FFF"})
	req2, _ := http.NewRequest("PUT", "/cats/"+resp.ID, bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK { t.Errorf("Expected 200, got %d", w2.Code) }
}

func TestHandler_ArchiveGoal(t *testing.T) {
	setupTestDB()
	uid, token := createBoostUser(t)

	r := gin.New()
	r.POST("/goals", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, CreateGoal)
	r.PATCH("/goals/:id/archive", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, ArchiveGoal)

	body, _ := json.Marshal(dto.CreateGoalRequest{Name: "ArchiveGoal", TargetAmount: 500000})
	req, _ := http.NewRequest("POST", "/goals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var resp dto.GoalResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	req2, _ := http.NewRequest("PATCH", "/goals/"+resp.ID+"/archive", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK { t.Errorf("Expected 200, got %d", w2.Code) }
}

func TestHandler_UnarchiveGoal(t *testing.T) {
	setupTestDB()
	uid, token := createBoostUser(t)

	r := gin.New()
	r.POST("/goals", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, CreateGoal)
	r.PATCH("/goals/:id/archive", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, ArchiveGoal)
	r.PATCH("/goals/:id/unarchive", func(c *gin.Context) { c.Set("userId", uid); c.Next() }, UnarchiveGoal)

	body, _ := json.Marshal(dto.CreateGoalRequest{Name: "UnarcGoal", TargetAmount: 500000})
	req, _ := http.NewRequest("POST", "/goals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var resp dto.GoalResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	req2, _ := http.NewRequest("PATCH", "/goals/"+resp.ID+"/archive", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	req3, _ := http.NewRequest("PATCH", "/goals/"+resp.ID+"/unarchive", nil)
	req3.Header.Set("Authorization", "Bearer "+token)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK { t.Errorf("Expected 200, got %d", w3.Code) }
}
