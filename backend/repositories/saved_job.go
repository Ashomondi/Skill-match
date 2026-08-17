package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

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

func (r *SavedJobRepository) SaveJob(ctx context.Context, userID, jobID string) (*SavedJob, error) {
	query := `
		INSERT INTO saved_jobs (user_id, job_id)
		VALUES ($1, $2)
		RETURNING id, user_id, job_id, created_at
	`
	sj := &SavedJob{}
	err := r.db.QueryRow(ctx, query, userID, jobID).Scan(&sj.ID, &sj.UserID, &sj.JobID, &sj.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("repositories: save job: %w", err)
	}
	return sj, nil
}

func (r *SavedJobRepository) GetSavedJobsByUserID(ctx context.Context, userID string) ([]*SavedJob, error) {
	query := `
		SELECT sj.id, sj.user_id, sj.job_id, sj.created_at,
		       j.id, j.external_id, j.title, j.company, j.location, j.description, j.salary, j.remote, j.source_url, j.created_at, j.updated_at
		FROM saved_jobs sj
		JOIN jobs j ON sj.job_id = j.id
		WHERE sj.user_id = $1
		ORDER BY sj.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repositories: get saved jobs: %w", err)
	}
	defer rows.Close()

	var results []*SavedJob
	for rows.Next() {
		sj := &SavedJob{}
		j := &Job{}
		if err := rows.Scan(
			&sj.ID, &sj.UserID, &sj.JobID, &sj.CreatedAt,
			&j.ID, &j.ExternalID, &j.Title, &j.Company, &j.Location,
			&j.Description, &j.Salary, &j.Remote, &j.SourceURL,
			&j.CreatedAt, &j.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repositories: scan saved job: %w", err)
		}
		sj.Job = j
		results = append(results, sj)
	}
	return results, nil
}

func (r *SavedJobRepository) DeleteSavedJob(ctx context.Context, id, userID string) error {
	query := `
		DELETE FROM saved_jobs
		WHERE id = $1 AND user_id = $2
	`
	cmd, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("repositories: delete saved job: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("not_found")
	}
	return nil
}