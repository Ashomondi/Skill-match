package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"skill-match/backend/models"
)

var (
	ErrSavedJobNotFound     = errors.New("repositories: saved job not found")
	ErrSavedJobDuplicate    = errors.New("repositories: job already saved")
	ErrInvalidSavedJobInput = errors.New("repositories: invalid saved job input")
)

type SavedJobRepository struct{ db *pgxpool.Pool }

func NewSavedJobRepository(db *pgxpool.Pool) *SavedJobRepository { return &SavedJobRepository{db: db} }

func (r *SavedJobRepository) Create(ctx context.Context, userID, jobID string) (*models.SavedJob, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(jobID) == "" {
		return nil, ErrInvalidSavedJobInput
	}
	const q = `INSERT INTO saved_jobs (user_id, job_id) VALUES ($1, $2) RETURNING saved_at`
	out := &models.SavedJob{UserID: userID, JobID: jobID}
	if err := r.db.QueryRow(ctx, q, userID, jobID).Scan(&out.SavedAt); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrSavedJobDuplicate
		}
		if isForeignKeyViolation(err) {
			return nil, ErrJobNotFound
		}
		return nil, fmt.Errorf("repositories: save job: %w", err)
	}
	return out, nil
}

func (r *SavedJobRepository) ListByUserID(ctx context.Context, userID string) ([]*models.SavedJob, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidSavedJobInput
	}
	const q = `SELECT s.user_id, s.job_id, s.saved_at, j.id, j.external_id, j.title, j.company, j.location, j.description, j.salary, j.remote, j.source_url, j.created_at, j.updated_at FROM saved_jobs s JOIN jobs j ON j.id = s.job_id WHERE s.user_id = $1 ORDER BY s.saved_at DESC`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("repositories: list saved jobs: %w", err)
	}
	defer rows.Close()
	var out []*models.SavedJob
	for rows.Next() {
		s := &models.SavedJob{Job: &models.Job{}}
		if err := rows.Scan(&s.UserID, &s.JobID, &s.SavedAt, &s.Job.ID, &s.Job.ExternalID, &s.Job.Title, &s.Job.Company, &s.Job.Location, &s.Job.Description, &s.Job.Salary, &s.Job.Remote, &s.Job.SourceURL, &s.Job.CreatedAt, &s.Job.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repositories: scan saved job: %w", err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: iterate saved jobs: %w", err)
	}
	return out, nil
}

func (r *SavedJobRepository) Delete(ctx context.Context, userID, jobID string) error {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(jobID) == "" {
		return ErrInvalidSavedJobInput
	}
	const q = `DELETE FROM saved_jobs WHERE user_id = $1 AND job_id = $2`
	tag, err := r.db.Exec(ctx, q, userID, jobID)
	if err != nil {
		return fmt.Errorf("repositories: delete saved job: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrSavedJobNotFound
	}
	return nil
}

func isForeignKeyViolation(err error) bool {
	var pgErr interface{ SQLState() string }
	return errors.As(err, &pgErr) && pgErr.SQLState() == "23503"
}
