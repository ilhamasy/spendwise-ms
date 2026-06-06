package repository

import (
	"fmt"
	"math"

	"spendwise-ms/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GoalRepository struct {
	db *gorm.DB
}

func NewGoalRepository(db *gorm.DB) *GoalRepository {
	return &GoalRepository{db: db}
}

func (r *GoalRepository) FindAll(userID, status string) ([]model.SavingGoal, error) {
	var goals []model.SavingGoal
	query := r.db.Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("created_at DESC").Find(&goals).Error; err != nil {
		return nil, err
	}
	return goals, nil
}

func (r *GoalRepository) FindByID(id, userID string) (*model.SavingGoal, error) {
	var goal model.SavingGoal
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&goal).Error
	if err != nil {
		return nil, fmt.Errorf("goal not found")
	}
	return &goal, nil
}

func (r *GoalRepository) Create(goal *model.SavingGoal) error {
	if goal.ID == "" {
		goal.ID = uuid.New().String()
	}
	return r.db.Create(goal).Error
}

func (r *GoalRepository) Update(goal *model.SavingGoal) error {
	return r.db.Model(&model.SavingGoal{}).Where("id = ? AND user_id = ?", goal.ID, goal.UserID).Updates(map[string]interface{}{
		"name":          goal.Name,
		"target_amount": goal.TargetAmount,
		"current_saved": goal.CurrentSaved,
		"target_date":   goal.TargetDate,
	}).Error
}

func (r *GoalRepository) UpdateStatus(id, userID, status string) error {
	return r.db.Model(&model.SavingGoal{}).Where("id = ? AND user_id = ?", id, userID).Update("status", status).Error
}

func (r *GoalRepository) Delete(id, userID string) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.SavingGoal{})
	if result.RowsAffected == 0 {
		return fmt.Errorf("goal not found")
	}
	return result.Error
}

func (r *GoalRepository) CreateContribution(contribution *model.GoalContribution) error {
	contribution.ID = uuid.New().String()
	return r.db.Create(contribution).Error
}

func (r *GoalRepository) FindContributionsByGoalID(goalID string, page, limit int) ([]model.GoalContribution, int64, error) {
	var contributions []model.GoalContribution
	var total int64

	query := r.db.Model(&model.GoalContribution{}).Where("goal_id = ?", goalID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	if err := query.Order("date DESC, created_at DESC").Limit(limit).Offset(offset).Find(&contributions).Error; err != nil {
		return nil, 0, err
	}

	return contributions, total, nil
}

func CalculateGoalProgress(current, target int64) float64 {
	if target == 0 {
		return 0
	}
	progress := float64(current) / float64(target) * 100
	return math.Round(progress*100) / 100
}
