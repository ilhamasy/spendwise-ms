package dto

import "time"

type CreateBudgetRequest struct {
	Name       string `json:"name" validate:"required,min=1,max=100"`
	Amount     int64  `json:"amount" validate:"required,gt=0"`
	Period     string `json:"period" validate:"required,oneof=daily weekly monthly yearly"`
	CategoryID string `json:"categoryId" validate:"required,min=1"`
}

type UpdateBudgetRequest struct {
	Name       string `json:"name" validate:"omitempty,min=1,max=100"`
	Amount     int64  `json:"amount" validate:"omitempty,gt=0"`
	Period     string `json:"period" validate:"omitempty,oneof=daily weekly monthly yearly"`
	CategoryID string `json:"categoryId" validate:"omitempty,min=1"`
}

type BudgetResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Amount     int64     `json:"amount"`
	Period     string    `json:"period"`
	CategoryID string    `json:"categoryId"`
	Spent      int64     `json:"spent"`
	Remaining  int64     `json:"remaining"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
