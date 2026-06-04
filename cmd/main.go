package main

import (
	"log"
	"spendwise-ms/internal/config"
	"spendwise-ms/internal/handler"
	"spendwise-ms/internal/model"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	config.InitDB(cfg)

	config.AutoMigrate(
		&model.User{},
		&model.Transaction{},
		&model.Category{},
		&model.SavingGoal{},
		&model.GoalContribution{},
		&model.Budget{},
	)

	r := gin.Default()

	r.Use(config.CORS())

	r.GET("/health", handler.HealthCheck)

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
