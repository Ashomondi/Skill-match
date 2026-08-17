package models

import "time"

type SavedJob struct {
	UserID  string    `json:"user_id"`
	JobID   string    `json:"job_id"`
	SavedAt time.Time `json:"saved_at"`
	Job     *Job      `json:"job"`
}
