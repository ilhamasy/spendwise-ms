package service

import (
	"errors"
	"fmt"

	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"
	"spendwise-ms/internal/repository"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var catValidate = validator.New()

type CategoryService struct {
	repo repository.ICategoryRepository
}

func NewCategoryService(db *gorm.DB) *CategoryService {
	return &CategoryService{repo: repository.NewCategoryRepository(db)}
}

func NewCategoryServiceWithRepo(r repository.ICategoryRepository) *CategoryService {
	return &CategoryService{repo: r}
}

func (s *CategoryService) ListCategories(userID, filterType string) ([]dto.CategoryResponse, error) {
	categories, err := s.repo.FindAll(userID, filterType)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve categories: %w", err)
	}

	var resp []dto.CategoryResponse
	for _, cat := range categories {
		resp = append(resp, catToResponse(cat))
	}
	return resp, nil
}

func (s *CategoryService) CreateCategory(userID string, req dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
	if err := catValidate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	existing, _ := s.repo.FindByNameAndType(req.Name, req.Type, userID)
	if existing != nil {
		return nil, fmt.Errorf("category with name '%s' already exists for type '%s'", req.Name, req.Type)
	}

	cat := model.Category{
		UserID:    userID,
		Name:      dto.SanitizeString(req.Name),
		Type:      req.Type,
		Icon:      dto.SanitizeString(req.Icon),
		Color:     dto.SanitizeString(req.Color),
		IsDefault: false,
	}

	if err := s.repo.Create(&cat); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	resp := catToResponse(cat)
	return &resp, nil
}

func (s *CategoryService) UpdateCategory(userID, id string, req dto.UpdateCategoryRequest) (*dto.CategoryResponse, error) {
	if err := catValidate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	cat, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, err
	}
	if cat.IsDefault {
		return nil, errors.New("cannot update default category")
	}

	if req.Name != "" {
		sanitizedName := dto.SanitizeString(req.Name)
		existing, _ := s.repo.FindByNameAndType(sanitizedName, cat.Type, userID)
		if existing != nil && existing.ID != cat.ID {
			return nil, fmt.Errorf("category with name '%s' already exists for type '%s'", sanitizedName, cat.Type)
		}
		cat.Name = sanitizedName
	}
	if req.Icon != "" {
		cat.Icon = dto.SanitizeString(req.Icon)
	}
	if req.Color != "" {
		cat.Color = dto.SanitizeString(req.Color)
	}

	if err := s.repo.Update(cat); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	resp := catToResponse(*cat)
	return &resp, nil
}

func (s *CategoryService) DeleteCategory(userID, id, reassignToCategoryID string) error {
	cat, err := s.repo.FindByID(id, userID)
	if err != nil {
		return err
	}
	if cat.IsDefault {
		return errors.New("cannot delete default category")
	}

	count, err := s.repo.CountTransactionsByCategoryID(id, userID)
	if err != nil {
		return fmt.Errorf("failed to check transaction references: %w", err)
	}

	if count > 0 {
		if reassignToCategoryID == "" {
			return fmt.Errorf("category has %d transactions, provide reassignTo to reassign and delete", count)
		}

		reassignCat, err := s.repo.FindByID(reassignToCategoryID, userID)
		if err != nil {
			return fmt.Errorf("reassign target category not found")
		}

		if err := s.repo.ReassignTransactions(id, reassignCat.ID, userID); err != nil {
			return fmt.Errorf("failed to reassign transactions: %w", err)
		}
	}

	if err := s.repo.Delete(id, userID); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}

func catToResponse(cat model.Category) dto.CategoryResponse {
	return dto.CategoryResponse{
		ID:        cat.ID,
		Name:      cat.Name,
		Type:      cat.Type,
		Icon:      cat.Icon,
		Color:     cat.Color,
		IsDefault: cat.IsDefault,
	}
}
