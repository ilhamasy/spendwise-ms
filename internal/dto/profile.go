package dto

type ProfileResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
}

type UpdateProfileRequest struct {
	Name  string `json:"name" validate:"omitempty,min=1,max=100"`
	Email string `json:"email" validate:"omitempty,email,max=100"`
}
