package service

import (
	"fmt"

	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"
	"spendwise-ms/internal/repository"

	"gorm.io/gorm"
)

type ProfileService struct {
	db      *gorm.DB
	txnRepo *repository.TransactionRepository
	catRepo *repository.CategoryRepository
	goalRepo *repository.GoalRepository
}

func NewProfileService(db *gorm.DB) *ProfileService {
	return &ProfileService{
		db:      db,
		txnRepo: repository.NewTransactionRepository(db),
		catRepo: repository.NewCategoryRepository(db),
		goalRepo: repository.NewGoalRepository(db),
	}
}

func (s *ProfileService) GetProfile(userID string) (*dto.ProfileResponse, error) {
	var user model.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}
	resp := userToProfile(user)
	return &resp, nil
}

func (s *ProfileService) UpdateProfile(userID string, req dto.UpdateProfileRequest) (*dto.ProfileResponse, error) {
	updates := map[string]interface{}{}

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Currency != "" {
		updates["currency"] = req.Currency
	}
	if req.Theme != "" {
		updates["theme"] = req.Theme
	}
	updates["starting_balance"] = req.StartingBalance

	if len(updates) > 1 {
		if err := s.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update profile: %w", err)
		}
	}

	return s.GetProfile(userID)
}

func (s *ProfileService) ChangePassword(userID, currentPassword, newPassword string) error {
	var user model.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return fmt.Errorf("user not found")
	}

	if !CheckPassword(currentPassword, user.Password) {
		return fmt.Errorf("current password is incorrect")
	}

	hash, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.db.Model(&model.User{}).Where("id = ?", userID).Update("password", hash).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *ProfileService) DeleteAccount(userID, confirmation string) error {
	if confirmation != "DELETE" {
		return fmt.Errorf("type 'DELETE' to confirm account deletion")
	}

	tx := s.db.Begin()

	if err := tx.Where("goal_id IN (SELECT id FROM saving_goals WHERE user_id = ?)", userID).Delete(&model.GoalContribution{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete contributions: %w", err)
	}
	if err := tx.Where("user_id = ?", userID).Delete(&model.SavingGoal{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete goals: %w", err)
	}
	if err := tx.Where("user_id = ?", userID).Delete(&model.Transaction{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete transactions: %w", err)
	}
	if err := tx.Where("user_id = ?", userID).Delete(&model.Budget{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete budgets: %w", err)
	}
	if err := tx.Where("user_id = ?", userID).Delete(&model.Category{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete categories: %w", err)
	}
	if err := tx.Where("id = ?", userID).Delete(&model.User{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user: %w", err)
	}

	tx.Commit()
	return nil
}

func (s *ProfileService) ExportData(userID string) (*dto.ExportResponse, error) {
	profile, err := s.GetProfile(userID)
	if err != nil {
		return nil, err
	}

	txns, _, _ := s.txnRepo.FindAll(repository.TransactionFilter{UserID: userID, Page: 1, Limit: 10000})
	var transactions []dto.TransactionResponse
	for _, t := range txns {
		transactions = append(transactions, dto.TransactionResponse{
			ID: t.ID, Type: t.Type, Amount: t.Amount,
			CategoryID: t.CategoryID, OccurredAt: t.OccurredAt,
			Note: t.Note, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
		})
	}

	categories, _ := s.catRepo.FindAll(userID, "")
	var cats []dto.CategoryResponse
	for _, c := range categories {
		cats = append(cats, dto.CategoryResponse{
			ID: c.ID, Name: c.Name, Type: c.Type,
			Icon: c.Icon, Color: c.Color, IsDefault: c.IsDefault,
		})
	}

	goals, _ := s.goalRepo.FindAll(userID, "")
	var goalList []dto.GoalResponse
	for _, g := range goals {
		goalList = append(goalList, dto.GoalResponse{
			ID: g.ID, Name: g.Name, TargetAmount: g.TargetAmount,
			CurrentSaved: g.CurrentSaved, TargetDate: g.TargetDate,
			Status: g.Status, Progress: repository.CalculateGoalProgress(g.CurrentSaved, g.TargetAmount),
			CreatedAt: g.CreatedAt, UpdatedAt: g.UpdatedAt,
		})
	}

	var contributions []dto.ContributionResponse
	for _, g := range goals {
		contribs, _, _ := s.goalRepo.FindContributionsByGoalID(g.ID, 1, 10000)
		for _, c := range contribs {
			contributions = append(contributions, dto.ContributionResponse{
				ID: c.ID, GoalID: c.GoalID, Amount: c.Amount,
				Note: c.Note, Date: c.Date, CreatedAt: c.CreatedAt,
			})
		}
	}

	return &dto.ExportResponse{
		Profile:       *profile,
		Transactions:  transactions,
		Categories:    cats,
		Goals:         goalList,
		Contributions: contributions,
	}, nil
}

func userToProfile(user model.User) dto.ProfileResponse {
	return dto.ProfileResponse{
		ID:              user.ID,
		Name:            user.Name,
		Email:           user.Email,
		Currency:        user.Currency,
		Theme:           user.Theme,
		StartingBalance: user.StartingBalance,
		CreatedAt:       user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
