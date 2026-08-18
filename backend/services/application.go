package services

import (
	"context"
	"fmt"

	"skill-match/backend/models"
	"skill-match/backend/repositories"
)

type ApplicationService struct {
	repo *repositories.ApplicationRepository
}

func NewApplicationService(repo *repositories.ApplicationRepository) *ApplicationService {
	return &ApplicationService{repo: repo}
}

func (s *ApplicationService) CreateApplication(ctx context.Context, userID, jobID string, status repositories.ApplicationStatus, notes string) (*models.Application, error) {
	if jobID == "" {
		return nil, fmt.Errorf("validation: job_id is required")
	}
	return s.repo.Create(ctx, userID, jobID, status, notes)
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
	return s.repo.UpdateStatus(ctx, id, userID, status, notes)
}

func (s *ApplicationService) DeleteApplication(ctx context.Context, id, userID string) error {
	return s.repo.Delete(ctx, id, userID)
}
