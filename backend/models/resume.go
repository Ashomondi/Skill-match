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
	ID               string
	UserID           string
	S3Key            string
	OriginalFilename string
	ContentType      string
	FileSizeBytes    int64
	Status           ResumeStatus
	ParsedText       *string
	FailureReason    *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
