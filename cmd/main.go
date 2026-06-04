package main

import (
	"log"
	"spendwise-ms/internal/config"
	"spendwise-ms/internal/handler"
	"spendwise-ms/internal/middleware"
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

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", handler.Register)
			auth.POST("/login", handler.Login)
			auth.POST("/refresh", handler.RefreshToken)
		}

		transactions := api.Group("/transactions")
		transactions.Use(middleware.AuthRequired())
		{
			transactions.GET("", handler.GetTransactions)
			transactions.POST("", handler.CreateTransaction)
			transactions.GET("/:id", handler.GetTransactionByID)
			transactions.PUT("/:id", handler.UpdateTransaction)
			transactions.DELETE("/:id", handler.DeleteTransaction)
			transactions.POST("/sync", handler.SyncTransactions)
		}
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
