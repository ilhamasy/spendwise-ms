package service

import (
	"fmt"

	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"
	"spendwise-ms/internal/repository"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var budgetValidate = validator.New()

type BudgetService struct {
	repo    *repository.BudgetRepository
	catRepo *repository.CategoryRepository
}

func NewBudgetService(db *gorm.DB) *BudgetService {
	return &BudgetService{
		repo:    repository.NewBudgetRepository(db),
		catRepo: repository.NewCategoryRepository(db),
	}
}

func (s *BudgetService) ListBudgets(userID, period string) ([]dto.BudgetResponse, error) {
	budgets, err := s.repo.FindAll(userID, period)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve budgets: %w", err)
	}

	var resp []dto.BudgetResponse
	for _, b := range budgets {
		resp = append(resp, s.budgetToResponse(b))
	}
	return resp, nil
}

func (s *BudgetService) CreateBudget(userID string, req dto.CreateBudgetRequest) (*dto.BudgetResponse, error) {
	if err := budgetValidate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	_, err := s.catRepo.FindByID(req.CategoryID, userID)
	if err != nil {
		return nil, fmt.Errorf("category not found")
	}

	budget := model.Budget{
		UserID:     userID,
		Name:       req.Name,
		Amount:     req.Amount,
		Period:     req.Period,
		CategoryID: req.CategoryID,
	}

	if err := s.repo.Create(&budget); err != nil {
		return nil, fmt.Errorf("failed to create budget: %w", err)
	}

	resp := s.budgetToResponse(budget)
	return &resp, nil
}

func (s *BudgetService) GetBudget(userID, id string) (*dto.BudgetResponse, error) {
	budget, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, err
	}
	resp := s.budgetToResponse(*budget)
	return &resp, nil
}

func (s *BudgetService) UpdateBudget(userID, id string, req dto.UpdateBudgetRequest) (*dto.BudgetResponse, error) {
	if err := budgetValidate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	budget, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		budget.Name = req.Name
	}
	if req.Amount > 0 {
		budget.Amount = req.Amount
	}
	if req.Period != "" {
		budget.Period = req.Period
	}
	if req.CategoryID != "" {
		_, err := s.catRepo.FindByID(req.CategoryID, userID)
		if err != nil {
			return nil, fmt.Errorf("category not found")
		}
		budget.CategoryID = req.CategoryID
	}

	if err := s.repo.Update(budget); err != nil {
		return nil, fmt.Errorf("failed to update budget: %w", err)
	}

	updated, _ := s.repo.FindByID(id, userID)
	resp := s.budgetToResponse(*updated)
	return &resp, nil
}

func (s *BudgetService) DeleteBudget(userID, id string) error {
	return s.repo.Delete(id, userID)
}

func (s *BudgetService) budgetToResponse(b model.Budget) dto.BudgetResponse {
	spent, _ := s.repo.GetSpentAmount(b.UserID, b.CategoryID, b.Period)
	remaining := b.Amount - spent
	if remaining < 0 {
		remaining = 0
	}

	return dto.BudgetResponse{
		ID:         b.ID,
		Name:       b.Name,
		Amount:     b.Amount,
		Period:     b.Period,
		CategoryID: b.CategoryID,
		Spent:      spent,
		Remaining:  remaining,
		CreatedAt:  b.CreatedAt,
		UpdatedAt:  b.UpdatedAt,
	}
}
