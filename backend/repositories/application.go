package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ApplicationStatus string

const (
	StatusApplied     ApplicationStatus = "applied"
	StatusInterviewing ApplicationStatus = "interviewing"
	StatusOffered     ApplicationStatus = "offered"
	StatusRejected    ApplicationStatus = "rejected"
	StatusWithdrawn   ApplicationStatus = "withdrawn"
)

func (s ApplicationStatus) IsValid() bool {
	switch s {
	case StatusApplied, StatusInterviewing, StatusOffered, StatusRejected, StatusWithdrawn:
		return true
	}
	return false
}

type Application struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	JobID     string            `json:"job_id"`
	Status    ApplicationStatus `json:"status"`
	Notes     string            `json:"notes"`
	AppliedAt time.Time         `json:"applied_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Job       *Job              `json:"job,omitempty"`
}

type ApplicationRepository struct {
	db *pgxpool.Pool
}

func NewApplicationRepository(db *pgxpool.Pool) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

func (r *ApplicationRepository) Create(ctx context.Context, userID, jobID string, status ApplicationStatus, notes string) (*Application, error) {
	if !status.IsValid() {
		status = StatusApplied
	}

	query := `
		INSERT INTO applications (user_id, job_id, status, notes, applied_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, user_id, job_id, status, notes, applied_at, updated_at
	`
	app := &Application{}
	err := r.db.QueryRow(ctx, query, userID, jobID, status, notes).Scan(
		&app.ID, &app.UserID, &app.JobID, &app.Status, &app.Notes, &app.AppliedAt, &app.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("repositories: create application: %w", err)
	}
	return app, nil
}

func (r *ApplicationRepository) ListByUserID(ctx context.Context, userID string) ([]*Application, error) {
	query := `
		SELECT a.id, a.user_id, a.job_id, a.status, a.notes, a.applied_at, a.updated_at,
		       j.id, j.external_id, j.title, j.company, j.location, j.description, j.salary, j.remote, j.source_url, j.created_at, j.updated_at
		FROM applications a
		LEFT JOIN jobs j ON a.job_id = j.id
		WHERE a.user_id = $1
		ORDER BY a.updated_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repositories: list applications: %w", err)
	}
	defer rows.Close()

	var apps []*Application
	for rows.Next() {
		app := &Application{}
		j := &Job{}
		if err := rows.Scan(
			&app.ID, &app.UserID, &app.JobID, &app.Status, &app.Notes, &app.AppliedAt, &app.UpdatedAt,
			&j.ID, &j.ExternalID, &j.Title, &j.Company, &j.Location,
			&j.Description, &j.Salary, &j.Remote, &j.SourceURL,
			&j.CreatedAt, &j.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repositories: scan application: %w", err)
		}
		app.Job = j
		apps = append(apps, app)
	}
	return apps, nil
}

func (r *ApplicationRepository) GetByID(ctx context.Context, id, userID string) (*Application, error) {
	query := `
		SELECT a.id, a.user_id, a.job_id, a.status, a.notes, a.applied_at, a.updated_at,
		       j.id, j.external_id, j.title, j.company, j.location, j.description, j.salary, j.remote, j.source_url, j.created_at, j.updated_at
		FROM applications a
		LEFT JOIN jobs j ON a.job_id = j.id
		WHERE a.id = $1 AND a.user_id = $2
	`
	app := &Application{}
	j := &Job{}
	err := r.db.QueryRow(ctx, query, id, userID).Scan(
		&app.ID, &app.UserID, &app.JobID, &app.Status, &app.Notes, &app.AppliedAt, &app.UpdatedAt,
		&j.ID, &j.ExternalID, &j.Title, &j.Company, &j.Location,
		&j.Description, &j.Salary, &j.Remote, &j.SourceURL,
		&j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("not_found")
	}
	app.Job = j
	return app, nil
}

func (r *ApplicationRepository) UpdateStatus(ctx context.Context, id, userID string, status ApplicationStatus, notes string) (*Application, error) {
	query := `
		UPDATE applications
		SET status = $1, notes = COALESCE(NULLIF($2, ''), notes), updated_at = NOW()
		WHERE id = $3 AND user_id = $4
		RETURNING id, user_id, job_id, status, notes, applied_at, updated_at
	`
	app := &Application{}
	err := r.db.QueryRow(ctx, query, status, notes, id, userID).Scan(
		&app.ID, &app.UserID, &app.JobID, &app.Status, &app.Notes, &app.AppliedAt, &app.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("not_found")
	}
	return app, nil
}

func (r *ApplicationRepository) Delete(ctx context.Context, id, userID string) error {
	query := `DELETE FROM applications WHERE id = $1 AND user_id = $2`
	cmd, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("repositories: delete application: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("not_found")
	}
	return nil
}