package services

import (
	"context"
	"errors"
	"fmt"

	"skill-match/backend/models"
	"skill-match/backend/repositories"
)

var (
	ErrApplicationNotFound     = errors.New("application not found")
	ErrApplicationAccessDenied = errors.New("application access denied")
	ErrInvalidApplication      = errors.New("invalid application")
)

// ApplicationRepository defines the persistence operations required by the
// application service. repositories.ApplicationRepository satisfies it.
type ApplicationRepository interface {
	Create(ctx context.Context, a *models.Application) (*models.Application, error)
	GetByID(ctx context.Context, id string) (*models.Application, error)
	ListByUserID(ctx context.Context, userID string, limit int) ([]*models.Application, error)
	UpdateStatus(ctx context.Context, id string, status models.ApplicationStatus) (*models.Application, error)
	Delete(ctx context.Context, id string) error
}

// ApplicationService coordinates application tracking. Every operation is
// scoped to (and ownership-checked against) the authenticated user.
type ApplicationService struct {
	repo ApplicationRepository
}

func NewApplicationService(repo ApplicationRepository) *ApplicationService {
	return &ApplicationService{repo: repo}
}

// Apply creates an application for a user, optionally tied to a job.
func (s *ApplicationService) Apply(ctx context.Context, userID string, jobID *string, status models.ApplicationStatus) (*models.Application, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user is required", ErrInvalidApplication)
	}
	if !status.Valid() {
		return nil, fmt.Errorf("%w: invalid application status", ErrInvalidApplication)
	}

	created, err := s.repo.Create(ctx, &models.Application{
		UserID: userID,
		JobID:  jobID,
		Status: status,
	})
	if err != nil {
		if errors.Is(err, repositories.ErrInvalidApplicationInput) {
			return nil, fmt.Errorf("%w: %v", ErrInvalidApplication, err)
		}
		return nil, err
	}
	return created, nil
}

// List returns the authenticated user's applications, most recently updated
// first.
func (s *ApplicationService) List(ctx context.Context, userID string, limit int) ([]*models.Application, error) {
	return s.repo.ListByUserID(ctx, userID, limit)
}

// UpdateStatus transitions an application the user owns.
func (s *ApplicationService) UpdateStatus(ctx context.Context, userID, id string, status models.ApplicationStatus) (*models.Application, error) {
	if !status.Valid() {
		return nil, fmt.Errorf("%w: invalid application status", ErrInvalidApplication)
	}

	a, err := s.getOwned(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	updated, err := s.repo.UpdateStatus(ctx, a.ID, status)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// Delete removes an application the user owns.
func (s *ApplicationService) Delete(ctx context.Context, userID, id string) error {
	a, err := s.getOwned(ctx, userID, id)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, a.ID)
}

// getOwned fetches an application and enforces that it belongs to userID.
func (s *ApplicationService) getOwned(ctx context.Context, userID, id string) (*models.Application, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrApplicationNotFound) {
			return nil, ErrApplicationNotFound
		}
		return nil, err
	}
	if a.UserID != userID {
		return nil, ErrApplicationAccessDenied
	}
	return a, nil
}
