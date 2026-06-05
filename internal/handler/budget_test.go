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
	"spendwise-ms/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func setupBudgetTestDB() {
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

func setupBudgetRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	budgets := r.Group("/api/budgets")
	budgets.Use(middleware.AuthRequired())
	{
		budgets.GET("", GetBudgets)
		budgets.POST("", CreateBudget)
		budgets.GET("/:id", GetBudgetByID)
		budgets.PUT("/:id", UpdateBudget)
		budgets.DELETE("/:id", DeleteBudget)
	}

	return r
}

func createUserAndTokenForBudget() (string, string, string) {
	userID := uuid.New().String()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	config.DB.Create(&model.User{
		ID:       userID,
		Name:     "Budget Handler Test",
		Email:    "budgethandler@spendwise.com",
		Password: string(hash),
	})

	catID := uuid.New().String()
	config.DB.Create(&model.Category{
		ID: catID, UserID: userID, Name: "Food", Type: "expense", IsDefault: false,
	})

	token, _, _ := service.GenerateTokens(userID, "budgethandler@spendwise.com")
	return token, userID, catID
}

func TestCreateBudgetHandler_Success(t *testing.T) {
	setupBudgetTestDB()
	r := setupBudgetRouter()
	token, _, catID := createUserAndTokenForBudget()

	body, _ := json.Marshal(dto.CreateBudgetRequest{
		Name:       "Food Budget",
		Amount:     2000000,
		Period:     "monthly",
		CategoryID: catID,
	})

	req, _ := http.NewRequest("POST", "/api/budgets", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateBudgetHandler_Unauthenticated(t *testing.T) {
	setupBudgetTestDB()
	r := setupBudgetRouter()

	body, _ := json.Marshal(dto.CreateBudgetRequest{
		Name: "Food", Amount: 1000, Period: "daily", CategoryID: "some-id",
	})
	req, _ := http.NewRequest("POST", "/api/budgets", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestGetBudgetsHandler(t *testing.T) {
	setupBudgetTestDB()
	r := setupBudgetRouter()
	token, userID, catID := createUserAndTokenForBudget()

	svc := service.NewBudgetService(config.DB)
	svc.CreateBudget(userID, dto.CreateBudgetRequest{
		Name: "Food", Amount: 1000000, Period: "monthly", CategoryID: catID,
	})

	req, _ := http.NewRequest("GET", "/api/budgets", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetBudgetByIDHandler(t *testing.T) {
	setupBudgetTestDB()
	r := setupBudgetRouter()
	token, userID, catID := createUserAndTokenForBudget()

	svc := service.NewBudgetService(config.DB)
	created, _ := svc.CreateBudget(userID, dto.CreateBudgetRequest{
		Name: "Transport", Amount: 500000, Period: "weekly", CategoryID: catID,
	})

	req, _ := http.NewRequest("GET", "/api/budgets/"+created.ID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdateBudgetHandler(t *testing.T) {
	setupBudgetTestDB()
	r := setupBudgetRouter()
	token, userID, catID := createUserAndTokenForBudget()

	svc := service.NewBudgetService(config.DB)
	created, _ := svc.CreateBudget(userID, dto.CreateBudgetRequest{
		Name: "Old", Amount: 100000, Period: "monthly", CategoryID: catID,
	})

	body, _ := json.Marshal(dto.UpdateBudgetRequest{Name: "Updated Budget"})
	req, _ := http.NewRequest("PUT", "/api/budgets/"+created.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteBudgetHandler(t *testing.T) {
	setupBudgetTestDB()
	r := setupBudgetRouter()
	token, userID, catID := createUserAndTokenForBudget()

	svc := service.NewBudgetService(config.DB)
	created, _ := svc.CreateBudget(userID, dto.CreateBudgetRequest{
		Name: "DeleteMe", Amount: 10000, Period: "daily", CategoryID: catID,
	})

	req, _ := http.NewRequest("DELETE", "/api/budgets/"+created.ID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}
