package models

import "time"

// InteractionType mirrors the CHECK constraint in migrations/004_jobs.sql.
// Keep in sync if the constraint changes.
type InteractionType string

const (
	InteractionView    InteractionType = "view"
	InteractionSave    InteractionType = "save"
	InteractionApply   InteractionType = "apply"
	InteractionDismiss InteractionType = "dismiss"
	InteractionSearch  InteractionType = "search"
)

func (t InteractionType) Valid() bool {
	switch t {
	case InteractionView, InteractionSave, InteractionApply, InteractionDismiss, InteractionSearch:
		return true
	default:
		return false
	}
}

// JobInteraction records a user's interaction with a job. The collected
// history is a signal used by the recommendation engine.
type JobInteraction struct {
	ID        string
	UserID    string
	JobID     string
	Type      InteractionType
	CreatedAt time.Time
}
