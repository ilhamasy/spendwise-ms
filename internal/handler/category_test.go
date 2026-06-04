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

func setupCategoryTestDB() {
	if config.DB != nil {
		config.DB.Where("1 = 1").Delete(&model.Transaction{})
		config.DB.Where("1 = 1").Delete(&model.Category{})
		config.DB.Where("1 = 1").Delete(&model.User{})
		return
	}
	config.InitDB(config.Load())
	config.AutoMigrate(&model.Transaction{}, &model.Category{}, &model.User{})
	config.DB.Where("1 = 1").Delete(&model.Transaction{})
	config.DB.Where("1 = 1").Delete(&model.Category{})
	config.DB.Where("1 = 1").Delete(&model.User{})
}

func setupCategoryRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	categories := r.Group("/api/categories")
	categories.Use(middleware.AuthRequired())
	{
		categories.GET("", GetCategories)
		categories.POST("", CreateCategory)
		categories.PUT("/:id", UpdateCategory)
		categories.DELETE("/:id", DeleteCategory)
	}

	return r
}

func createUserAndTokenForCat() (string, string) {
	userID := uuid.New().String()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	config.DB.Create(&model.User{
		ID:       userID,
		Name:     "Cat Handler Test",
		Email:    "cathandler@spendwise.com",
		Password: string(hash),
	})
	token, _, _ := service.GenerateTokens(userID, "cathandler@spendwise.com")
	return token, userID
}

func TestCreateCategoryHandler_Success(t *testing.T) {
	setupCategoryTestDB()
	r := setupCategoryRouter()
	token, _ := createUserAndTokenForCat()

	body, _ := json.Marshal(dto.CreateCategoryRequest{
		Name:  "Food",
		Type:  "expense",
		Icon:  "food",
		Color: "#FF0000",
	})

	req, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateCategoryHandler_Unauthenticated(t *testing.T) {
	setupCategoryTestDB()
	r := setupCategoryRouter()

	body, _ := json.Marshal(dto.CreateCategoryRequest{Name: "Food", Type: "expense"})
	req, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestGetCategoriesHandler_Success(t *testing.T) {
	setupCategoryTestDB()
	r := setupCategoryRouter()
	token, _ := createUserAndTokenForCat()

	body, _ := json.Marshal(dto.CreateCategoryRequest{Name: "Food", Type: "expense"})
	for i := 0; i < 2; i++ {
		req, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	// Second creation may fail as duplicate — that's fine, test the list
	req, _ := http.NewRequest("GET", "/api/categories", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetCategoriesHandler_FilterByType(t *testing.T) {
	setupCategoryTestDB()
	r := setupCategoryRouter()
	token, userID := createUserAndTokenForCat()

	svc := service.NewCategoryService(config.DB)
	svc.CreateCategory(userID, dto.CreateCategoryRequest{Name: "Food", Type: "expense"})
	svc.CreateCategory(userID, dto.CreateCategoryRequest{Name: "Salary", Type: "income"})

	req, _ := http.NewRequest("GET", "/api/categories?type=income", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp []dto.CategoryResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 1 {
		t.Errorf("Expected 1 income category, got %d", len(resp))
	}
}

func TestUpdateCategoryHandler_Success(t *testing.T) {
	setupCategoryTestDB()
	r := setupCategoryRouter()
	token, userID := createUserAndTokenForCat()

	svc := service.NewCategoryService(config.DB)
	created, _ := svc.CreateCategory(userID, dto.CreateCategoryRequest{Name: "Old", Type: "expense"})

	body, _ := json.Marshal(dto.UpdateCategoryRequest{
		Name:  "New Name",
		Icon:  "star",
		Color: "#00FF00",
	})
	req, _ := http.NewRequest("PUT", "/api/categories/"+created.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteCategoryHandler_WithReassign(t *testing.T) {
	setupCategoryTestDB()
	r := setupCategoryRouter()
	token, userID := createUserAndTokenForCat()

	svc := service.NewCategoryService(config.DB)
	cat1, _ := svc.CreateCategory(userID, dto.CreateCategoryRequest{Name: "Old", Type: "expense"})
	cat2, _ := svc.CreateCategory(userID, dto.CreateCategoryRequest{Name: "New", Type: "expense"})

	req, _ := http.NewRequest("DELETE", "/api/categories/"+cat1.ID+"?reassignTo="+cat2.ID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d: %s", w.Code, w.Body.String())
	}
}
