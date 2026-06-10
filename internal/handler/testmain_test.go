package handler

import (
	"os"
	"testing"

	"gorm.io/gorm/logger"
	"spendwise-ms/internal/config"
	"spendwise-ms/internal/model"
)

func TestMain(m *testing.M) {
	os.Setenv("DB_NAME", "spendwise_test_db")
	config.DB = nil
	cfg := config.Load()
	config.InitDB(cfg)
	config.DB.Logger = logger.Default.LogMode(logger.Silent)
	config.AutoMigrate(&model.User{}, &model.Transaction{}, &model.Category{}, &model.SavingGoal{}, &model.GoalContribution{}, &model.Budget{})
	code := m.Run()
	sqlDB, err := config.DB.DB()
	if err == nil {
		sqlDB.Close()
	}
	os.Exit(code)
}
