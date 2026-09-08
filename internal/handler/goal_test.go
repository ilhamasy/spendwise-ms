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

func setupGoalTestDB() {
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

func setupGoalRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	goals := r.Group("/api/goals")
	goals.Use(middleware.AuthRequired())
	{
		goals.GET("", GetGoals)
		goals.POST("", CreateGoal)
		goals.GET("/:id", GetGoalByID)
		goals.PUT("/:id", UpdateGoal)
		goals.PATCH("/:id/archive", ArchiveGoal)
		goals.PATCH("/:id/unarchive", UnarchiveGoal)
		goals.DELETE("/:id", DeleteGoal)
		goals.POST("/:id/contributions", AddContribution)
		goals.GET("/:id/contributions", GetContributionHistory)
	}

	return r
}

func createUserAndTokenForGoal() (string, string) {
	userID := uuid.New().String()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	config.DB.Create(&model.User{
		ID:       userID,
		Name:     "Goal Handler Test",
		Email:    "goalhandler@spendwise.com",
		Password: string(hash),
	})
	token, _, _ := service.GenerateTokens(userID, "goalhandler@spendwise.com")
	return token, userID
}

func TestCreateGoalHandler_Success(t *testing.T) {
	setupGoalTestDB()
	r := setupGoalRouter()
	token, _ := createUserAndTokenForGoal()

	body, _ := json.Marshal(dto.CreateGoalRequest{
		Name:         "New Car",
		TargetAmount: 100000000,
		CurrentSaved: 5000000,
	})

	req, _ := http.NewRequest("POST", "/api/goals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateGoalHandler_Unauthenticated(t *testing.T) {
	setupGoalTestDB()
	r := setupGoalRouter()

	body, _ := json.Marshal(dto.CreateGoalRequest{Name: "Car", TargetAmount: 50000})
	req, _ := http.NewRequest("POST", "/api/goals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestGetGoalsHandler_Success(t *testing.T) {
	setupGoalTestDB()
	r := setupGoalRouter()
	token, userID := createUserAndTokenForGoal()

	svc := service.NewGoalService(config.DB)
	svc.CreateGoal(userID, dto.CreateGoalRequest{Name: "Car", TargetAmount: 500000})
	svc.CreateGoal(userID, dto.CreateGoalRequest{Name: "House", TargetAmount: 500000000})

	req, _ := http.NewRequest("GET", "/api/goals", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp []dto.GoalResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 2 {
		t.Errorf("Expected 2 goals, got %d", len(resp))
	}
}

func TestGetGoalByIDHandler_Success(t *testing.T) {
	setupGoalTestDB()
	r := setupGoalRouter()
	token, userID := createUserAndTokenForGoal()

	svc := service.NewGoalService(config.DB)
	created, _ := svc.CreateGoal(userID, dto.CreateGoalRequest{Name: "Car", TargetAmount: 500000})

	req, _ := http.NewRequest("GET", "/api/goals/"+created.ID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestArchiveUnarchiveHandler(t *testing.T) {
	setupGoalTestDB()
	r := setupGoalRouter()
	token, userID := createUserAndTokenForGoal()

	svc := service.NewGoalService(config.DB)
	created, _ := svc.CreateGoal(userID, dto.CreateGoalRequest{Name: "Savings", TargetAmount: 100000})

	req, _ := http.NewRequest("PATCH", "/api/goals/"+created.ID+"/archive", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for archive, got %d: %s", w.Code, w.Body.String())
	}

	req2, _ := http.NewRequest("PATCH", "/api/goals/"+created.ID+"/unarchive", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status 200 for unarchive, got %d", w2.Code)
	}
}

func TestDeleteGoalHandler(t *testing.T) {
	setupGoalTestDB()
	r := setupGoalRouter()
	token, userID := createUserAndTokenForGoal()

	svc := service.NewGoalService(config.DB)
	created, _ := svc.CreateGoal(userID, dto.CreateGoalRequest{Name: "DeleteMe", TargetAmount: 50000})

	req, _ := http.NewRequest("DELETE", "/api/goals/"+created.ID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestAddContributionHandler(t *testing.T) {
	setupGoalTestDB()
	r := setupGoalRouter()
	token, userID := createUserAndTokenForGoal()

	svc := service.NewGoalService(config.DB)
	created, _ := svc.CreateGoal(userID, dto.CreateGoalRequest{Name: "Savings", TargetAmount: 100000})

	body, _ := json.Marshal(dto.CreateContributionRequest{
		Amount: 25000,
		Note:   "Monthly saving",
	})

	req, _ := http.NewRequest("POST", "/api/goals/"+created.ID+"/contributions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetContributionsHandler(t *testing.T) {
	setupGoalTestDB()
	r := setupGoalRouter()
	token, userID := createUserAndTokenForGoal()

	svc := service.NewGoalService(config.DB)
	created, _ := svc.CreateGoal(userID, dto.CreateGoalRequest{Name: "Savings", TargetAmount: 100000})
	svc.AddContribution(userID, created.ID, dto.CreateContributionRequest{Amount: 10000, Note: "Deposit"})
	svc.AddContribution(userID, created.ID, dto.CreateContributionRequest{Amount: 5000, Note: "Extra"})

	req, _ := http.NewRequest("GET", "/api/goals/"+created.ID+"/contributions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}
