package models

import "time"

// ApplicationStatus mirrors the CHECK constraint in migrations/006_applications.sql.
// Keep in sync if the constraint changes.
type ApplicationStatus string

const (
	ApplicationStatusApplied   ApplicationStatus = "applied"
	ApplicationStatusScreening ApplicationStatus = "screening"
	ApplicationStatusInterview ApplicationStatus = "interview"
	ApplicationStatusOffer     ApplicationStatus = "offer"
	ApplicationStatusRejected  ApplicationStatus = "rejected"
	ApplicationStatusWithdrawn ApplicationStatus = "withdrawn"
)

func (s ApplicationStatus) Valid() bool {
	switch s {
	case ApplicationStatusApplied, ApplicationStatusScreening, ApplicationStatusInterview,
		ApplicationStatusOffer, ApplicationStatusRejected, ApplicationStatusWithdrawn:
		return true
	default:
		return false
	}
}

// Application is a user's job application and its current status. JobID is
// optional (the job may have been removed; the application is kept).
type Application struct {
	ID        string
	UserID    string
	JobID     *string
	Status    ApplicationStatus
	AppliedAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
