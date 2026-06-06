package repository

import (
	"fmt"
	"time"

	"spendwise-ms/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BudgetRepository struct {
	db *gorm.DB
}

func NewBudgetRepository(db *gorm.DB) *BudgetRepository {
	return &BudgetRepository{db: db}
}

func (r *BudgetRepository) FindAll(userID, period string) ([]model.Budget, error) {
	var budgets []model.Budget
	query := r.db.Where("user_id = ?", userID)
	if period != "" {
		query = query.Where("period = ?", period)
	}
	if err := query.Order("created_at DESC").Find(&budgets).Error; err != nil {
		return nil, err
	}
	return budgets, nil
}

func (r *BudgetRepository) FindByID(id, userID string) (*model.Budget, error) {
	var budget model.Budget
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&budget).Error
	if err != nil {
		return nil, fmt.Errorf("budget not found")
	}
	return &budget, nil
}

func (r *BudgetRepository) Create(budget *model.Budget) error {
	if budget.ID == "" {
		budget.ID = uuid.New().String()
	}
	return r.db.Create(budget).Error
}

func (r *BudgetRepository) Update(budget *model.Budget) error {
	return r.db.Model(&model.Budget{}).Where("id = ? AND user_id = ?", budget.ID, budget.UserID).Updates(map[string]interface{}{
		"name":        budget.Name,
		"amount":      budget.Amount,
		"period":      budget.Period,
		"category_id": budget.CategoryID,
	}).Error
}

func (r *BudgetRepository) Delete(id, userID string) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Budget{})
	if result.RowsAffected == 0 {
		return fmt.Errorf("budget not found")
	}
	return result.Error
}

func (r *BudgetRepository) GetSpentAmount(userID, categoryID, period string) (int64, error) {
	var total int64
	start, end := getPeriodRange(period)

	query := r.db.Model(&model.Transaction{}).
		Where("user_id = ? AND category_id = ? AND type = 'expense'", userID, categoryID)

	if start != "" {
		query = query.Where("occurred_at >= ?", start)
	}
	if end != "" {
		query = query.Where("occurred_at <= ?", end)
	}

	if err := query.Select("COALESCE(SUM(amount), 0)").Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func getPeriodRange(period string) (string, string) {
	now := time.Now()
	switch period {
	case "daily":
		d := now.Format("2006-01-02")
		return d, d
	case "weekly":
		weekday := now.Weekday()
		if weekday == 0 {
			weekday = 7
		}
		start := now.AddDate(0, 0, -int(weekday-1)).Format("2006-01-02")
		end := now.AddDate(0, 0, 7-int(weekday)).Format("2006-01-02")
		return start, end
	case "monthly":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		end := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		return start, end
	case "yearly":
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		end := time.Date(now.Year(), 12, 31, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		return start, end
	}
	return "", ""
}
