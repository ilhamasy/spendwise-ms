package dto

type ProfileResponse struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Currency        string `json:"currency"`
	Theme           string `json:"theme"`
	StartingBalance int64  `json:"startingBalance"`
	CreatedAt       string `json:"createdAt"`
}

type UpdateProfileRequest struct {
	Name            string `json:"name" validate:"omitempty,min=1,max=100"`
	Currency        string `json:"currency" validate:"omitempty"`
	Theme           string `json:"theme" validate:"omitempty,oneof=light dark system"`
	StartingBalance int64  `json:"startingBalance" validate:"omitempty,gte=0"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
}

type DeleteAccountRequest struct {
	Confirmation string `json:"confirmation" validate:"required,eq=DELETE"`
}

type ExportResponse struct {
	Profile       ProfileResponse          `json:"profile"`
	Transactions  []TransactionResponse    `json:"transactions"`
	Categories    []CategoryResponse       `json:"categories"`
	Goals         []GoalResponse           `json:"goals"`
	Contributions []ContributionResponse   `json:"contributions"`
}
