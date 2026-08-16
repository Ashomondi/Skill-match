package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"skill-match/backend/models"
)

// Sentinel errors for application operations.
var (
	ErrApplicationNotFound      = errors.New("repositories: application not found")
	ErrInvalidApplicationInput  = errors.New("repositories: invalid application input")
	ErrInvalidApplicationStatus = errors.New("repositories: invalid application status")
)

// ApplicationRepository provides persistence operations for job applications
// backed by CockroachDB. It operates on the canonical models.Application type.
type ApplicationRepository struct {
	db *pgxpool.Pool
}

func NewApplicationRepository(db *pgxpool.Pool) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

const applicationColumns = `id, user_id, job_id, status, applied_at, created_at, updated_at`

// Create inserts a new application. jobID may be nil when the application is
// not tied to a job listing.
func (r *ApplicationRepository) Create(ctx context.Context, a *models.Application) (*models.Application, error) {
	if a == nil || a.UserID == "" || !a.Status.Valid() {
		return nil, fmt.Errorf("%w: user_id and a valid status are required", ErrInvalidApplicationInput)
	}

	q := fmt.Sprintf(`
		INSERT INTO applications (user_id, job_id, status)
		VALUES ($1, $2, $3)
		RETURNING %s`, applicationColumns)

	row := r.db.QueryRow(ctx, q, a.UserID, a.JobID, a.Status)
	return scanApplication(row)
}

// GetByID fetches an application by primary key. Returns
// ErrApplicationNotFound if no row matches.
func (r *ApplicationRepository) GetByID(ctx context.Context, id string) (*models.Application, error) {
	q := fmt.Sprintf(`SELECT %s FROM applications WHERE id = $1`, applicationColumns)
	return scanApplication(r.db.QueryRow(ctx, q, id))
}

// ListByUserID returns a user's applications, most recently updated first.
// limit <= 0 defaults to 100.
func (r *ApplicationRepository) ListByUserID(ctx context.Context, userID string, limit int) ([]*models.Application, error) {
	if limit <= 0 {
		limit = 100
	}

	q := fmt.Sprintf(`
		SELECT %s FROM applications
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2`, applicationColumns)

	rows, err := r.db.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("repositories: list applications: %w", err)
	}
	defer rows.Close()

	var out []*models.Application
	for rows.Next() {
		a, err := scanApplicationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("repositories: scan application row: %w", err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: list applications: %w", err)
	}
	return out, nil
}

// UpdateStatus transitions an application's status and bumps updated_at.
// Returns the updated row. Returns ErrApplicationNotFound if no row matches.
func (r *ApplicationRepository) UpdateStatus(ctx context.Context, id string, status models.ApplicationStatus) (*models.Application, error) {
	if !status.Valid() {
		return nil, fmt.Errorf("%w: %q", ErrInvalidApplicationStatus, status)
	}

	q := fmt.Sprintf(`
		UPDATE applications
		SET status = $1, updated_at = now()
		WHERE id = $2
		RETURNING %s`, applicationColumns)

	return scanApplication(r.db.QueryRow(ctx, q, status, id))
}

// Delete removes an application. Returns ErrApplicationNotFound if no row
// matches.
func (r *ApplicationRepository) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM applications WHERE id = $1`

	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("repositories: delete application: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrApplicationNotFound
	}
	return nil
}

type applicationRow interface {
	Scan(dest ...any) error
}

func scanApplication(rw pgx.Row) (*models.Application, error) {
	a, err := scanApplicationRow(rw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrApplicationNotFound
		}
		return nil, fmt.Errorf("repositories: query application: %w", err)
	}
	return a, nil
}

func scanApplicationRow(rw applicationRow) (*models.Application, error) {
	a := &models.Application{}
	var jobID *string
	var status string
	if err := rw.Scan(&a.ID, &a.UserID, &jobID, &status, &a.AppliedAt, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return nil, err
	}
	a.JobID = jobID
	a.Status = models.ApplicationStatus(status)
	return a, nil
}
