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

		categories := api.Group("/categories")
		categories.Use(middleware.AuthRequired())
		{
			categories.GET("", handler.GetCategories)
			categories.POST("", handler.CreateCategory)
			categories.PUT("/:id", handler.UpdateCategory)
			categories.DELETE("/:id", handler.DeleteCategory)
		}

		goals := api.Group("/goals")
		goals.Use(middleware.AuthRequired())
		{
			goals.GET("", handler.GetGoals)
			goals.POST("", handler.CreateGoal)
			goals.GET("/:id", handler.GetGoalByID)
			goals.PUT("/:id", handler.UpdateGoal)
			goals.PATCH("/:id/archive", handler.ArchiveGoal)
			goals.PATCH("/:id/unarchive", handler.UnarchiveGoal)
			goals.DELETE("/:id", handler.DeleteGoal)
			goals.POST("/:id/contributions", handler.AddContribution)
			goals.GET("/:id/contributions", handler.GetContributionHistory)
		}

		budgets := api.Group("/budgets")
		budgets.Use(middleware.AuthRequired())
		{
			budgets.GET("", handler.GetBudgets)
			budgets.POST("", handler.CreateBudget)
			budgets.GET("/:id", handler.GetBudgetByID)
			budgets.PUT("/:id", handler.UpdateBudget)
			budgets.DELETE("/:id", handler.DeleteBudget)
		}

		sync := api.Group("/sync")
		sync.Use(middleware.AuthRequired())
		{
			sync.POST("", handler.SyncData)
		}
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
