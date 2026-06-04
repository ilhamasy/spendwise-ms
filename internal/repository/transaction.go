package repository

import (
	"fmt"
	"math"
	"spendwise-ms/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

type TransactionFilter struct {
	Type       string
	CategoryID string
	StartDate  string
	EndDate    string
	Sort       string
	Page       int
	Limit      int
	Search     string
	UserID     string
}

func (r *TransactionRepository) Create(tx *model.Transaction) error {
	tx.ID = uuid.New().String()
	return r.db.Create(tx).Error
}

func (r *TransactionRepository) FindByID(id, userID string) (*model.Transaction, error) {
	var tx model.Transaction
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *TransactionRepository) FindAll(filter TransactionFilter) ([]model.Transaction, int64, error) {
	var transactions []model.Transaction
	var total int64

	query := r.db.Model(&model.Transaction{}).Where("user_id = ?", filter.UserID)

	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.CategoryID != "" {
		query = query.Where("category_id = ?", filter.CategoryID)
	}
	if filter.StartDate != "" {
		query = query.Where("occurred_at >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		query = query.Where("occurred_at <= ?", filter.EndDate)
	}
	if filter.Search != "" {
		query = query.Where("note ILIKE ?", "%"+filter.Search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "occurred_at DESC"
	switch filter.Sort {
	case "date_asc":
		orderClause = "occurred_at ASC"
	case "amount_desc":
		orderClause = "amount DESC"
	case "amount_asc":
		orderClause = "amount ASC"
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	if err := query.Order(orderClause).Limit(limit).Offset(offset).Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

func (r *TransactionRepository) Update(tx *model.Transaction) error {
	return r.db.Model(&model.Transaction{}).Where("id = ? AND user_id = ?", tx.ID, tx.UserID).Updates(map[string]interface{}{
		"type":        tx.Type,
		"amount":      tx.Amount,
		"category_id": tx.CategoryID,
		"occurred_at": tx.OccurredAt,
		"note":        tx.Note,
	}).Error
}

func (r *TransactionRepository) Delete(id, userID string) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Transaction{}).Error
}

func (r *TransactionRepository) FindCategoryByID(id, userID string) (*model.Category, error) {
	var cat model.Category
	err := r.db.Where("id = ? AND (user_id = ? OR is_default = true)", id, userID).First(&cat).Error
	if err != nil {
		return nil, fmt.Errorf("category not found")
	}
	return &cat, nil
}

func (r *TransactionRepository) BatchCreate(transactions []model.Transaction) (int, []int, []string) {
	var synced int
	var failedIndices []int
	var failedErrors []string

	for i, tx := range transactions {
		tx.ID = uuid.New().String()
		if err := r.db.Create(&tx).Error; err != nil {
			failedIndices = append(failedIndices, i)
			failedErrors = append(failedErrors, err.Error())
		} else {
			synced++
		}
	}

	return synced, failedIndices, failedErrors
}

func CalculateTotalPages(total int64, limit int) int64 {
	return int64(math.Ceil(float64(total) / float64(limit)))
}
