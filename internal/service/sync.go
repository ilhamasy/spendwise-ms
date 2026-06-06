package service

import (
	"encoding/json"
	"fmt"
	"time"

	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"
	"spendwise-ms/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SyncService struct {
	db         *gorm.DB
	txnRepo    *repository.TransactionRepository
	catRepo    *repository.CategoryRepository
	goalRepo   *repository.GoalRepository
	budgetRepo *repository.BudgetRepository
}

func NewSyncService(db *gorm.DB) *SyncService {
	return &SyncService{
		db:         db,
		txnRepo:    repository.NewTransactionRepository(db),
		catRepo:    repository.NewCategoryRepository(db),
		goalRepo:   repository.NewGoalRepository(db),
		budgetRepo: repository.NewBudgetRepository(db),
	}
}

func (s *SyncService) Sync(userID string, req dto.DataSyncRequest) (*dto.DataSyncResponse, error) {
	var conflicts []dto.SyncConflict
	processedChanges := 0

	for _, change := range req.Changes {
		conflict, err := s.applyChange(userID, change)
		if err != nil {
			continue
		}
		if conflict != nil {
			conflicts = append(conflicts, *conflict)
		}
		processedChanges++
	}

	serverChanges, err := s.getServerChanges(userID, req.LastSyncTimestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to get server changes: %w", err)
	}

	return &dto.DataSyncResponse{
		ServerChanges:   serverChanges,
		NewSyncTimestamp: dto.NowTimestamp(),
		Conflicts:       conflicts,
	}, nil
}

func (s *SyncService) applyChange(userID string, change dto.SyncChange) (*dto.SyncConflict, error) {
	switch change.EntityType {
	case "transaction":
		return s.applyTransactionChange(userID, change)
	case "category":
		return s.applyCategoryChange(userID, change)
	case "goal":
		return s.applyGoalChange(userID, change)
	case "budget":
		return s.applyBudgetChange(userID, change)
	default:
		return nil, fmt.Errorf("unknown entity type: %s", change.EntityType)
	}
}

func (s *SyncService) applyTransactionChange(userID string, change dto.SyncChange) (*dto.SyncConflict, error) {
	switch change.Operation {
	case "CREATE":
		txn, err := parsePayload[model.Transaction](change.Payload)
		if err != nil {
			return nil, err
		}
		txn.UserID = userID
		txn.ID = change.EntityID
		return nil, s.txnRepo.Create(txn)

	case "UPDATE":
		existing, err := s.txnRepo.FindByID(change.EntityID, userID)
		if err != nil {
			return nil, err
		}
		serverTS := existing.UpdatedAt.Format(time.RFC3339)
		if change.Timestamp > serverTS {
			txn, err := parsePayload[model.Transaction](change.Payload)
			if err != nil {
				return nil, err
			}
			existing.Type = txn.Type
			existing.Amount = txn.Amount
			existing.CategoryID = txn.CategoryID
			existing.OccurredAt = txn.OccurredAt
			existing.Note = txn.Note
			return nil, s.txnRepo.Update(existing)
		}
		return &dto.SyncConflict{
			EntityType:   "transaction",
			EntityID:     change.EntityID,
			LocalChange:  change,
			ServerChange: existing,
			Resolution:   "server_wins",
		}, nil

	case "DELETE":
		txn, err := s.txnRepo.FindByID(change.EntityID, userID)
		if err != nil {
			return nil, nil
		}
		serverTS := txn.UpdatedAt.Format(time.RFC3339)
		if change.Timestamp > serverTS {
			return nil, s.txnRepo.Delete(change.EntityID, userID)
		}
		return &dto.SyncConflict{
			EntityType:   "transaction",
			EntityID:     change.EntityID,
			LocalChange:  change,
			ServerChange: txn,
			Resolution:   "server_wins",
		}, nil
	}
	return nil, fmt.Errorf("unknown operation: %s", change.Operation)
}

func (s *SyncService) applyCategoryChange(userID string, change dto.SyncChange) (*dto.SyncConflict, error) {
	switch change.Operation {
	case "CREATE":
		cat, err := parsePayload[model.Category](change.Payload)
		if err != nil {
			return nil, err
		}
		cat.UserID = userID
		cat.ID = change.EntityID
		return nil, s.catRepo.Create(cat)
	case "UPDATE":
		cat, err := s.catRepo.FindByID(change.EntityID, userID)
		if err != nil {
			return nil, err
		}
		updated, err := parsePayload[model.Category](change.Payload)
		if err != nil {
			return nil, err
		}
		cat.Name = updated.Name
		cat.Icon = updated.Icon
		cat.Color = updated.Color
		payloadMap := change.Payload.(map[string]interface{})
		if status, ok := payloadMap["status"].(string); ok && status == "archived" {
			now := time.Now()
			cat.DeletedAt = &now
		}
		return nil, s.catRepo.Update(cat)
	case "DELETE":
		s.catRepo.Delete(change.EntityID, userID)
		return nil, nil
	}
	return nil, fmt.Errorf("unknown operation: %s", change.Operation)
}

func (s *SyncService) applyGoalChange(userID string, change dto.SyncChange) (*dto.SyncConflict, error) {
	switch change.Operation {
	case "CREATE":
		goal, err := parsePayload[model.SavingGoal](change.Payload)
		if err != nil {
			return nil, err
		}
		goal.UserID = userID
		goal.ID = change.EntityID
		return nil, s.goalRepo.Create(goal)
	case "UPDATE":
		goal, err := s.goalRepo.FindByID(change.EntityID, userID)
		if err != nil {
			return nil, err
		}
		updated, err := parsePayload[model.SavingGoal](change.Payload)
		if err != nil {
			return nil, err
		}
		goal.Name = updated.Name
		goal.TargetAmount = updated.TargetAmount
		goal.CurrentSaved = updated.CurrentSaved
		goal.TargetDate = updated.TargetDate
		goal.Status = updated.Status
		return nil, s.goalRepo.Update(goal)
	case "DELETE":
		s.goalRepo.Delete(change.EntityID, userID)
		return nil, nil
	}
	return nil, fmt.Errorf("unknown operation: %s", change.Operation)
}

func (s *SyncService) applyBudgetChange(userID string, change dto.SyncChange) (*dto.SyncConflict, error) {
	switch change.Operation {
	case "CREATE":
		budget, err := parsePayload[model.Budget](change.Payload)
		if err != nil {
			return nil, err
		}
		budget.UserID = userID
		budget.ID = change.EntityID
		return nil, s.budgetRepo.Create(budget)
	case "UPDATE":
		budget, err := s.budgetRepo.FindByID(change.EntityID, userID)
		if err != nil {
			return nil, err
		}
		updated, err := parsePayload[model.Budget](change.Payload)
		if err != nil {
			return nil, err
		}
		budget.Name = updated.Name
		budget.Amount = updated.Amount
		budget.Period = updated.Period
		budget.CategoryID = updated.CategoryID
		return nil, s.budgetRepo.Update(budget)
	case "DELETE":
		s.budgetRepo.Delete(change.EntityID, userID)
		return nil, nil
	}
	return nil, fmt.Errorf("unknown operation: %s", change.Operation)
}

func (s *SyncService) getServerChanges(userID, since string) ([]dto.SyncChangeItem, error) {
	var changes []dto.SyncChangeItem

	txns, _, _ := s.txnRepo.FindAll(repository.TransactionFilter{UserID: userID, Page: 1, Limit: 1000})
	for _, t := range txns {
		timestamp := t.UpdatedAt.Format(time.RFC3339)
		if since != "" && timestamp <= since {
			continue
		}
		changes = append(changes, dto.SyncChangeItem{
			EntityType: "transaction",
			EntityID:   t.ID,
			Data: map[string]interface{}{
				"id":         t.ID,
				"type":       t.Type,
				"amount":     t.Amount,
				"categoryId": t.CategoryID,
				"occurredAt": t.OccurredAt,
				"note":       t.Note,
				"createdAt":  t.CreatedAt.Format(time.RFC3339),
				"updatedAt":  t.UpdatedAt.Format(time.RFC3339),
			},
			Timestamp: timestamp,
		})
	}

	cats, _ := s.catRepo.FindAllForSync(userID)
	for _, c := range cats {
		if c.DeletedAt != nil {
			changes = append(changes, dto.SyncChangeItem{
				EntityType: "category",
				EntityID:   c.ID,
				Data:       map[string]interface{}{"id": c.ID, "isDeleted": true},
				Timestamp:  c.UpdatedAt.Format(time.RFC3339),
			})
			continue
		}
		timestamp := c.UpdatedAt.Format(time.RFC3339)
		if since != "" && timestamp <= since {
			continue
		}
		cData := map[string]interface{}{
			"id":        c.ID,
			"name":      c.Name,
			"type":      c.Type,
			"icon":      c.Icon,
			"color":     c.Color,
			"isDefault": c.IsDefault,
			"updatedAt": c.UpdatedAt.Format(time.RFC3339),
		}
		changes = append(changes, dto.SyncChangeItem{
			EntityType: "category",
			EntityID:   c.ID,
			Data:       cData,
			Timestamp:  timestamp,
		})
	}

	goals, _ := s.goalRepo.FindAll(userID, "")
	for _, g := range goals {
		timestamp := g.UpdatedAt.Format(time.RFC3339)
		if since != "" && timestamp <= since {
			continue
		}
		changes = append(changes, dto.SyncChangeItem{
			EntityType: "goal",
			EntityID:   g.ID,
			Data: map[string]interface{}{
				"id":           g.ID,
				"name":         g.Name,
				"targetAmount": g.TargetAmount,
				"currentSaved": g.CurrentSaved,
				"targetDate":   g.TargetDate,
				"status":       g.Status,
				"createdAt":    g.CreatedAt.Format(time.RFC3339),
				"updatedAt":    g.UpdatedAt.Format(time.RFC3339),
			},
			Timestamp: timestamp,
		})
	}

	budgets, _ := s.budgetRepo.FindAll(userID, "")
	for _, b := range budgets {
		timestamp := b.UpdatedAt.Format(time.RFC3339)
		if since != "" && timestamp <= since {
			continue
		}
		changes = append(changes, dto.SyncChangeItem{
			EntityType: "budget",
			EntityID:   b.ID,
			Data: map[string]interface{}{
				"id":         b.ID,
				"name":       b.Name,
				"amount":     b.Amount,
				"period":     b.Period,
				"categoryId": b.CategoryID,
				"createdAt":  b.CreatedAt.Format(time.RFC3339),
				"updatedAt":  b.UpdatedAt.Format(time.RFC3339),
			},
			Timestamp: timestamp,
		})
	}

	return changes, nil
}

func parsePayload[T any](payload interface{}) (*T, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse payload: %w", err)
	}
	return &result, nil
}

func generateID() string {
	return uuid.New().String()
}
