package main

import (
	"log"
	"net/http"
	"spendwise-ms/internal/config"
	"spendwise-ms/internal/handler"
	"spendwise-ms/internal/middleware"
	"spendwise-ms/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not loaded: %v", err)
	}

	cfg := config.Load()

	config.InitDB(cfg)

	if err := config.RunMigrations(cfg); err != nil {
		log.Printf("Warning: Migration failed: %v", err)
		log.Println("Falling back to AutoMigrate...")
		config.AutoMigrate(
			&model.User{},
			&model.Transaction{},
			&model.Category{},
			&model.SavingGoal{},
			&model.GoalContribution{},
			&model.Budget{},
		)
	}

	r := gin.Default()

	r.Use(middleware.RateLimitGlobal())

	r.GET("/health", handler.HealthCheck)

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		auth.Use(middleware.RateLimitAuth())
		{
			auth.POST("/register", handler.Register)
			auth.POST("/login", handler.Login)
			auth.POST("/refresh", handler.RefreshToken)
			auth.POST("/logout", handler.Logout)
		}

		users := v1.Group("/users/me")
		users.Use(middleware.AuthRequired())
		{
			users.GET("", handler.GetProfile)
			users.PUT("", handler.UpdateProfile)
			users.PUT("/password", handler.ChangePassword)
			users.DELETE("", handler.DeleteAccount)
			users.GET("/export", handler.ExportUserData)
		}

		transactions := v1.Group("/transactions")
		transactions.Use(middleware.AuthRequired())
		{
			transactions.GET("", handler.GetTransactions)
			transactions.POST("", handler.CreateTransaction)
			transactions.GET("/:id", handler.GetTransactionByID)
			transactions.PUT("/:id", handler.UpdateTransaction)
			transactions.DELETE("/:id", handler.DeleteTransaction)
			transactions.POST("/sync", handler.SyncTransactions)
		}

		categories := v1.Group("/categories")
		categories.Use(middleware.AuthRequired())
		{
			categories.GET("", handler.GetCategories)
			categories.POST("", handler.CreateCategory)
			categories.PUT("/:id", handler.UpdateCategory)
			categories.DELETE("/:id", handler.DeleteCategory)
		}

		goals := v1.Group("/goals")
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

		budgets := v1.Group("/budgets")
		budgets.Use(middleware.AuthRequired())
		{
			budgets.GET("", handler.GetBudgets)
			budgets.POST("", handler.CreateBudget)
			budgets.GET("/:id", handler.GetBudgetByID)
			budgets.PUT("/:id", handler.UpdateBudget)
			budgets.DELETE("/:id", handler.DeleteBudget)
		}

		sync := v1.Group("/sync")
		sync.Use(middleware.AuthRequired())
		{
			sync.POST("", handler.SyncData)
		}

		docs := v1.Group("/docs")
		{
			docs.GET("", func(c *gin.Context) {
				c.Redirect(http.StatusMovedPermanently, "/api/v1/docs/index.html")
			})
		}
	}
	r.Static("/api/v1/docs", "./docs")

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
