package repository

import (
	"spendwise-ms/internal/model"
)

// ITransactionRepository defines the contract for transaction data access
type ITransactionRepository interface {
	Create(tx *model.Transaction) error
	FindByID(id, userID string) (*model.Transaction, error)
	FindAll(filter TransactionFilter) ([]model.Transaction, int64, error)
	Update(tx *model.Transaction) error
	Delete(id, userID string) error
	BatchCreate(transactions []model.Transaction) (int, []int, []string)
	FindCategoryByID(id, userID string) (*model.Category, error)
}

// ICategoryRepository defines the contract for category data access
type ICategoryRepository interface {
	Create(cat *model.Category) error
	FindByID(id, userID string) (*model.Category, error)
	FindAll(userID, filterType string) ([]model.Category, error)
	FindAllForSync(userID string) ([]model.Category, error)
	Update(cat *model.Category) error
	Delete(id, userID string) error
	FindByNameAndType(name, catType, userID string) (*model.Category, error)
	CountTransactionsByCategoryID(categoryID, userID string) (int64, error)
	ReassignTransactions(fromCategoryID, toCategoryID, userID string) error
}

// IGoalRepository defines the contract for goal data access
type IGoalRepository interface {
	Create(goal *model.SavingGoal) error
	FindByID(id, userID string) (*model.SavingGoal, error)
	FindAll(userID, status string) ([]model.SavingGoal, error)
	Update(goal *model.SavingGoal) error
	UpdateStatus(id, userID, status string) error
	Delete(id, userID string) error
	CreateContribution(c *model.GoalContribution) error
	FindContributionsByGoalID(goalID string, page, limit int) ([]model.GoalContribution, int64, error)
}

// IBudgetRepository defines the contract for budget data access
type IBudgetRepository interface {
	Create(b *model.Budget) error
	FindByID(id, userID string) (*model.Budget, error)
	FindAll(userID, period string) ([]model.Budget, error)
	Update(b *model.Budget) error
	Delete(id, userID string) error
	GetSpentAmount(userID, categoryID, period string) (int64, error)
}
