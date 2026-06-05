package service

import (
	"os"
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/model"
)

func TestMain(m *testing.M) {
	config.InitDB(config.Load())
	config.AutoMigrate(
		&model.User{},
		&model.Transaction{},
		&model.Category{},
		&model.SavingGoal{},
		&model.GoalContribution{},
		&model.Budget{},
	)
	os.Exit(m.Run())
}
