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
	ID          string    `json:"id"`
	ExternalID  string    `json:"external_id"`
	Title       string    `json:"title"`
	Company     string    `json:"company"`
	Location    string    `json:"location"`
	Description string    `json:"description"`
	Salary      string    `json:"salary,omitempty"`
	Remote      bool      `json:"remote"`
	Seniority   string    `json:"seniority,omitempty"`
	WorkType    string    `json:"work_type,omitempty"`
	SourceURL   string    `json:"source_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type JobSearchFilter struct {
	Query     string
	Location  string
	Company   string
	Seniority string
	WorkType  string
	Remote    *bool
	Limit     int
	Offset    int
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
	Job   *Job    `json:"job"`
	Score float64 `json:"score"`
}

type JobRepository struct {
	db *pgxpool.Pool
}

func NewJobRepository(db *pgxpool.Pool) *JobRepository {
	return &JobRepository{db: db}
}

func (r *JobRepository) Create(ctx context.Context, j *Job) (*Job, error) {
	if j == nil || strings.TrimSpace(j.ExternalID) == "" || strings.TrimSpace(j.Title) == "" {
		return nil, fmt.Errorf("%w: external_id and title are required", ErrInvalidJobInput)
	}

	const q = `
		INSERT INTO jobs (external_id, title, company, location, description, salary, remote, seniority, work_type, source_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, external_id, title, company, location, description, salary, remote, seniority, work_type, source_url, created_at, updated_at`

	row := r.db.QueryRow(ctx, q,
		j.ExternalID, j.Title, j.Company, j.Location, j.Description, j.Salary, j.Remote, j.Seniority, j.WorkType, j.SourceURL)

	out := &Job{}
	if err := row.Scan(&out.ID, &out.ExternalID, &out.Title, &out.Company, &out.Location,
		&out.Description, &out.Salary, &out.Remote, &out.Seniority, &out.WorkType, &out.SourceURL, &out.CreatedAt, &out.UpdatedAt); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrJobAlreadyExists
		}
		return nil, fmt.Errorf("repositories: create job: %w", err)
	}

	return out, nil
}

func (r *JobRepository) GetByID(ctx context.Context, id string) (*Job, error) {
	const q = `
		SELECT id, external_id, title, company, location, description, salary, remote, seniority, work_type, source_url, created_at, updated_at
		FROM jobs
		WHERE id = $1`
	return r.scanOne(ctx, q, id)
}

func (r *JobRepository) GetByIDs(ctx context.Context, ids []string) ([]*Job, error) {
	if len(ids) == 0 {
		return []*Job{}, nil
	}

	const q = `
		SELECT id, external_id, title, company, location, description, salary, remote, seniority, work_type, source_url, created_at, updated_at
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
			&j.Description, &j.Salary, &j.Remote, &j.Seniority, &j.WorkType, &j.SourceURL, &j.CreatedAt, &j.UpdatedAt); err != nil {
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

func (r *JobRepository) List(ctx context.Context, limit int) ([]*Job, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	const q = `
		SELECT id, external_id, title, company, location, description, salary, remote, seniority, work_type, source_url, created_at, updated_at
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
			&j.Description, &j.Salary, &j.Remote, &j.Seniority, &j.WorkType, &j.SourceURL, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repositories: scan job: %w", err)
		}
		jobs = append(jobs, j)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: iterate jobs: %w", err)
	}

	return jobs, nil
}

func (r *JobRepository) Search(ctx context.Context, filter JobSearchFilter) (*JobSearchResult, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if q := strings.TrimSpace(filter.Query); q != "" {
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR company ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, "%"+q+"%")
		argIdx++
	}
	if loc := strings.TrimSpace(filter.Location); loc != "" {
		conditions = append(conditions, fmt.Sprintf("location ILIKE $%d", argIdx))
		args = append(args, "%"+loc+"%")
		argIdx++
	}
	if comp := strings.TrimSpace(filter.Company); comp != "" {
		conditions = append(conditions, fmt.Sprintf("company ILIKE $%d", argIdx))
		args = append(args, "%"+comp+"%")
		argIdx++
	}
	if sen := strings.TrimSpace(filter.Seniority); sen != "" {
		conditions = append(conditions, fmt.Sprintf("seniority ILIKE $%d", argIdx))
		args = append(args, sen)
		argIdx++
	}
	if wt := strings.TrimSpace(filter.WorkType); wt != "" {
		conditions = append(conditions, fmt.Sprintf("work_type ILIKE $%d", argIdx))
		args = append(args, wt)
		argIdx++
	}
	if filter.Remote != nil {
		conditions = append(conditions, fmt.Sprintf("remote = $%d", argIdx))
		args = append(args, *filter.Remote)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// 1. Fetch real total matching count
	countQuery := "SELECT COUNT(*) FROM jobs" + whereClause
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("repositories: count search jobs: %w", err)
	}

	// Defaults & bounds for pagination
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	// 2. Query paginated results
	selectQuery := fmt.Sprintf(`
		SELECT id, external_id, title, company, location, description, salary, remote, seniority, work_type, source_url, created_at, updated_at
		FROM jobs%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.Query(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("repositories: search jobs query: %w", err)
	}
	defer rows.Close()

	jobs := make([]*Job, 0)
	for rows.Next() {
		j := &Job{}
		if err := rows.Scan(&j.ID, &j.ExternalID, &j.Title, &j.Company, &j.Location,
			&j.Description, &j.Salary, &j.Remote, &j.Seniority, &j.WorkType, &j.SourceURL, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repositories: scan searched job: %w", err)
		}
		jobs = append(jobs, j)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: iterate searched jobs: %w", err)
	}

	return &JobSearchResult{
		Jobs:  jobs,
		Total: total,
	}, nil
}

func (r *JobRepository) MatchJobs(ctx context.Context, filter SemanticMatchFilter) ([]*Job, error) {
	if len(filter.UserSkills) == 0 {
		return []*Job{}, nil
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// Fetch candidate jobs to be scored by services layer
	jobs, err := r.List(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("repositories: list candidate jobs for matching: %w", err)
	}

	return jobs, nil
}

func (r *JobRepository) scanOne(ctx context.Context, query string, args ...any) (*Job, error) {
	row := r.db.QueryRow(ctx, query, args...)

	j := &Job{}
	err := row.Scan(
		&j.ID, &j.ExternalID, &j.Title, &j.Company, &j.Location,
		&j.Description, &j.Salary, &j.Remote, &j.Seniority, &j.WorkType,
		&j.SourceURL, &j.CreatedAt, &j.UpdatedAt,
	)

	switch {
	case err == nil:
		return j, nil
	case errors.Is(err, pgx.ErrNoRows):
		return nil, ErrJobNotFound
	default:
		return nil, fmt.Errorf("repositories: query job: %w", err)
	}
}