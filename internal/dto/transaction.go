package dto

import "time"

type CreateTransactionRequest struct {
	Type       string `json:"type" validate:"required,oneof=income expense"`
	Amount     int64  `json:"amount" validate:"required,gt=0,lte=1000000000000"`
	CategoryID string `json:"categoryId" validate:"required,min=1"`
	OccurredAt string `json:"occurredAt" validate:"required,datetime=2006-01-02"`
	Note       string `json:"note" validate:"max=200"`
}

type UpdateTransactionRequest struct {
	Type       string `json:"type" validate:"omitempty,oneof=income expense"`
	Amount     int64  `json:"amount" validate:"omitempty,gt=0,lte=1000000000000"`
	CategoryID string `json:"categoryId" validate:"omitempty,min=1"`
	OccurredAt string `json:"occurredAt" validate:"omitempty,datetime=2006-01-02"`
	Note       string `json:"note" validate:"max=200"`
}

type TransactionResponse struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Amount     int64  `json:"amount"`
	CategoryID string `json:"categoryId"`
	OccurredAt string `json:"occurredAt"`
	Note       string `json:"note"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type TransactionListResponse struct {
	Data []TransactionResponse `json:"data"`
	Meta Meta                  `json:"meta"`
}

type Meta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int64 `json:"totalPages"`
}

type SyncTransactionRequest struct {
	Type       string `json:"type" validate:"required,oneof=income expense"`
	Amount     int64  `json:"amount" validate:"required,gt=0,lte=1000000000000"`
	CategoryID string `json:"categoryId" validate:"required,min=1"`
	OccurredAt string `json:"occurredAt" validate:"required,datetime=2006-01-02"`
	Note       string `json:"note" validate:"max=200"`
}

type SyncRequest struct {
	Transactions []SyncTransactionRequest `json:"transactions" validate:"required,min=1"`
}

type SyncFailedItem struct {
	Index int    `json:"index"`
	Error string `json:"error"`
}

type SyncResponse struct {
	Synced int              `json:"synced"`
	Failed []SyncFailedItem `json:"failed"`
}
