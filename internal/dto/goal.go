package dto

import "time"

type CreateGoalRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=100"`
	TargetAmount int64  `json:"targetAmount" validate:"required,gt=0,lte=1000000000000"`
	CurrentSaved int64  `json:"currentSaved" validate:"omitempty,gte=0,lte=1000000000000"`
	TargetDate   string `json:"targetDate" validate:"omitempty,datetime=2006-01-02"`
}

type UpdateGoalRequest struct {
	Name         string `json:"name" validate:"omitempty,min=1,max=100"`
	TargetAmount int64  `json:"targetAmount" validate:"omitempty,gt=0,lte=1000000000000"`
	CurrentSaved int64  `json:"currentSaved" validate:"omitempty,gte=0,lte=1000000000000"`
	TargetDate   string `json:"targetDate" validate:"omitempty,datetime=2006-01-02"`
}

type GoalResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	TargetAmount int64     `json:"targetAmount"`
	CurrentSaved int64     `json:"currentSaved"`
	TargetDate   string    `json:"targetDate"`
	Status       string    `json:"status"`
	Progress     float64   `json:"progress"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type GoalListResponse struct {
	Data []GoalResponse `json:"data"`
}

type CreateContributionRequest struct {
	Amount int64  `json:"amount" validate:"required,gt=0,lte=1000000000000"`
	Note   string `json:"note" validate:"max=200"`
	Date   string `json:"date" validate:"omitempty,datetime=2006-01-02"`
}

type ContributionResponse struct {
	ID        string    `json:"id"`
	GoalID    string    `json:"goalId"`
	Amount    int64     `json:"amount"`
	Note      string    `json:"note"`
	Date      string    `json:"date"`
	CreatedAt time.Time `json:"createdAt"`
}

type ContributionListResponse struct {
	Data []ContributionResponse `json:"data"`
	Meta Meta                   `json:"meta"`
}
