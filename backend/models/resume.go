package models

import "time"

// ResumeStatus mirrors the CHECK constraint in migrations/002_resume.sql.
// Keep in sync if the constraint changes.
type ResumeStatus string

const (
	ResumeStatusUploaded ResumeStatus = "uploaded"
	ResumeStatusParsing  ResumeStatus = "parsing"
	ResumeStatusParsed   ResumeStatus = "parsed"
	ResumeStatusFailed   ResumeStatus = "failed"
)

func (s ResumeStatus) Valid() bool {
	switch s {
	case ResumeStatusUploaded, ResumeStatusParsing, ResumeStatusParsed, ResumeStatusFailed:
		return true
	default:
		return false
	}
}

// Resume is a user's uploaded resume. Binary content lives in S3; this
// model holds metadata and processing state only.
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
