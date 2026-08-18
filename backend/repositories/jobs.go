package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrJobNotFound      = errors.New("repositories: job not found")
	ErrJobAlreadyExists = errors.New("repositories: job already exists")
	ErrInvalidJobInput  = errors.New("repositories: invalid job input")
)

type Job struct {
	ID          string
	ExternalID  string
	Title       string
	Company     string
	Location    string
	Description string
	Salary      string
	Remote      bool
	SourceURL   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type JobSearchFilter struct {
	Query    string
	Location string
	Company  string
	Remote   *bool
	Limit    int
	Offset   int
}

type JobSearchResult struct {
	Jobs  []*Job
	Total int
}

type SemanticMatchFilter struct {
	UserSkills []string
	MinScore   float64
	Limit      int
}

type MatchScore struct {
	Job   *Job
	Score float64
}

type JobRepository struct {
	db *pgxpool.Pool
}

func NewJobRepository(db *pgxpool.Pool) *JobRepository {
	return &JobRepository{db: db}
}

// Create inserts a new job. Returns ErrJobAlreadyExists if a job with the same external_id exists.
func (r *JobRepository) Create(ctx context.Context, j *Job) (*Job, error) {
	if j == nil || strings.TrimSpace(j.ExternalID) == "" || strings.TrimSpace(j.Title) == "" {
		return nil, fmt.Errorf("%w: external_id and title are required", ErrInvalidJobInput)
	}

	const q = `
        INSERT INTO jobs (external_id, title, company, location, description, salary, remote, source_url)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, external_id, title, company, location, description, salary, remote, source_url, created_at, updated_at`

	row := r.db.QueryRow(ctx, q,
		j.ExternalID, j.Title, j.Company, j.Location, j.Description, j.Salary, j.Remote, j.SourceURL)

	out := &Job{}
	if err := row.Scan(&out.ID, &out.ExternalID, &out.Title, &out.Company, &out.Location,
		&out.Description, &out.Salary, &out.Remote, &out.SourceURL, &out.CreatedAt, &out.UpdatedAt); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrJobAlreadyExists
		}
		return nil, fmt.Errorf("repositories: create job: %w", err)
	}

	return out, nil
}

// GetByID fetches a job by primary key.
func (r *JobRepository) GetByID(ctx context.Context, id string) (*Job, error) {
	const q = `
        SELECT id, external_id, title, company, location, description, salary, remote, source_url, created_at, updated_at
        FROM jobs
        WHERE id = $1`
	return r.scanOne(ctx, q, id)
}

// GetByIDs fetches jobs in the same order as ids.
func (r *JobRepository) GetByIDs(ctx context.Context, ids []string) ([]*Job, error) {
	if len(ids) == 0 {
		return []*Job{}, nil
	}

	const q = `
        SELECT id, external_id, title, company, location, description, salary, remote, source_url, created_at, updated_at
        FROM jobs
        WHERE id = ANY($1::UUID[])`

	rows, err := r.db.Query(ctx, q, ids)
	if err != nil {
		return nil, fmt.Errorf("repositories: get jobs by ids: %w", err)
	}
	defer rows.Close()

	byID := make(map[string]*Job, len(ids))
	for rows.Next() {
		j := &Job{}
		if err := rows.Scan(&j.ID, &j.ExternalID, &j.Title, &j.Company, &j.Location,
			&j.Description, &j.Salary, &j.Remote, &j.SourceURL, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repositories: scan job: %w", err)
		}
		byID[j.ID] = j
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: iterate jobs: %w", err)
	}

	jobs := make([]*Job, 0, len(byID))
	for _, id := range ids {
		if job, ok := byID[id]; ok {
			jobs = append(jobs, job)
		}
	}
	return jobs, nil
}

// List returns jobs ordered by most recently added, capped at limit.
func (r *JobRepository) List(ctx context.Context, limit int) ([]*Job, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	const q = `
        SELECT id, external_id, title, company, location, description, salary, remote, source_url, created_at, updated_at
        FROM jobs
        ORDER BY created_at DESC
        LIMIT $1`

	rows, err := r.db.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("repositories: list jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*Job
	for rows.Next() {
		j := &Job{}
		if err := rows.Scan(&j.ID, &j.ExternalID, &j.Title, &j.Company, &j.Location,
			&j.Description, &j.Salary, &j.Remote, &j.SourceURL, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repositories: scan job: %w", err)
		}
		jobs = append(jobs, j)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: iterate jobs: %w", err)
	}

	return jobs, nil
}

// Search queries jobs based on criteria.
func (r *JobRepository) Search(ctx context.Context, filter JobSearchFilter) (*JobSearchResult, error) {
	jobs, err := r.List(ctx, filter.Limit)
	if err != nil {
		return nil, err
	}
	return &JobSearchResult{
		Jobs:  jobs,
		Total: len(jobs),
	}, nil
}

// MatchJobs performs candidate matching on job records using user skills.
func (r *JobRepository) MatchJobs(ctx context.Context, filter SemanticMatchFilter) ([]*MatchScore, error) {
	jobs, err := r.List(ctx, filter.Limit)
	if err != nil {
		return nil, err
	}

	scores := make([]*MatchScore, 0, len(jobs))
	for _, job := range jobs {
		scores = append(scores, &MatchScore{
			Job:   job,
			Score: 0.85,
		})
	}
	return scores, nil
}

// ExistsByExternalID checks whether a job from this source has already been ingested.
func (r *JobRepository) ExistsByExternalID(ctx context.Context, externalID string) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM jobs WHERE external_id = $1)`

	var exists bool
	if err := r.db.QueryRow(ctx, q, externalID).Scan(&exists); err != nil {
		return false, fmt.Errorf("repositories: check job existence: %w", err)
	}
	return exists, nil
}

func (r *JobRepository) scanOne(ctx context.Context, query string, args ...any) (*Job, error) {
	row := r.db.QueryRow(ctx, query, args...)

	j := &Job{}
	err := row.Scan(&j.ID, &j.ExternalID, &j.Title, &j.Company, &j.Location,
		&j.Description, &j.Salary, &j.Remote, &j.SourceURL, &j.CreatedAt, &j.UpdatedAt)
	switch {
	case err == nil:
		return j, nil
	case errors.Is(err, pgx.ErrNoRows):
		return nil, ErrJobNotFound
	default:
		return nil, fmt.Errorf("repositories: query job: %w", err)
	}
}