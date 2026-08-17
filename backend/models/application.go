package models

import "time"

type ApplicationStatus string

const (
	ApplicationSaved     ApplicationStatus = "saved"
	ApplicationApplied   ApplicationStatus = "applied"
	ApplicationScreening ApplicationStatus = "screening"
	ApplicationInterview ApplicationStatus = "interview"
	ApplicationOffer     ApplicationStatus = "offer"
	ApplicationRejected  ApplicationStatus = "rejected"
	ApplicationWithdrawn ApplicationStatus = "withdrawn"
)

func (s ApplicationStatus) Valid() bool {
	switch s {
	case ApplicationSaved, ApplicationApplied, ApplicationScreening, ApplicationInterview, ApplicationOffer, ApplicationRejected, ApplicationWithdrawn:
		return true
	}
	return false
}

type ApplicationStatusChange struct {
	Status    ApplicationStatus `json:"status"`
	ChangedAt time.Time         `json:"changed_at"`
}
type Application struct {
	ID        string                    `json:"id"`
	UserID    string                    `json:"user_id"`
	JobID     string                    `json:"job_id"`
	Status    ApplicationStatus         `json:"status"`
	CreatedAt time.Time                 `json:"created_at"`
	UpdatedAt time.Time                 `json:"updated_at"`
	History   []ApplicationStatusChange `json:"history,omitempty"`
}
