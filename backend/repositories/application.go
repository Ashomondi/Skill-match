package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"skill-match/backend/models"
)

// ApplicationStatus aliases the canonical model status type so the values
// match the applications CHECK constraint and the frontend (saved|applied|
// screening|interview|offer|rejected|withdrawn).
type ApplicationStatus = models.ApplicationStatus

var (
	ErrApplicationNotFound     = errors.New("repositories: application not found")
	ErrInvalidApplicationInput = errors.New("repositories: invalid application input")
)

// (fix: repair the feat/data merge — jobs search/matching, ingestion, and auth)
type ApplicationRepository struct {
	pool *pgxpool.Pool
}

func NewApplicationRepository(pool *pgxpool.Pool) *ApplicationRepository {
	return &ApplicationRepository{pool: pool}
}

// Create inserts a new application and records the initial status history
// row. notes is accepted for API compatibility but not persisted (the
// applications table has no notes column).
func (r *ApplicationRepository) Create(ctx context.Context, userID, jobID string, status ApplicationStatus, notes string) (*models.Application, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(jobID) == "" {
		return nil, ErrInvalidApplicationInput
	}
	if !status.Valid() {
		status = models.ApplicationApplied
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("repositories: begin application creation: %w", err)
	}
	defer tx.Rollback(ctx)

	const q = `INSERT INTO applications (user_id, job_id, status) VALUES ($1,$2,$3) RETURNING id, user_id, job_id, status, created_at, updated_at`
	a := &models.Application{}
	if err := tx.QueryRow(ctx, q, userID, jobID, status).Scan(&a.ID, &a.UserID, &a.JobID, &a.Status, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return nil, fmt.Errorf("repositories: create application: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO application_status_history (application_id, status) VALUES ($1,$2)`, a.ID, a.Status); err != nil {
		return nil, fmt.Errorf("repositories: record initial application status: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("repositories: commit application creation: %w", err)
	}
	return a, nil
}

// ListByUserID returns a user's applications, most recently updated first,
// each enriched with its job's title and company.
func (r *ApplicationRepository) ListByUserID(ctx context.Context, userID string) ([]*models.Application, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidApplicationInput
	}

	const q = `
		SELECT a.id, a.user_id, a.job_id, a.status, a.created_at, a.updated_at,
		       j.id, j.title, j.company
		FROM applications a
		LEFT JOIN jobs j ON j.id = a.job_id
		WHERE a.user_id = $1
		ORDER BY a.updated_at DESC`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("repositories: list applications: %w", err)
	}
	defer rows.Close()

	var out []*models.Application
	for rows.Next() {
		a := &models.Application{}
		var jobID, jobTitle, jobCompany *string
		if err := rows.Scan(&a.ID, &a.UserID, &a.JobID, &a.Status, &a.CreatedAt, &a.UpdatedAt,
			&jobID, &jobTitle, &jobCompany); err != nil {
			return nil, fmt.Errorf("repositories: scan application row: %w", err)
		}
		if jobID != nil {
			a.Job = &models.Job{ID: *jobID, Title: valueOr(jobTitle, ""), Company: valueOr(jobCompany, "")}
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: iterate applications: %w", err)
	}
	return out, nil
}

// GetByID fetches one of the user's applications with its job details.
func (r *ApplicationRepository) GetByID(ctx context.Context, id, userID string) (*models.Application, error) {
	const q = `
		SELECT a.id, a.user_id, a.job_id, a.status, a.created_at, a.updated_at,
		       j.id, j.title, j.company
		FROM applications a
		LEFT JOIN jobs j ON j.id = a.job_id
		WHERE a.id = $1 AND a.user_id = $2`

	a := &models.Application{}
	var jobID, jobTitle, jobCompany *string
	err := r.db.QueryRow(ctx, q, id, userID).Scan(&a.ID, &a.UserID, &a.JobID, &a.Status, &a.CreatedAt, &a.UpdatedAt,
		&jobID, &jobTitle, &jobCompany)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrApplicationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repositories: get application: %w", err)
	}
	if jobID != nil {
		a.Job = &models.Job{ID: *jobID, Title: valueOr(jobTitle, ""), Company: valueOr(jobCompany, "")}
	}
	return a, nil
}

// UpdateStatus transitions an application's status, records the change in
// the history table, and bumps updated_at.
func (r *ApplicationRepository) UpdateStatus(ctx context.Context, id, userID string, status ApplicationStatus, notes string) (*models.Application, error) {
	if !status.Valid() {
		return nil, fmt.Errorf("repositories: invalid application status %q", status)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("repositories: begin status update: %w", err)
	}
	defer tx.Rollback(ctx)

	a := &models.Application{}
	const uq = `UPDATE applications SET status=$1, updated_at=now() WHERE id=$2 AND user_id=$3 RETURNING id, user_id, job_id, status, created_at, updated_at`
	err = tx.QueryRow(ctx, uq, status, id, userID).Scan(&a.ID, &a.UserID, &a.JobID, &a.Status, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrApplicationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repositories: update application status: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO application_status_history (application_id, status) VALUES ($1,$2)`, id, status); err != nil {
		return nil, fmt.Errorf("repositories: record application status: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("repositories: commit status update: %w", err)
	}
	return a, nil
}

// Delete removes one of the user's applications.
func (r *ApplicationRepository) Delete(ctx context.Context, id, userID string) error {
	cmd, err := r.db.Exec(ctx, `DELETE FROM applications WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return fmt.Errorf("repositories: delete application: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrApplicationNotFound
	}
	return nil
}

func valueOr(p *string, fallback string) string {
	if p == nil {
		return fallback
	}
	return *p
}
// (fix: repair the feat/data merge — jobs search/matching, ingestion, and auth)
