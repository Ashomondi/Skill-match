package models

import "time"

type Resume struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Filename   string    `json:"filename"`
	FileURL    string    `json:"file_url"`
	FileType   string    `json:"file_type"`
	FileSize   int64     `json:"file_size"`
	Status     string    `json:"status"`
	Version    int       `json:"version"`
	ParsedText *string   `json:"parsed_text,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
