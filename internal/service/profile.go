package service

import (
	"fmt"

	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"

	"gorm.io/gorm"
)

type ProfileService struct {
	db *gorm.DB
}

func NewProfileService(db *gorm.DB) *ProfileService {
	return &ProfileService{db: db}
}

func (s *ProfileService) GetProfile(userID string) (*dto.ProfileResponse, error) {
	var user model.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return &dto.ProfileResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (s *ProfileService) UpdateProfile(userID string, req dto.UpdateProfileRequest) (*dto.ProfileResponse, error) {
	updates := map[string]interface{}{}

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Email != "" {
		var existing model.User
		if err := s.db.Where("email = ? AND id != ?", req.Email, userID).First(&existing).Error; err == nil {
			return nil, fmt.Errorf("email already in use")
		}
		updates["email"] = req.Email
	}

	if len(updates) == 0 {
		return s.GetProfile(userID)
	}

	if err := s.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return s.GetProfile(userID)
}
