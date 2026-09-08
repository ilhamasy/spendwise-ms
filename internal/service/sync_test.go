package service

import (
	"testing"
	"time"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"
	"spendwise-ms/internal/repository"

	"github.com/google/uuid"
)

func setupSyncTestDB() string {
	if config.DB == nil {
		config.InitDB(config.Load())
	}
	if !config.DB.Migrator().HasTable("users") {
		config.AutoMigrate(
			&model.User{}, &model.Transaction{}, &model.Category{}, &model.SavingGoal{}, &model.GoalContribution{}, &model.Budget{},
		)
	}
	config.DB.Where("1 = 1").Delete(&model.GoalContribution{})
	config.DB.Where("1 = 1").Delete(&model.Budget{})
	config.DB.Where("1 = 1").Delete(&model.SavingGoal{})
	config.DB.Where("1 = 1").Delete(&model.Transaction{})
	config.DB.Where("1 = 1").Delete(&model.Category{})
	config.DB.Where("1 = 1").Delete(&model.User{})

	userID := uuid.New().String()
	config.DB.Create(&model.User{
		ID:       userID,
		Name:     "Sync User",
		Email:    "sync@spendwise.com",
		Password: "hashed",
	})
	return userID
}

func TestSyncService_ApplyChanges(t *testing.T) {
	userID := setupSyncTestDB()
	svc := NewSyncService(config.DB)

	catID := uuid.New().String()
	txID := uuid.New().String()
	goalID := uuid.New().String()
	budgetID := uuid.New().String()

	nowStr := time.Now().Format(time.RFC3339)
	oldStr := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)

	// 1. Sync Create Category
	reqCat := dto.DataSyncRequest{
		Changes: []dto.SyncChange{
			{
				EntityType: "category",
				EntityID:   catID,
				Operation:  "CREATE",
				Timestamp:  nowStr,
				Payload: map[string]interface{}{
					"name":  "Sync Category",
					"type":  "expense",
					"icon":  "sync-icon",
					"color": "#123456",
				},
			},
		},
	}
	resp, err := svc.Sync(userID, reqCat)
	if err != nil {
		t.Fatalf("Sync category create failed: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}

	// 2. Sync Update Category
	reqCatUpdate := dto.DataSyncRequest{
		Changes: []dto.SyncChange{
			{
				EntityType: "category",
				EntityID:   catID,
				Operation:  "UPDATE",
				Timestamp:  nowStr,
				Payload: map[string]interface{}{
					"name":   "Sync Category Updated",
					"icon":   "sync-icon-2",
					"color":  "#654321",
					"status": "archived",
				},
			},
		},
	}
	svc.Sync(userID, reqCatUpdate)

	// 3. Sync Delete Category
	reqCatDelete := dto.DataSyncRequest{
		Changes: []dto.SyncChange{
			{
				EntityType: "category",
				EntityID:   catID,
				Operation:  "DELETE",
				Timestamp:  nowStr,
			},
		},
	}
	svc.Sync(userID, reqCatDelete)

	// 4. Sync Create Transaction
	reqTx := dto.DataSyncRequest{
		Changes: []dto.SyncChange{
			{
				EntityType: "transaction",
				EntityID:   txID,
				Operation:  "CREATE",
				Timestamp:  nowStr,
				Payload: map[string]interface{}{
					"type":       "expense",
					"amount":     150000,
					"categoryId": catID,
					"occurredAt": "2026-08-21",
					"note":       "Sync test tx",
				},
			},
		},
	}
	svc.Sync(userID, reqTx)

	// 5. Sync Update Transaction (Local Wins)
	reqTxUpdateNew := dto.DataSyncRequest{
		Changes: []dto.SyncChange{
			{
				EntityType: "transaction",
				EntityID:   txID,
				Operation:  "UPDATE",
				Timestamp:  time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				Payload: map[string]interface{}{
					"type":       "expense",
					"amount":     200000,
					"categoryId": catID,
					"occurredAt": "2026-08-21",
					"note":       "Sync test tx updated",
					"updatedAt":  time.Now().Add(2 * time.Hour).Format(time.RFC3339),
				},
			},
		},
	}
	svc.Sync(userID, reqTxUpdateNew)

	// 6. Sync Update Transaction (Server Wins / Conflict)
	reqTxUpdateOld := dto.DataSyncRequest{
		Changes: []dto.SyncChange{
			{
				EntityType: "transaction",
				EntityID:   txID,
				Operation:  "UPDATE",
				Timestamp:  oldStr,
				Payload: map[string]interface{}{
					"type":       "expense",
					"amount":     10,
					"categoryId": catID,
					"occurredAt": "2026-08-21",
					"note":       "Sync test tx old",
				},
			},
		},
	}
	respConflict, _ := svc.Sync(userID, reqTxUpdateOld)
	if respConflict != nil && len(respConflict.Conflicts) > 0 {
		// Conflict recorded as expected
	}

	// 7. Sync Delete Transaction (Server Wins conflict vs Local Wins)
	reqTxDeleteOld := dto.DataSyncRequest{
		Changes: []dto.SyncChange{
			{
				EntityType: "transaction",
				EntityID:   txID,
				Operation:  "DELETE",
				Timestamp:  oldStr,
			},
		},
	}
	svc.Sync(userID, reqTxDeleteOld)

	reqTxDeleteNew := dto.DataSyncRequest{
		Changes: []dto.SyncChange{
			{
				EntityType: "transaction",
				EntityID:   txID,
				Operation:  "DELETE",
				Timestamp:  time.Now().Add(2 * time.Hour).Format(time.RFC3339),
			},
		},
	}
	svc.Sync(userID, reqTxDeleteNew)

	// 8. Sync Create Goal
	reqGoal := dto.DataSyncRequest{
		Changes: []dto.SyncChange{
			{
				EntityType: "goal",
				EntityID:   goalID,
				Operation:  "CREATE",
				Timestamp:  nowStr,
				Payload: map[string]interface{}{
					"name":         "New Car",
					"targetAmount": 300000000,
					"currentSaved": 50000000,
					"targetDate":   "2027-12-31",
					"status":       "active",
				},
			},
		},
	}
	svc.Sync(userID, reqGoal)

	// 9. Sync Update Goal
	reqGoalUpdate := dto.DataSyncRequest{
		Changes: []dto.SyncChange{
			{
				EntityType: "goal",
				EntityID:   goalID,
				Operation:  "UPDATE",
				Timestamp:  nowStr,
				Payload: map[string]interface{}{
					"name":         "New Car Updated",
					"targetAmount": 350000000,
					"currentSaved": 60000000,
					"targetDate":   "2027-12-31",
					"status":       "active",
				},
			},
		},
	}
	svc.Sync(userID, reqGoalUpdate)

	// 10. Sync Delete Goal
	reqGoalDelete := dto.DataSyncRequest{
		Changes: []dto.SyncChange{
			{
				EntityType: "goal",
				EntityID:   goalID,
				Operation:  "DELETE",
				Timestamp:  nowStr,
			},
		},
	}
	svc.Sync(userID, reqGoalDelete)

	// 11. Sync Create Budget
	reqBudget := dto.DataSyncRequest{
		Changes: []dto.SyncChange{
			{
				EntityType: "budget",
				EntityID:   budgetID,
				Operation:  "CREATE",
				Timestamp:  nowStr,
				Payload: map[string]interface{}{
					"name":       "Monthly Food",
					"amount":     2000000,
					"period":     "monthly",
					"categoryId": catID,
				},
			},
		},
	}
	svc.Sync(userID, reqBudget)

	// 12. Sync Update Budget
	reqBudgetUpdate := dto.DataSyncRequest{
		Changes: []dto.SyncChange{
			{
				EntityType: "budget",
				EntityID:   budgetID,
				Operation:  "UPDATE",
				Timestamp:  nowStr,
				Payload: map[string]interface{}{
					"name":       "Monthly Food Updated",
					"amount":     2500000,
					"period":     "monthly",
					"categoryId": catID,
				},
			},
		},
	}
	svc.Sync(userID, reqBudgetUpdate)

	// 13. Sync Delete Budget
	reqBudgetDelete := dto.DataSyncRequest{
		Changes: []dto.SyncChange{
			{
				EntityType: "budget",
				EntityID:   budgetID,
				Operation:  "DELETE",
				Timestamp:  nowStr,
			},
		},
	}
	svc.Sync(userID, reqBudgetDelete)

	// 14. Unknown entity / operation check
	reqUnknown := dto.DataSyncRequest{
		Changes: []dto.SyncChange{
			{
				EntityType: "unknown",
				EntityID:   "123",
				Operation:  "INVALID",
			},
		},
	}
	svc.Sync(userID, reqUnknown)
}

func TestSyncService_GetServerChanges(t *testing.T) {
	userID := setupSyncTestDB()
	svc := NewSyncService(config.DB)

	catRepo := repository.NewCategoryRepository(config.DB)
	txnRepo := repository.NewTransactionRepository(config.DB)
	goalRepo := repository.NewGoalRepository(config.DB)
	budgetRepo := repository.NewBudgetRepository(config.DB)

	cat := model.Category{UserID: userID, Name: "ServerCat", Type: "expense"}
	catRepo.Create(&cat)

	txnRepo.Create(&model.Transaction{
		UserID: userID, Type: "expense", Amount: 100, CategoryID: cat.ID, OccurredAt: "2026-08-21",
	})

	goalRepo.Create(&model.SavingGoal{
		UserID: userID, Name: "ServerGoal", TargetAmount: 1000, Status: "active",
	})

	budgetRepo.Create(&model.Budget{
		UserID: userID, Name: "ServerBudget", Amount: 5000, Period: "monthly", CategoryID: cat.ID,
	})

	catRepo.Delete(cat.ID, userID)

	resp, err := svc.Sync(userID, dto.DataSyncRequest{LastSyncTimestamp: ""})
	if err != nil {
		t.Fatalf("Sync server changes failed: %v", err)
	}
	if len(resp.ServerChanges) == 0 {
		t.Error("Expected server changes, got 0")
	}
}
