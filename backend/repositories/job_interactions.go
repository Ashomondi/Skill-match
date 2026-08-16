package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"skill-match/backend/models"
)

// Sentinel errors for job interaction operations.
var (
	ErrJobInteractionNotFound     = errors.New("repositories: job interaction not found")
	ErrInvalidJobInteractionInput = errors.New("repositories: invalid job interaction input")
)

// JobInteractionRepository provides persistence operations for user-job
// interactions backed by CockroachDB. It operates on the canonical
// models.JobInteraction type.
type JobInteractionRepository struct {
	db *pgxpool.Pool
}

func NewJobInteractionRepository(db *pgxpool.Pool) *JobInteractionRepository {
	return &JobInteractionRepository{db: db}
}

const jobInteractionColumns = `id, user_id, job_id, interaction_type, created_at`

// Create records a single job interaction for a user.
func (r *JobInteractionRepository) Create(ctx context.Context, in *models.JobInteraction) (*models.JobInteraction, error) {
	if in == nil || in.UserID == "" || in.JobID == "" || !in.Type.Valid() {
		return nil, fmt.Errorf("%w: user_id, job_id, and a valid interaction_type are required", ErrInvalidJobInteractionInput)
	}

	q := fmt.Sprintf(`
		INSERT INTO job_interactions (user_id, job_id, interaction_type)
		VALUES ($1, $2, $3)
		RETURNING %s`, jobInteractionColumns)

	row := r.db.QueryRow(ctx, q, in.UserID, in.JobID, in.Type)
	return scanJobInteraction(row)
}

// GetByID fetches an interaction by primary key. Returns
// ErrJobInteractionNotFound if no row matches.
func (r *JobInteractionRepository) GetByID(ctx context.Context, id string) (*models.JobInteraction, error) {
	q := fmt.Sprintf(`SELECT %s FROM job_interactions WHERE id = $1`, jobInteractionColumns)
	return scanJobInteraction(r.db.QueryRow(ctx, q, id))
}

// ListByUserID returns a user's interactions, most recent first. limit <= 0
// defaults to 100 to avoid unbounded scans on the caller's behalf.
func (r *JobInteractionRepository) ListByUserID(ctx context.Context, userID string, limit int) ([]*models.JobInteraction, error) {
	if limit <= 0 {
		limit = 100
	}

	q := fmt.Sprintf(`
		SELECT %s FROM job_interactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`, jobInteractionColumns)

	rows, err := r.db.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("repositories: list job interactions: %w", err)
	}
	defer rows.Close()

	var out []*models.JobInteraction
	for rows.Next() {
		ji, err := scanJobInteractionRow(rows)
		if err != nil {
			return nil, fmt.Errorf("repositories: scan job interaction row: %w", err)
		}
		out = append(out, ji)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: list job interactions: %w", err)
	}
	return out, nil
}

// Delete removes an interaction. Returns ErrJobInteractionNotFound if no row
// matches.
func (r *JobInteractionRepository) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM job_interactions WHERE id = $1`

	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("repositories: delete job interaction: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrJobInteractionNotFound
	}
	return nil
}

type jobInteractionRow interface {
	Scan(dest ...any) error
}

func scanJobInteraction(rw pgx.Row) (*models.JobInteraction, error) {
	ji, err := scanJobInteractionRow(rw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrJobInteractionNotFound
		}
		return nil, fmt.Errorf("repositories: query job interaction: %w", err)
	}
	return ji, nil
}

func scanJobInteractionRow(rw jobInteractionRow) (*models.JobInteraction, error) {
	ji := &models.JobInteraction{}
	var interactionType string
	if err := rw.Scan(&ji.ID, &ji.UserID, &ji.JobID, &interactionType, &ji.CreatedAt); err != nil {
		return nil, err
	}
	ji.Type = models.InteractionType(interactionType)
	return ji, nil
}
