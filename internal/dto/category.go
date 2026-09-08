package dto

type CreateCategoryRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=50"`
	Type  string `json:"type" validate:"required,oneof=income expense"`
	Icon  string `json:"icon" validate:"max=10"`
	Color string `json:"color" validate:"max=10"`
}

type UpdateCategoryRequest struct {
	Name  string `json:"name" validate:"omitempty,min=1,max=50"`
	Icon  string `json:"icon" validate:"max=10"`
	Color string `json:"color" validate:"max=10"`
}

type CategoryResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Icon      string `json:"icon"`
	Color     string `json:"color"`
	IsDefault bool   `json:"isDefault"`
}

type ReassignAndDeleteRequest struct {
	ReassignToCategoryID string `json:"reassignToCategoryId" validate:"required,min=1"`
}
