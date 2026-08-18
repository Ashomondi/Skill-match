package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"skill-match/backend/models"
)

var (
	ErrApplicationNotFound     = errors.New("application not found")
	ErrApplicationInvalidInput = errors.New("invalid application input")
)

// ApplicationRepository defines the interface for repository access.
type ApplicationRepository interface {
	Create(ctx context.Context, userID, jobID string) (*models.Application, error)
	GetByID(ctx context.Context, userID, id string) (*models.Application, error)
	UpdateStatus(ctx context.Context, userID, id string, status models.ApplicationStatus) (*models.Application, error)
	History(ctx context.Context, userID, id string) ([]models.ApplicationStatusChange, error)
	ListByUserID(ctx context.Context, userID string) ([]*models.Application, error)
}

type ApplicationService struct {
	repo ApplicationRepository
}

func NewApplicationService(repo ApplicationRepository) *ApplicationService {
	return &ApplicationService{repo: repo}
}

func (s *ApplicationService) CreateApplication(ctx context.Context, userID, jobID string) (*models.Application, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(jobID) == "" {
		return nil, ErrApplicationInvalidInput
	}
	return s.repo.Create(ctx, userID, jobID)
}

func (s *ApplicationService) List(ctx context.Context, userID string) ([]*models.Application, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrApplicationInvalidInput
	}
	return s.repo.ListByUserID(ctx, userID)
}

func (s *ApplicationService) ListApplications(ctx context.Context, userID string) ([]*models.Application, error) {
	return s.List(ctx, userID)
}

func (s *ApplicationService) GetApplicationByID(ctx context.Context, userID, id string) (*models.Application, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(userID) == "" {
		return nil, ErrApplicationInvalidInput
	}
	return s.repo.GetByID(ctx, userID, id)
}

func (s *ApplicationService) UpdateApplicationStatus(ctx context.Context, userID, id string, status models.ApplicationStatus) (*models.Application, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(userID) == "" {
		return nil, ErrApplicationInvalidInput
	}

	if !isValidApplicationStatus(status) {
		return nil, fmt.Errorf("%w: invalid application status", ErrApplicationInvalidInput)
	}

	return s.repo.UpdateStatus(ctx, userID, id, status)
}

func isValidApplicationStatus(status models.ApplicationStatus) bool {
	// Evaluates the string representation to stay resilient across model variations
	switch strings.ToLower(string(status)) {
	case "saved", "applied", "interviewing", "rejected", "accepted", "offer":
		return true
	default:
		return false
	}
}