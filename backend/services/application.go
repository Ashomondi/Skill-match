package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"skill-match/backend/models"
	"skill-match/backend/repositories"
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


func (s *ApplicationService) CreateApplication(ctx context.Context, userID, jobID string, status repositories.ApplicationStatus, notes string) (*models.Application, error) {
	if jobID == "" {
		return nil, fmt.Errorf("validation: job_id is required")

	}
	return s.repo.Create(ctx, userID, jobID)
}

func (s *ApplicationService) ListApplications(ctx context.Context, userID string) ([]*models.Application, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *ApplicationService) GetApplicationByID(ctx context.Context, id, userID string) (*models.Application, error) {
	return s.repo.GetByID(ctx, id, userID)
}

func (s *ApplicationService) UpdateApplicationStatus(ctx context.Context, id, userID string, status repositories.ApplicationStatus, notes string) (*models.Application, error) {
	if !status.Valid() {
		return nil, fmt.Errorf("validation: invalid application status")
	}
	return s.repo.GetByID(ctx, userID, id)
}

// func (s *ApplicationService) UpdateApplicationStatus(ctx context.Context, userID, id string, status models.ApplicationStatus) (*models.Application, error) {
// 	if strings.TrimSpace(id) == "" || strings.TrimSpace(userID) == "" {
// 		return nil, ErrApplicationInvalidInput
// 	}

// 	if !isValidApplicationStatus(status) {
// 		return nil, fmt.Errorf("%w: invalid application status", ErrApplicationInvalidInput)
// 	}

// 	return s.repo.UpdateStatus(ctx, userID, id, status)
// }

func isValidApplicationStatus(status models.ApplicationStatus) bool {
	// Evaluates the string representation to stay resilient across model variations
	switch strings.ToLower(string(status)) {
	case "saved", "applied", "interviewing", "rejected", "accepted", "offer":
		return true
	default:
		return false
	}
}

func (s *ApplicationService) DeleteApplication(ctx context.Context, id, userID string) error {
	return s.repo.Delete(ctx, id, userID)
}

