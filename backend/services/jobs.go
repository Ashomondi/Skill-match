package services

import (
	"context"
	"fmt"
	"log"
	"time"

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
	Seniority   string
	WorkType    string
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
		_, err = s.repo.Create(ctx, &repositories.Job{
			ExternalID:  source.ExternalID,
			Title:       source.Title,
			Company:     source.Company,
			Location:    source.Location,
			Description: source.Description,
			Salary:      source.Salary,
			Remote:      source.Remote,
			Seniority:   source.Seniority,
			WorkType:    source.WorkType,
			SourceURL:   source.SourceURL,
		})
		if err != nil {
			return ingested, skipped, err
		}
		ingested++
	}
	return ingested, skipped, nil
}

func (s *JobService) GetJob(ctx context.Context, id string) (*repositories.Job, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *JobService) SearchJobs(ctx context.Context, filter repositories.JobSearchFilter) (*repositories.JobSearchResult, error) {
	return s.repo.Search(ctx, filter)
}

func (s *JobService) MatchJobs(ctx context.Context, filter repositories.SemanticMatchFilter) ([]*repositories.MatchScore, error) {
	if len(filter.UserSkills) == 0 {
		return []*repositories.MatchScore{}, nil
	}

	candidateJobs, err := s.repo.MatchJobs(ctx, filter)
	if err != nil {
		return nil, err
	}

	matchingSvc := NewMatchingService()
	return matchingSvc.RankJobs(filter.UserSkills, candidateJobs, filter.MinScore), nil
}

func (s *JobService) IngestJobsWithRetry(ctx context.Context, maxRetries int) (int, int, error) {
	var (
		ingested int
		skipped  int
		err      error
	)

	backoff := 1 * time.Second
	for attempt := 1; attempt <= maxRetries; attempt++ {
		ingested, skipped, err = s.IngestJobs(ctx)
		if err == nil {
			log.Printf("job ingestion succeeded (attempt %d): %d ingested, %d skipped", attempt, ingested, skipped)
			return ingested, skipped, nil
		}

		log.Printf("WARNING: job ingestion attempt %d/%d failed: %v", attempt, maxRetries, err)

		if attempt < maxRetries {
			select {
			case <-ctx.Done():
				return 0, 0, ctx.Err()
			case <-time.After(backoff):
				backoff *= 2
			}
		}
	}

	return ingested, skipped, fmt.Errorf("job_service: ingestion failed after %d retries: %w", maxRetries, err)
}
