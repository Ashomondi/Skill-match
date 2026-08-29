package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"skill-match/backend/models"
)

var (
	ErrApplicationNotFound      = errors.New("repositories: application not found")
	ErrApplicationConflict      = errors.New("repositories: application already exists")
	ErrApplicationAlreadyExists = ErrApplicationConflict
)

type ApplicationRepository struct {
	pool *pgxpool.Pool
}

func NewApplicationRepository(pool *pgxpool.Pool) *ApplicationRepository {
	return &ApplicationRepository{pool: pool}
}

const applicationColumns = `
	id, user_id, job_id, status, created_at, updated_at
`

func (r *ApplicationRepository) Create(ctx context.Context, userID, jobID string) (*models.Application, error) {
	const q = `
		INSERT INTO applications (user_id, job_id)
		VALUES ($1, $2)
		RETURNING ` + applicationColumns

	out := &models.Application{}
	if err := r.pool.QueryRow(ctx, q, userID, jobID).Scan(
		&out.ID, &out.UserID, &out.JobID, &out.Status, &out.CreatedAt, &out.UpdatedAt,
	); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrApplicationConflict
		}
		if isForeignKeyViolation(err) {
			return nil, ErrApplicationNotFound
		}
		return nil, fmt.Errorf("repositories: create application: %w", err)
	}
	return out, nil
}

func (r *ApplicationRepository) GetByID(ctx context.Context, userID, id string) (*models.Application, error) {
	const q = `
		SELECT ` + applicationColumns + `
		FROM applications
		WHERE id = $1 AND user_id = $2`

	out := &models.Application{}
	if err := r.pool.QueryRow(ctx, q, id, userID).Scan(
		&out.ID, &out.UserID, &out.JobID, &out.Status, &out.CreatedAt, &out.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrApplicationNotFound
		}
		return nil, fmt.Errorf("repositories: get application: %w", err)
	}
	return out, nil
}

func (r *ApplicationRepository) UpdateStatus(ctx context.Context, userID, id string, status models.ApplicationStatus) (*models.Application, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("repositories: begin application update: %w", err)
	}
	defer tx.Rollback(ctx)

	const updateQuery = `
		UPDATE applications
		SET status = $1, updated_at = now()
		WHERE id = $2 AND user_id = $3
		RETURNING ` + applicationColumns

	out := &models.Application{}
	if err := tx.QueryRow(ctx, updateQuery, status, id, userID).Scan(
		&out.ID, &out.UserID, &out.JobID, &out.Status, &out.CreatedAt, &out.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrApplicationNotFound
		}
		return nil, fmt.Errorf("repositories: update application: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO application_status_history (application_id, status)
		VALUES ($1, $2)`, out.ID, status); err != nil {
		return nil, fmt.Errorf("repositories: record application status: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("repositories: commit application update: %w", err)
	}
	return out, nil
}

func (r *ApplicationRepository) Delete(ctx context.Context, userID, id string) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM applications
		WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("repositories: delete application: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrApplicationNotFound
	}
	return nil
}

func (r *ApplicationRepository) History(ctx context.Context, userID, id string) ([]models.ApplicationStatusChange, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT h.status, h.changed_at
		FROM application_status_history h
		JOIN applications a ON a.id = h.application_id
		WHERE h.application_id = $1 AND a.user_id = $2
		ORDER BY h.changed_at ASC`, id, userID)
	if err != nil {
		return nil, fmt.Errorf("repositories: application history: %w", err)
	}
	defer rows.Close()

	history := make([]models.ApplicationStatusChange, 0)
	for rows.Next() {
		var change models.ApplicationStatusChange
		if err := rows.Scan(&change.Status, &change.ChangedAt); err != nil {
			return nil, fmt.Errorf("repositories: scan application history: %w", err)
		}
		history = append(history, change)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: iterate application history: %w", err)
	}
	return history, nil
}

func (r *ApplicationRepository) ListByUserID(ctx context.Context, userID string) ([]*models.Application, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT a.id, a.user_id, a.job_id, a.status, a.created_at, a.updated_at,
		       j.id, j.external_id, j.title, j.company, j.location, j.description,
		       j.salary, j.remote, j.source_url, j.created_at, j.updated_at
		FROM applications a
		JOIN jobs j ON j.id = a.job_id
		WHERE a.user_id = $1
		ORDER BY a.updated_at DESC, a.id DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("repositories: list applications: %w", err)
	}
	defer rows.Close()

	applications := make([]*models.Application, 0)
	for rows.Next() {
		app := &models.Application{Job: &models.Job{}}
		if err := rows.Scan(
			&app.ID, &app.UserID, &app.JobID, &app.Status, &app.CreatedAt, &app.UpdatedAt,
			&app.Job.ID, &app.Job.ExternalID, &app.Job.Title, &app.Job.Company,
			&app.Job.Location, &app.Job.Description, &app.Job.Salary, &app.Job.Remote,
			&app.Job.SourceURL, &app.Job.CreatedAt, &app.Job.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repositories: scan application: %w", err)
		}
		applications = append(applications, app)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: iterate applications: %w", err)
	}
	return applications, nil
}
