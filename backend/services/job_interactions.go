package services

import (
	"context"
	"errors"
	"fmt"

	"skill-match/backend/models"
	"skill-match/backend/repositories"
)

var (
	ErrJobInteractionNotFound     = errors.New("job interaction not found")
	ErrJobInteractionAccessDenied = errors.New("job interaction access denied")
	ErrInvalidJobInteraction      = errors.New("invalid job interaction")
)

// JobInteractionRepository defines the persistence operations required by the
// interaction service. repositories.JobInteractionRepository satisfies it.
type JobInteractionRepository interface {
	Create(ctx context.Context, in *models.JobInteraction) (*models.JobInteraction, error)
	GetByID(ctx context.Context, id string) (*models.JobInteraction, error)
	ListByUserID(ctx context.Context, userID string, limit int) ([]*models.JobInteraction, error)
	Delete(ctx context.Context, id string) error
}

// JobInteractionService records and retrieves user-job interactions, the
// behavioral signal that powers personalized recommendations. Ownership is
// enforced on every operation.
type JobInteractionService struct {
	repo JobInteractionRepository
}

func NewJobInteractionService(repo JobInteractionRepository) *JobInteractionService {
	return &JobInteractionService{repo: repo}
}

// Record creates a job interaction for a user. userID and jobID must be
// non-empty and the interaction type must be one of view|save|apply|dismiss|search.
func (s *JobInteractionService) Record(ctx context.Context, userID, jobID string, interactionType models.InteractionType) (*models.JobInteraction, error) {
	if userID == "" || jobID == "" {
		return nil, fmt.Errorf("%w: user_id and job_id are required", ErrInvalidJobInteraction)
	}
	if !interactionType.Valid() {
		return nil, fmt.Errorf("%w: interaction_type must be one of view|save|apply|dismiss|search", ErrInvalidJobInteraction)
	}

	created, err := s.repo.Create(ctx, &models.JobInteraction{
		UserID: userID,
		JobID:  jobID,
		Type:   interactionType,
	})
	if err != nil {
		if errors.Is(err, repositories.ErrInvalidJobInteractionInput) {
			return nil, fmt.Errorf("%w: %v", ErrInvalidJobInteraction, err)
		}
		return nil, err
	}
	return created, nil
}

// List returns the authenticated user's own interactions, most recent first.
func (s *JobInteractionService) List(ctx context.Context, userID string, limit int) ([]*models.JobInteraction, error) {
	return s.repo.ListByUserID(ctx, userID, limit)
}

// Delete removes an interaction, but only if it belongs to userID.
func (s *JobInteractionService) Delete(ctx context.Context, userID, id string) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrJobInteractionNotFound) {
			return ErrJobInteractionNotFound
		}
		return err
	}
	if existing.UserID != userID {
		return ErrJobInteractionAccessDenied
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}
