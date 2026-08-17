package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvalidInteractionInput = errors.New("repositories: invalid interaction input")

type InteractionType string

const (
	InteractionViewed    InteractionType = "viewed"
	InteractionSaved     InteractionType = "saved"
	InteractionApplied   InteractionType = "applied"
	InteractionDismissed InteractionType = "dismissed"
)

type JobInteraction struct {
	ID              string
	UserID          string
	JobID           string
	InteractionType InteractionType
	CreatedAt       time.Time
}

type JobInteractionRepository struct {
	db *pgxpool.Pool
}

func NewJobInteractionRepository(db *pgxpool.Pool) *JobInteractionRepository {
	return &JobInteractionRepository{db: db}
}

func (r *JobInteractionRepository) Record(ctx context.Context, userID, jobID string, t InteractionType) error {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(jobID) == "" {
		return fmt.Errorf("%w: user_id and job_id are required", ErrInvalidInteractionInput)
	}

	const q = `
		INSERT INTO job_interactions (user_id, job_id, interaction_type)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, job_id, interaction_type)
		DO UPDATE SET created_at = now()`

	if _, err := r.db.Exec(ctx, q, userID, jobID, t); err != nil {
		return fmt.Errorf("repositories: record interaction: %w", err)
	}
	return nil
}

func (r *JobInteractionRepository) InteractedJobIDs(ctx context.Context, userID string) (map[string]bool, error) {
	const q = `SELECT DISTINCT job_id FROM job_interactions WHERE user_id = $1`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("repositories: list interacted jobs: %w", err)
	}
	defer rows.Close()

	seen := make(map[string]bool)
	for rows.Next() {
		var jobID string
		if err := rows.Scan(&jobID); err != nil {
			return nil, fmt.Errorf("repositories: scan interacted job: %w", err)
		}
		seen[jobID] = true
	}
	return seen, rows.Err()
}
