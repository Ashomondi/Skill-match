package services

import (
	"context"

	"skill-match/backend/repositories"
)

type JobSource interface {
	FetchJobs(ctx context.Context) ([]SourceJob, error)
}

type SourceJob struct {
	ExternalID  string
	Title       string
	Company     string
	Location    string
	Description string
	Salary      string
	Remote      bool
	SourceURL   string
}

type JobService struct {
	repo   *repositories.JobRepository
	source JobSource
}

func NewJobService(
	repo *repositories.JobRepository,
	source JobSource,
) *JobService {
	return &JobService{
		repo:   repo,
		source: source,
	}
}

func (s *JobService) SearchJobs(ctx context.Context, filter repositories.JobSearchFilter) (*repositories.JobSearchResult, error) {
	return s.repo.Search(ctx, filter)
}

func (s *JobService) MatchJobs(ctx context.Context, filter repositories.SemanticMatchFilter) ([]*repositories.MatchScore, error) {
	return s.repo.MatchJobs(ctx, filter)
}