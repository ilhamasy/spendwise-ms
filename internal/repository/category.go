package repository

import (
	"fmt"
	"time"

	"spendwise-ms/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) FindAll(userID, filterType string) ([]model.Category, error) {
	var categories []model.Category
	query := r.db.Where("(user_id = ? OR is_default = true) AND deleted_at IS NULL", userID)
	if filterType != "" {
		query = query.Where("type = ?", filterType)
	}
	if err := query.Order("is_default DESC, name ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *CategoryRepository) FindAllForSync(userID string) ([]model.Category, error) {
	var categories []model.Category
	if err := r.db.Unscoped().Where("user_id = ? OR is_default = true", userID).Order("is_default DESC, name ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *CategoryRepository) FindByID(id, userID string) (*model.Category, error) {
	var cat model.Category
	err := r.db.Where("id = ? AND (user_id = ? OR is_default = true)", id, userID).First(&cat).Error
	if err != nil {
		return nil, fmt.Errorf("category not found")
	}
	return &cat, nil
}

func (r *CategoryRepository) FindByNameAndType(name, catType, userID string) (*model.Category, error) {
	var cat model.Category
	err := r.db.Where("name = ? AND type = ? AND (user_id = ? OR is_default = true)", name, catType, userID).First(&cat).Error
	if err != nil {
		return nil, nil
	}
	return &cat, nil
}

func (r *CategoryRepository) Create(cat *model.Category) error {
	if cat.ID == "" {
		cat.ID = uuid.New().String()
	}
	return r.db.Create(cat).Error
}

func (r *CategoryRepository) Update(cat *model.Category) error {
	return r.db.Model(&model.Category{}).Where("id = ? AND user_id = ?", cat.ID, cat.UserID).Updates(map[string]interface{}{
		"name":       cat.Name,
		"icon":       cat.Icon,
		"color":      cat.Color,
		"deleted_at": cat.DeletedAt,
		"updated_at": cat.UpdatedAt,
	}).Error
}

func (r *CategoryRepository) Delete(id, userID string) error {
	now := time.Now()
	return r.db.Model(&model.Category{}).Where("id = ? AND user_id = ? AND is_default = false", id, userID).Update("deleted_at", now).Error
}

func (r *CategoryRepository) CountTransactionsByCategoryID(categoryID string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Transaction{}).Where("category_id = ?", categoryID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *CategoryRepository) ReassignTransactions(fromCategoryID, toCategoryID, userID string) error {
	return r.db.Model(&model.Transaction{}).Where("category_id = ? AND user_id = ?", fromCategoryID, userID).Update("category_id", toCategoryID).Error
}
