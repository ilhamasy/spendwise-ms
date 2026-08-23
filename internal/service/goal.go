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

var goalValidate = validator.New()

type GoalService struct {
	repo repository.IGoalRepository
}

func NewGoalService(db *gorm.DB) *GoalService {
	return &GoalService{repo: repository.NewGoalRepository(db)}
}

func NewGoalServiceWithRepo(r repository.IGoalRepository) *GoalService {
	return &GoalService{repo: r}
}

func (s *GoalService) ListGoals(userID, status string) ([]dto.GoalResponse, error) {
	goals, err := s.repo.FindAll(userID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve goals: %w", err)
	}

	var resp []dto.GoalResponse
	for _, goal := range goals {
		resp = append(resp, goalToResponse(goal))
	}
	return resp, nil
}

func (s *GoalService) CreateGoal(userID string, req dto.CreateGoalRequest) (*dto.GoalResponse, error) {
	if err := goalValidate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	if req.TargetDate != "" {
		today := time.Now().Format("2006-01-02")
		if req.TargetDate < today {
			return nil, fmt.Errorf("target date cannot be in the past")
		}
	}

	goal := model.SavingGoal{
		UserID:       userID,
		Name:         dto.SanitizeString(req.Name),
		TargetAmount: req.TargetAmount,
		CurrentSaved: req.CurrentSaved,
		TargetDate:   req.TargetDate,
		Status:       "active",
	}

	if err := s.repo.Create(&goal); err != nil {
		return nil, fmt.Errorf("failed to create goal: %w", err)
	}

	resp := goalToResponse(goal)
	return &resp, nil
}

func (s *GoalService) GetGoal(userID, id string) (*dto.GoalResponse, error) {
	goal, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, err
	}
	resp := goalToResponse(*goal)
	return &resp, nil
}

func (s *GoalService) UpdateGoal(userID, id string, req dto.UpdateGoalRequest) (*dto.GoalResponse, error) {
	if err := goalValidate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	goal, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		goal.Name = dto.SanitizeString(req.Name)
	}
	if req.TargetAmount > 0 {
		goal.TargetAmount = req.TargetAmount
	}
	goal.CurrentSaved = req.CurrentSaved
	if req.TargetDate != "" {
		goal.TargetDate = req.TargetDate
	}

	if err := s.repo.Update(goal); err != nil {
		return nil, fmt.Errorf("failed to update goal: %w", err)
	}

	updated, err := s.repo.FindByID(id, userID)
	if err != nil || updated == nil {
		updated = goal
	}
	resp := goalToResponse(*updated)
	return &resp, nil
}

func (s *GoalService) ArchiveGoal(userID, id string) (*dto.GoalResponse, error) {
	existing, err := s.repo.FindByID(id, userID)
	if err != nil || existing == nil {
		return nil, errors.New("goal not found")
	}

	if err := s.repo.UpdateStatus(id, userID, "archived"); err != nil {
		return nil, fmt.Errorf("failed to archive goal: %w", err)
	}

	goal, err := s.repo.FindByID(id, userID)
	if err != nil || goal == nil {
		goal = existing
		goal.Status = "archived"
	}
	resp := goalToResponse(*goal)
	return &resp, nil
}

func (s *GoalService) UnarchiveGoal(userID, id string) (*dto.GoalResponse, error) {
	existing, err := s.repo.FindByID(id, userID)
	if err != nil || existing == nil {
		return nil, errors.New("goal not found")
	}

	if err := s.repo.UpdateStatus(id, userID, "active"); err != nil {
		return nil, fmt.Errorf("failed to unarchive goal: %w", err)
	}

	goal, err := s.repo.FindByID(id, userID)
	if err != nil || goal == nil {
		goal = existing
		goal.Status = "active"
	}
	resp := goalToResponse(*goal)
	return &resp, nil
}

func (s *GoalService) DeleteGoal(userID, id string) error {
	return s.repo.Delete(id, userID)
}

func (s *GoalService) AddContribution(userID, goalID string, req dto.CreateContributionRequest) (*dto.ContributionResponse, *dto.GoalResponse, error) {
	if err := goalValidate.Struct(req); err != nil {
		return nil, nil, fmt.Errorf("validation error: %w", err)
	}

	goal, err := s.repo.FindByID(goalID, userID)
	if err != nil {
		return nil, nil, err
	}

	date := req.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	contribution := model.GoalContribution{
		GoalID: goalID,
		Amount: req.Amount,
		Note:   req.Note,
		Date:   date,
	}

	if err := s.repo.CreateContribution(&contribution); err != nil {
		return nil, nil, fmt.Errorf("failed to create contribution: %w", err)
	}

	goal.CurrentSaved += req.Amount
	if err := s.repo.Update(goal); err != nil {
		return nil, nil, fmt.Errorf("failed to update goal progress: %w", err)
	}

	updatedGoal, _ := s.repo.FindByID(goalID, userID)
	contribResp := contributionToResponse(contribution)
	goalResp := goalToResponse(*updatedGoal)
	return &contribResp, &goalResp, nil
}

func (s *GoalService) GetContributions(userID, goalID string, page, limit int) (*dto.ContributionListResponse, error) {
	_, err := s.repo.FindByID(goalID, userID)
	if err != nil {
		return nil, err
	}

	contributions, total, err := s.repo.FindContributionsByGoalID(goalID, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve contributions: %w", err)
	}

	var data []dto.ContributionResponse
	for _, c := range contributions {
		data = append(data, contributionToResponse(c))
	}

	return &dto.ContributionListResponse{
		Data: data,
		Meta: dto.Meta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: repository.CalculateTotalPages(total, limit),
		},
	}, nil
}

func goalToResponse(goal model.SavingGoal) dto.GoalResponse {
	return dto.GoalResponse{
		ID:           goal.ID,
		Name:         goal.Name,
		TargetAmount: goal.TargetAmount,
		CurrentSaved: goal.CurrentSaved,
		TargetDate:   goal.TargetDate,
		Status:       goal.Status,
		Progress:     repository.CalculateGoalProgress(goal.CurrentSaved, goal.TargetAmount),
		CreatedAt:    goal.CreatedAt,
		UpdatedAt:    goal.UpdatedAt,
	}
}

func contributionToResponse(c model.GoalContribution) dto.ContributionResponse {
	return dto.ContributionResponse{
		ID:        c.ID,
		GoalID:    c.GoalID,
		Amount:    c.Amount,
		Note:      c.Note,
		Date:      c.Date,
		CreatedAt: c.CreatedAt,
	}
}
