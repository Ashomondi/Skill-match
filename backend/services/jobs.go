package services

import (
	"context"
	"fmt"
	"strings"

	"skill-match/backend/repositories"
)

// JobSource is implemented by whatever supplies job listings — a real
// external API, or (for the hackathon) a static seed set. Swapping the
// source later requires no changes here.
type JobSource interface {
	FetchJobs(ctx context.Context) ([]SourceJob, error)
}

// SourceJob is the shape a JobSource returns, before persistence.
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

func NewJobService(repo *repositories.JobRepository, source JobSource) *JobService {
	return &JobService{repo: repo, source: source}
}

// IngestJobs pulls jobs from the configured source, validates each one,
// and persists new jobs to CockroachDB. Duplicates (by external_id) are
// skipped, not treated as errors.
func (s *JobService) IngestJobs(ctx context.Context) (ingested int, skipped int, err error) {
	jobs, err := s.source.FetchJobs(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("fetching jobs from source: %w", err)
	}

	for _, j := range jobs {
		if err := validateSourceJob(j); err != nil {
			skipped++
			continue
		}

		exists, err := s.repo.ExistsByExternalID(ctx, j.ExternalID)
		if err != nil {
			return ingested, skipped, fmt.Errorf("checking job existence: %w", err)
		}
		if exists {
			skipped++
			continue
		}

		_, err = s.repo.Create(ctx, &repositories.Job{
			ExternalID:  j.ExternalID,
			Title:       j.Title,
			Company:     j.Company,
			Location:    j.Location,
			Description: j.Description,
			Salary:      j.Salary,
			Remote:      j.Remote,
			SourceURL:   j.SourceURL,
		})
		if err != nil {
			return ingested, skipped, fmt.Errorf("persisting job %s: %w", j.ExternalID, err)
		}
		ingested++
	}

	return ingested, skipped, nil
}

// ListJobs returns stored jobs available to the application.
func (s *JobService) ListJobs(ctx context.Context, limit int) ([]*repositories.Job, error) {
	return s.repo.List(ctx, limit)
}

func validateSourceJob(j SourceJob) error {
	if strings.TrimSpace(j.ExternalID) == "" {
		return fmt.Errorf("external_id is required")
	}
	if strings.TrimSpace(j.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if strings.TrimSpace(j.Company) == "" {
		return fmt.Errorf("company is required")
	}
	return nil
}
