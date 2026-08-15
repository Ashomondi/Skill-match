package repositories

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"skill-match/backend/models"
	"strings"
)

var (
	ErrApplicationNotFound     = errors.New("repositories: application not found")
	ErrApplicationDuplicate    = errors.New("repositories: application already exists")
	ErrInvalidApplicationInput = errors.New("repositories: invalid application input")
)

type ApplicationRepository struct{ db *pgxpool.Pool }

func NewApplicationRepository(db *pgxpool.Pool) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

func (r *ApplicationRepository) Create(ctx context.Context, userID, jobID string) (*models.Application, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(jobID) == "" {
		return nil, ErrInvalidApplicationInput
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("repositories: begin application creation: %w", err)
	}
	defer tx.Rollback(ctx)
	const q = `INSERT INTO applications (user_id,job_id,status) VALUES ($1,$2,$3) RETURNING id,user_id,job_id,status,created_at,updated_at`
	a := &models.Application{}
	if err := tx.QueryRow(ctx, q, userID, jobID, models.ApplicationSaved).Scan(&a.ID, &a.UserID, &a.JobID, &a.Status, &a.CreatedAt, &a.UpdatedAt); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrApplicationDuplicate
		}
		if isForeignKeyViolation(err) {
			return nil, ErrJobNotFound
		}
		return nil, fmt.Errorf("repositories: create application: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO application_status_history (application_id,status) VALUES ($1,$2)`, a.ID, a.Status); err != nil {
		return nil, fmt.Errorf("repositories: record initial application status: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("repositories: commit application creation: %w", err)
	}
	return a, nil
}
func (r *ApplicationRepository) GetByID(ctx context.Context, userID, id string) (*models.Application, error) {
	const q = `SELECT id,user_id,job_id,status,created_at,updated_at FROM applications WHERE id=$1 AND user_id=$2`
	a := &models.Application{}
	err := r.db.QueryRow(ctx, q, id, userID).Scan(&a.ID, &a.UserID, &a.JobID, &a.Status, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrApplicationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repositories: get application: %w", err)
	}
	return a, nil
}
func (r *ApplicationRepository) UpdateStatus(ctx context.Context, userID, id string, status models.ApplicationStatus) (*models.Application, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("repositories: begin status update: %w", err)
	}
	defer tx.Rollback(ctx)
	a := &models.Application{}
	const uq = `UPDATE applications SET status=$1,updated_at=now() WHERE id=$2 AND user_id=$3 RETURNING id,user_id,job_id,status,created_at,updated_at`
	if err = tx.QueryRow(ctx, uq, status, id, userID).Scan(&a.ID, &a.UserID, &a.JobID, &a.Status, &a.CreatedAt, &a.UpdatedAt); errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrApplicationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repositories: update application status: %w", err)
	}
	const hq = `INSERT INTO application_status_history (application_id,status) VALUES ($1,$2)`
	if _, err = tx.Exec(ctx, hq, id, status); err != nil {
		return nil, fmt.Errorf("repositories: record application status: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("repositories: commit status update: %w", err)
	}
	return a, nil
}
func (r *ApplicationRepository) History(ctx context.Context, userID, id string) ([]models.ApplicationStatusChange, error) {
	if _, err := r.GetByID(ctx, userID, id); err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, `SELECT status,changed_at FROM application_status_history WHERE application_id=$1 ORDER BY changed_at ASC`, id)
	if err != nil {
		return nil, fmt.Errorf("repositories: application history: %w", err)
	}
	defer rows.Close()
	var out []models.ApplicationStatusChange
	for rows.Next() {
		var h models.ApplicationStatusChange
		if err := rows.Scan(&h.Status, &h.ChangedAt); err != nil {
			return nil, fmt.Errorf("repositories: scan application history: %w", err)
		}
		out = append(out, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: iterate application history: %w", err)
	}
	return out, nil
}
