package services

import (
	"context"

	"skill-match/backend/repositories"
)

type SavedJobService struct {
	repo *repositories.SavedJobRepository
}

func NewSavedJobService(repo *repositories.SavedJobRepository) *SavedJobService {
	return &SavedJobService{repo: repo}
}

func (s *SavedJobService) SaveJob(ctx context.Context, userID, jobID string) (*repositories.SavedJob, error) {
	return s.repo.SaveJob(ctx, userID, jobID)
}

func (s *SavedJobService) GetSavedJobs(ctx context.Context, userID string) ([]*repositories.SavedJob, error) {
	return s.repo.GetSavedJobsByUserID(ctx, userID)
}

func (s *SavedJobService) DeleteSavedJob(ctx context.Context, id, userID string) error {
	return s.repo.DeleteSavedJob(ctx, id, userID)
}