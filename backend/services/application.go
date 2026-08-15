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
	ErrApplicationInvalidInput      = errors.New("services: invalid application input")
	ErrApplicationInvalidStatus     = errors.New("services: invalid application status")
	ErrApplicationInvalidTransition = errors.New("services: invalid application status transition")
	ErrApplicationNotFound          = errors.New("services: application not found")
	ErrApplicationDuplicate         = errors.New("services: application already exists")
)

type ApplicationRepository interface {
	Create(context.Context, string, string) (*models.Application, error)
	GetByID(context.Context, string, string) (*models.Application, error)
	UpdateStatus(context.Context, string, string, models.ApplicationStatus) (*models.Application, error)
	History(context.Context, string, string) ([]models.ApplicationStatusChange, error)
}
type ApplicationService struct{ repo ApplicationRepository }

func NewApplicationService(repo ApplicationRepository) *ApplicationService {
	return &ApplicationService{repo: repo}
}
func (s *ApplicationService) Create(ctx context.Context, userID, jobID string) (*models.Application, error) {
	if strings.TrimSpace(userID) == "" || uuid.Validate(jobID) != nil {
		return nil, ErrApplicationInvalidInput
	}
	application, err := s.repo.Create(ctx, userID, jobID)
	if errors.Is(err, repositories.ErrApplicationDuplicate) {
		return nil, ErrApplicationDuplicate
	}
	if errors.Is(err, repositories.ErrJobNotFound) {
		return nil, ErrApplicationNotFound
	}
	return application, err
}
func (s *ApplicationService) Get(ctx context.Context, userID, id string) (*models.Application, error) {
	if uuid.Validate(id) != nil || strings.TrimSpace(userID) == "" {
		return nil, ErrApplicationNotFound
	}
	a, e := s.repo.GetByID(ctx, userID, id)
	if errors.Is(e, repositories.ErrApplicationNotFound) {
		return nil, ErrApplicationNotFound
	}
	return a, e
}
func (s *ApplicationService) UpdateStatus(ctx context.Context, userID, id string, status models.ApplicationStatus) (*models.Application, error) {
	if !status.Valid() {
		return nil, ErrApplicationInvalidStatus
	}
	a, e := s.Get(ctx, userID, id)
	if e != nil {
		return nil, e
	}
	if !allowedTransition(a.Status, status) {
		return nil, ErrApplicationInvalidTransition
	}
	updated, err := s.repo.UpdateStatus(ctx, userID, id, status)
	if errors.Is(err, repositories.ErrApplicationNotFound) {
		return nil, ErrApplicationNotFound
	}
	return updated, err
}
func (s *ApplicationService) GetHistory(ctx context.Context, userID, id string) ([]models.ApplicationStatusChange, error) {
	if strings.TrimSpace(userID) == "" || uuid.Validate(id) != nil {
		return nil, ErrApplicationNotFound
	}
	history, err := s.repo.History(ctx, userID, id)
	if errors.Is(err, repositories.ErrApplicationNotFound) {
		return nil, ErrApplicationNotFound
	}
	return history, err
}
func allowedTransition(from, to models.ApplicationStatus) bool {
	switch from {
	case models.ApplicationSaved:
		return to == models.ApplicationApplied || to == models.ApplicationWithdrawn
	case models.ApplicationApplied:
		return to == models.ApplicationScreening || to == models.ApplicationInterview || to == models.ApplicationRejected || to == models.ApplicationWithdrawn
	case models.ApplicationScreening:
		return to == models.ApplicationInterview || to == models.ApplicationOffer || to == models.ApplicationRejected || to == models.ApplicationWithdrawn
	case models.ApplicationInterview:
		return to == models.ApplicationOffer || to == models.ApplicationRejected || to == models.ApplicationWithdrawn
	}
	return false
}
