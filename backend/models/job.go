package models

import "time"

type Job struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Company     string    `json:"company"`
	Location    string    `json:"location"`
	Description string    `json:"description"`
	Salary      string    `json:"salary,omitempty"`
	Remote      bool      `json:"remote"`
	SourceURL   string    `json:"source_url,omitempty"`
	ExternalID  string    `json:"external_id"` // dedup key from the source
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
