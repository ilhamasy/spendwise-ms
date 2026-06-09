package repository

import (
	"os"
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/model"
)

func TestMain(m *testing.M) {
	os.Setenv("DB_NAME", "spendwise_test_db")
	config.DB = nil
	cfg := config.Load()
	config.InitDB(cfg)
	config.AutoMigrate(&model.User{}, &model.Transaction{}, &model.Category{}, &model.SavingGoal{}, &model.GoalContribution{}, &model.Budget{})
	code := m.Run()
	os.Exit(code)
}
