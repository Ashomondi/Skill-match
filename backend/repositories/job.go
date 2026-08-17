package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Job struct {
	ID          string    `json:"id"`
	ExternalID  string    `json:"external_id"`
	Title       string    `json:"title"`
	Company     string    `json:"company"`
	Location    string    `json:"location"`
	Description string    `json:"description"`
	Salary      string    `json:"salary"`
	Remote      bool      `json:"remote"`
	SourceURL   string    `json:"source_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
	Jobs  []*Job `json:"jobs"`
	Total int    `json:"total"`
}

type JobRepository struct {
	db *pgxpool.Pool
}

type MatchScore struct {
	Job        *Job    `json:"job"`
	Score      float64 `json:"score"`
	MatchedOn []string `json:"matched_skills,omitempty"`
}

type SemanticMatchFilter struct {
	UserSkills []string `json:"user_skills"`
	MinScore   float64  `json:"min_score"`
	Limit      int      `json:"limit"`
}

func NewJobRepository(db *pgxpool.Pool) *JobRepository {
	return &JobRepository{db: db}
}

func (r *JobRepository) Search(ctx context.Context, filter JobSearchFilter) (*JobSearchResult, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	whereClause := []string{"1=1"}
	args := []any{}
	argIdx := 1

	if filter.Query != "" {
		whereClause = append(whereClause, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+filter.Query+"%")
		argIdx++
	}

	if filter.Location != "" {
		whereClause = append(whereClause, fmt.Sprintf("location ILIKE $%d", argIdx))
		args = append(args, "%"+filter.Location+"%")
		argIdx++
	}

	if filter.Company != "" {
		whereClause = append(whereClause, fmt.Sprintf("company ILIKE $%d", argIdx))
		args = append(args, "%"+filter.Company+"%")
		argIdx++
	}

	if filter.Remote != nil {
		whereClause = append(whereClause, fmt.Sprintf("remote = $%d", argIdx))
		args = append(args, *filter.Remote)
		argIdx++
	}

	whereStr := strings.Join(whereClause, " AND ")

	// Count total records matching filter
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM jobs WHERE %s", whereStr)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("repositories: search jobs count: %w", err)
	}

	// Fetch paginated results
	selectQuery := fmt.Sprintf(`
		SELECT id, external_id, title, company, location, description, salary, remote, source_url, created_at, updated_at
		FROM jobs
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereStr, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("repositories: search jobs query: %w", err)
	}
	defer rows.Close()

	jobs := []*Job{}
	for rows.Next() {
		j := &Job{}
		if err := rows.Scan(
			&j.ID, &j.ExternalID, &j.Title, &j.Company, &j.Location,
			&j.Description, &j.Salary, &j.Remote, &j.SourceURL,
			&j.CreatedAt, &j.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repositories: scan search job row: %w", err)
		}
		jobs = append(jobs, j)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: search jobs rows iteration: %w", err)
	}

	return &JobSearchResult{
		Jobs:  jobs,
		Total: total,
	}, nil
}

func (r *JobRepository) MatchJobs(ctx context.Context, filter SemanticMatchFilter) ([]*MatchScore, error) {
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	if len(filter.UserSkills) == 0 {
		return []*MatchScore{}, nil
	}

	// Prepare terms for ILIKE and array keyword matching
	var matches []*MatchScore
	query := `
		SELECT id, external_id, title, company, location, description, salary, remote, source_url, created_at, updated_at
		FROM jobs
		WHERE 1=1
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repositories: match jobs query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		j := &Job{}
		if err := rows.Scan(
			&j.ID, &j.ExternalID, &j.Title, &j.Company, &j.Location,
			&j.Description, &j.Salary, &j.Remote, &j.SourceURL,
			&j.CreatedAt, &j.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repositories: scan matched job: %w", err)
		}

		// Calculate term match ratio over job description and title
		matched := []string{}
		textToSearch := strings.ToLower(j.Title + " " + j.Description)
		
		for _, skill := range filter.UserSkills {
			skillLower := strings.ToLower(strings.TrimSpace(skill))
			if skillLower != "" && strings.Contains(textToSearch, skillLower) {
				matched = append(matched, skill)
			}
		}
		score := 0.0
		if len(filter.UserSkills) > 0 {
			score = float64(len(matched)) / float64(len(filter.UserSkills))
		}

		if score >= filter.MinScore {
			matches = append(matches, &MatchScore{
				Job:        j,
				Score:      score,
				MatchedOn: matched,
			})
		}
	}

	return matches, nil
}