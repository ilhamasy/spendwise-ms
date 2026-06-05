package service

import (
	"encoding/json"
	"testing"

	"spendwise-ms/internal/config"
	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"
	"spendwise-ms/internal/repository"

	"github.com/google/uuid"
)

func setupSyncTestDB() {
	if config.DB == nil {
		config.InitDB(config.Load())
	}
	config.AutoMigrate(
		&model.User{}, &model.Transaction{}, &model.Category{},
		&model.SavingGoal{}, &model.GoalContribution{}, &model.Budget{},
	)
}

func createSyncTestUser() string {
	userID := uuid.New().String()
	config.DB.Create(&model.User{
		ID: userID, Name: "Sync Test", Email: "sync@test.com", Password: "hash",
	})
	return userID
}

func createSyncTestCategory(userID string) string {
	catID := uuid.New().String()
	config.DB.Create(&model.Category{
		ID: catID, UserID: userID, Name: "Food", Type: "expense",
	})
	return catID
}

func TestSync_EmptyChanges(t *testing.T) {
	setupSyncTestDB()
	userID := createSyncTestUser()
	svc := NewSyncService(config.DB)

	resp, err := svc.Sync(userID, dto.DataSyncRequest{
		LastSyncTimestamp: "",
		Changes:           []dto.SyncChange{},
	})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if resp.NewSyncTimestamp == "" {
		t.Error("NewSyncTimestamp should not be empty")
	}
}

func TestSync_CreateTransaction(t *testing.T) {
	setupSyncTestDB()
	userID := createSyncTestUser()
	catID := createSyncTestCategory(userID)
	svc := NewSyncService(config.DB)

	txnPayload, _ := json.Marshal(model.Transaction{
		Type: "expense", Amount: 50000, CategoryID: catID,
		OccurredAt: "2026-06-01", Note: "Sync test",
	})
	var payloadMap interface{}
	json.Unmarshal(txnPayload, &payloadMap)

	resp, err := svc.Sync(userID, dto.DataSyncRequest{
		Changes: []dto.SyncChange{{
			EntityType: "transaction",
			EntityID:   uuid.New().String(),
			Operation:  "CREATE",
			Payload:    payloadMap,
			Timestamp:  dto.NowTimestamp(),
		}},
	})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if len(resp.Conflicts) != 0 {
		t.Errorf("Expected 0 conflicts, got %d", len(resp.Conflicts))
	}
}

func TestSync_PullServerChanges(t *testing.T) {
	setupSyncTestDB()
	userID := createSyncTestUser()
	catID := createSyncTestCategory(userID)
	svc := NewSyncService(config.DB)

	txnRepo := repository.NewTransactionRepository(config.DB)
	txnRepo.Create(&model.Transaction{
		UserID: userID, Type: "expense", Amount: 10000,
		CategoryID: catID, OccurredAt: "2026-06-01",
	})

	resp, err := svc.Sync(userID, dto.DataSyncRequest{
		LastSyncTimestamp: "",
		Changes:           []dto.SyncChange{},
	})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	found := false
	for _, c := range resp.ServerChanges {
		if c.EntityType == "transaction" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected server transaction changes in response")
	}
}

func TestSync_ConflictResolution(t *testing.T) {
	setupSyncTestDB()
	userID := createSyncTestUser()
	catID := createSyncTestCategory(userID)
	svc := NewSyncService(config.DB)

	txnRepo := repository.NewTransactionRepository(config.DB)
	txn := &model.Transaction{
		UserID: userID, Type: "expense", Amount: 10000,
		CategoryID: catID, OccurredAt: "2026-06-01", Note: "Original",
	}
	txnRepo.Create(txn)

	updatePayload, _ := json.Marshal(map[string]interface{}{
		"type": "expense", "amount": 20000, "categoryId": catID,
		"occurredAt": "2026-06-01", "note": "Locally updated",
	})
	var payloadMap interface{}
	json.Unmarshal(updatePayload, &payloadMap)

	resp, err := svc.Sync(userID, dto.DataSyncRequest{
		Changes: []dto.SyncChange{{
			EntityType: "transaction",
			EntityID:   txn.ID,
			Operation:  "UPDATE",
			Payload:    payloadMap,
			Timestamp:  "2025-01-01T00:00:00Z",
		}},
	})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if len(resp.Conflicts) == 0 {
		t.Error("Expected conflict for outdated local change")
	}
}

func TestSync_CreateCategory(t *testing.T) {
	setupSyncTestDB()
	userID := createSyncTestUser()
	svc := NewSyncService(config.DB)

	catPayload, _ := json.Marshal(model.Category{
		Name: "Transport", Type: "expense", Icon: "car", Color: "#0000FF",
	})
	var payloadMap interface{}
	json.Unmarshal(catPayload, &payloadMap)

	resp, err := svc.Sync(userID, dto.DataSyncRequest{
		Changes: []dto.SyncChange{{
			EntityType: "category",
			EntityID:   uuid.New().String(),
			Operation:  "CREATE",
			Payload:    payloadMap,
			Timestamp:  dto.NowTimestamp(),
		}},
	})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if len(resp.Conflicts) != 0 {
		t.Errorf("Expected 0 conflicts, got %d", len(resp.Conflicts))
	}
}
