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

// SavedJob represents a saved job. The saved_jobs table has no id column
// (the unique key is user_id + job_id), so ID mirrors JobID to keep the
// API shape (delete by id == delete by job id) stable for clients.
type SavedJob struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	JobID     string    `json:"job_id"`
	CreatedAt time.Time `json:"created_at"`
	Job       *Job      `json:"job,omitempty"`
}

type SavedJobRepository struct {
	db *pgxpool.Pool
}

func NewSavedJobRepository(db *pgxpool.Pool) *SavedJobRepository {
	return &SavedJobRepository{db: db}
}

// SaveJob saves a job for a user. Saving an already-saved job is a no-op
// that returns the existing row.
func (r *SavedJobRepository) SaveJob(ctx context.Context, userID, jobID string) (*SavedJob, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(jobID) == "" {
		return nil, fmt.Errorf("repositories: user_id and job_id are required")
	}

	sj := &SavedJob{}
	const insert = `INSERT INTO saved_jobs (user_id, job_id) VALUES ($1, $2) ON CONFLICT (user_id, job_id) DO NOTHING RETURNING user_id, job_id, saved_at`
	err := r.db.QueryRow(ctx, insert, userID, jobID).Scan(&sj.UserID, &sj.JobID, &sj.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		// Already saved — fetch the existing row.
		const fetch = `SELECT user_id, job_id, saved_at FROM saved_jobs WHERE user_id = $1 AND job_id = $2`
		if err := r.db.QueryRow(ctx, fetch, userID, jobID).Scan(&sj.UserID, &sj.JobID, &sj.CreatedAt); err != nil {
			return nil, fmt.Errorf("repositories: fetch saved job: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("repositories: save job: %w", err)
	}
	sj.ID = sj.JobID
	return sj, nil
}

// GetSavedJobsByUserID returns a user's saved jobs, most recently saved
// first, enriched with job details.
func (r *SavedJobRepository) GetSavedJobsByUserID(ctx context.Context, userID string) ([]*SavedJob, error) {
	const q = `
		SELECT sj.user_id, sj.job_id, sj.saved_at,
		       j.id, j.external_id, j.title, j.company, j.location, j.description,
		       j.salary, j.remote, j.source_url, j.created_at, j.updated_at
		FROM saved_jobs sj
		JOIN jobs j ON sj.job_id = j.id
		WHERE sj.user_id = $1
		ORDER BY sj.saved_at DESC`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("repositories: get saved jobs: %w", err)
	}
	defer rows.Close()

	var results []*SavedJob
	for rows.Next() {
		sj := &SavedJob{}
		j := &Job{}
		if err := rows.Scan(
			&sj.UserID, &sj.JobID, &sj.CreatedAt,
			&j.ID, &j.ExternalID, &j.Title, &j.Company, &j.Location, &j.Description,
			&j.Salary, &j.Remote, &j.SourceURL, &j.CreatedAt, &j.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repositories: scan saved job: %w", err)
		}
		sj.ID = sj.JobID
		sj.Job = j
		results = append(results, sj)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: iterate saved jobs: %w", err)
	}
	return results, nil
}

// DeleteSavedJob removes a saved job by job id (the id clients use).
func (r *SavedJobRepository) DeleteSavedJob(ctx context.Context, id, userID string) error {
	cmd, err := r.db.Exec(ctx, `DELETE FROM saved_jobs WHERE job_id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("repositories: delete saved job: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("not_found")
	}
	return nil
}
