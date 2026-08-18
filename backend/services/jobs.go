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

func (s *JobService) IngestJobs(ctx context.Context) (int, int, error) {
	jobs, err := s.source.FetchJobs(ctx)
	if err != nil {
		return 0, 0, err
	}
	ingested, skipped := 0, 0
	for _, source := range jobs {
		exists, err := s.repo.ExistsByExternalID(ctx, source.ExternalID)
		if err != nil {
			return ingested, skipped, err
		}
		if exists {
			skipped++
			continue
		}
		_, err = s.repo.Create(ctx, &repositories.Job{ExternalID: source.ExternalID, Title: source.Title, Company: source.Company, Location: source.Location, Description: source.Description, Salary: source.Salary, Remote: source.Remote, SourceURL: source.SourceURL})
		if err != nil {
			return ingested, skipped, err
		}
		ingested++
	}
	return ingested, skipped, nil
}

func (s *JobService) SearchJobs(ctx context.Context, filter repositories.JobSearchFilter) (*repositories.JobSearchResult, error) {
	return s.repo.Search(ctx, filter)
}

func (s *JobService) MatchJobs(ctx context.Context, filter repositories.SemanticMatchFilter) ([]*repositories.MatchScore, error) {
	return s.repo.MatchJobs(ctx, filter)
}
