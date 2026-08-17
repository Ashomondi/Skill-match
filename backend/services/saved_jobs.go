package services

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"skill-match/backend/models"
	"skill-match/backend/repositories"
	"strings"
)

var (
	ErrSavedJobInvalidInput = errors.New("services: invalid saved job input")
	ErrSavedJobNotFound     = errors.New("services: saved job not found")
	ErrSavedJobDuplicate    = errors.New("services: job already saved")
)

type SavedJobRepository interface {
	Create(context.Context, string, string) (*models.SavedJob, error)
	ListByUserID(context.Context, string) ([]*models.SavedJob, error)
	Delete(context.Context, string, string) error
}
type SavedJobService struct{ repo SavedJobRepository }

func NewSavedJobService(repo SavedJobRepository) *SavedJobService {
	return &SavedJobService{repo: repo}
}
func validIDs(userID, jobID string) error {
	if strings.TrimSpace(userID) == "" || uuid.Validate(jobID) != nil {
		return ErrSavedJobInvalidInput
	}
	return nil
}
func (s *SavedJobService) Save(ctx context.Context, userID, jobID string) (*models.SavedJob, error) {
	if err := validIDs(userID, jobID); err != nil {
		return nil, err
	}
	v, e := s.repo.Create(ctx, userID, jobID)
	if errors.Is(e, repositories.ErrSavedJobDuplicate) {
		return nil, ErrSavedJobDuplicate
	}
	if errors.Is(e, repositories.ErrJobNotFound) {
		return nil, ErrSavedJobNotFound
	}
	return v, e
}
func (s *SavedJobService) List(ctx context.Context, userID string) ([]*models.SavedJob, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrSavedJobInvalidInput
	}
	return s.repo.ListByUserID(ctx, userID)
}
func (s *SavedJobService) Remove(ctx context.Context, userID, jobID string) error {
	if err := validIDs(userID, jobID); err != nil {
		return err
	}
	err := s.repo.Delete(ctx, userID, jobID)
	if errors.Is(err, repositories.ErrSavedJobNotFound) {
		return ErrSavedJobNotFound
	}
	return err
}
