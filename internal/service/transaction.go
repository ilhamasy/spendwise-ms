package service

import (
	"errors"
	"fmt"
	"time"

	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"
	"spendwise-ms/internal/repository"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var txnValidate = validator.New()

type TransactionService struct {
	repo repository.ITransactionRepository
}

func NewTransactionService(db *gorm.DB) *TransactionService {
	return &TransactionService{
		repo: repository.NewTransactionRepository(db),
	}
}

func NewTransactionServiceWithRepo(r repository.ITransactionRepository) *TransactionService {
	return &TransactionService{repo: r}
}

func (s *TransactionService) CreateTransaction(userID string, req dto.CreateTransactionRequest) (*dto.TransactionResponse, error) {
	if err := txnValidate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	cat, err := s.repo.FindCategoryByID(req.CategoryID, userID)
	if err != nil {
		return nil, fmt.Errorf("category not found")
	}

	if _, err := time.Parse("2006-01-02", req.OccurredAt); err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD")
	}
	today := time.Now().Format("2006-01-02")
	if req.OccurredAt > today {
		return nil, fmt.Errorf("future transactions are not allowed")
	}

	tx := model.Transaction{
		UserID:     userID,
		Type:       req.Type,
		Amount:     req.Amount,
		CategoryID: cat.ID,
		OccurredAt: req.OccurredAt,
		Note:       dto.SanitizeString(req.Note),
	}

	if err := s.repo.Create(&tx); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	resp := toResponse(tx)
	return &resp, nil
}

func (s *TransactionService) GetTransactions(userID string, filter repository.TransactionFilter) (*dto.TransactionListResponse, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	filter.UserID = userID

	transactions, total, err := s.repo.FindAll(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve transactions: %w", err)
	}

	var data []dto.TransactionResponse
	for _, tx := range transactions {
		data = append(data, toResponse(tx))
	}

	return &dto.TransactionListResponse{
		Data: data,
		Meta: dto.Meta{
			Total:      total,
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalPages: repository.CalculateTotalPages(total, filter.Limit),
		},
	}, nil
}

func (s *TransactionService) GetTransactionByID(userID, id string) (*dto.TransactionResponse, error) {
	tx, err := s.repo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transaction not found")
		}
		return nil, fmt.Errorf("failed to retrieve transaction: %w", err)
	}

	resp := toResponse(*tx)
	return &resp, nil
}

func (s *TransactionService) UpdateTransaction(userID, id string, req dto.UpdateTransactionRequest) (*dto.TransactionResponse, error) {
	if err := txnValidate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	existing, err := s.repo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transaction not found")
		}
		return nil, fmt.Errorf("failed to retrieve transaction: %w", err)
	}

	if req.Type != "" {
		existing.Type = req.Type
	}
	if req.Amount > 0 {
		existing.Amount = req.Amount
	}
	if req.CategoryID != "" {
		cat, err := s.repo.FindCategoryByID(req.CategoryID, userID)
		if err != nil {
			return nil, fmt.Errorf("category not found")
		}
		existing.CategoryID = cat.ID
	}
	if req.OccurredAt != "" {
		if _, err := time.Parse("2006-01-02", req.OccurredAt); err != nil {
			return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD")
		}
		today := time.Now().Format("2006-01-02")
		if req.OccurredAt > today {
			return nil, fmt.Errorf("future transactions are not allowed")
		}
		existing.OccurredAt = req.OccurredAt
	}
	if req.Note != "" {
		existing.Note = dto.SanitizeString(req.Note)
	}

	if err := s.repo.Update(existing); err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	resp := toResponse(*existing)
	return &resp, nil
}

func (s *TransactionService) DeleteTransaction(userID, id string) error {
	err := s.repo.Delete(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("transaction not found")
		}
		return fmt.Errorf("failed to delete transaction: %w", err)
	}
	return nil
}

func (s *TransactionService) SyncTransactions(userID string, req dto.SyncRequest) (*dto.SyncResponse, error) {
	if len(req.Transactions) == 0 {
		return nil, fmt.Errorf("at least one transaction is required")
	}

	var transactions []model.Transaction
	var failedItems []dto.SyncFailedItem

	for i, item := range req.Transactions {
		if err := txnValidate.Struct(item); err != nil {
			failedItems = append(failedItems, dto.SyncFailedItem{Index: i, Error: err.Error()})
			continue
		}

		cat, err := s.repo.FindCategoryByID(item.CategoryID, userID)
		if err != nil {
			failedItems = append(failedItems, dto.SyncFailedItem{Index: i, Error: "category not found"})
			continue
		}

		if _, err := time.Parse("2006-01-02", item.OccurredAt); err != nil {
			failedItems = append(failedItems, dto.SyncFailedItem{Index: i, Error: "invalid date format"})
			continue
		}
		today := time.Now().Format("2006-01-02")
		if item.OccurredAt > today {
			failedItems = append(failedItems, dto.SyncFailedItem{Index: i, Error: "future transactions are not allowed"})
			continue
		}

		transactions = append(transactions, model.Transaction{
			UserID:     userID,
			Type:       item.Type,
			Amount:     item.Amount,
			CategoryID: cat.ID,
			OccurredAt: item.OccurredAt,
			Note:       item.Note,
		})
	}

	synced, failedIndices, failedErrors := s.repo.BatchCreate(transactions)
	for i, idx := range failedIndices {
		failedItems = append(failedItems, dto.SyncFailedItem{Index: idx, Error: failedErrors[i]})
	}

	return &dto.SyncResponse{
		Synced: synced,
		Failed: failedItems,
	}, nil
}

func toResponse(tx model.Transaction) dto.TransactionResponse {
	return dto.TransactionResponse{
		ID:         tx.ID,
		Type:       tx.Type,
		Amount:     tx.Amount,
		CategoryID: tx.CategoryID,
		OccurredAt: tx.OccurredAt,
		Note:       tx.Note,
		CreatedAt:  tx.CreatedAt,
		UpdatedAt:  tx.UpdatedAt,
	}
}
