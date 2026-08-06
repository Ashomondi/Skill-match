package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Sentinel errors for resume operations.
var (
	ErrResumeNotFound     = errors.New("repositories: resume not found")
	ErrInvalidResumeInput = errors.New("repositories: invalid resume input")
	ErrInvalidResumeStatus = errors.New("repositories: invalid resume status transition")
)

// ResumeStatus mirrors the CHECK constraint in
// migrations/002_resume.sql. Keep in sync if the constraint changes.
type ResumeStatus string

const (
	ResumeStatusUploaded ResumeStatus = "uploaded"
	ResumeStatusParsing  ResumeStatus = "parsing"
	ResumeStatusParsed   ResumeStatus = "parsed"
	ResumeStatusFailed   ResumeStatus = "failed"
)

func (s ResumeStatus) valid() bool {
	switch s {
	case ResumeStatusUploaded, ResumeStatusParsing, ResumeStatusParsed, ResumeStatusFailed:
		return true
	default:
		return false
	}
}

// Resume is the persistence-layer representation of a resume row.
//
// NOTE: models.Resume (Issue 6, owned by Ashley) is expected to become the
// canonical type. Same caveat as repositories/user.go — replace this local
// definition with models.Resume once that file exists.
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

// ResumeRepository provides persistence operations for resumes backed by
// CockroachDB.
type ResumeRepository struct {
	db *pgxpool.Pool
}

// NewResumeRepository constructs a ResumeRepository.
func NewResumeRepository(db *pgxpool.Pool) *ResumeRepository {
	return &ResumeRepository{db: db}
}

const resumeColumns = `id, user_id, s3_key, original_filename, content_type,
	file_size_bytes, status, parsed_text, failure_reason, created_at, updated_at`

// Create inserts a new resume row in the "uploaded" state. Called after the
// file has already been written to S3 (Issue 8); s3Key must reference an
// object that exists.
func (r *ResumeRepository) Create(ctx context.Context, res *Resume) (*Resume, error) {
	if res == nil || res.UserID == "" || res.S3Key == "" || res.OriginalFilename == "" {
		return nil, fmt.Errorf("%w: user_id, s3_key, and original_filename are required", ErrInvalidResumeInput)
	}
	if res.FileSizeBytes <= 0 {
		return nil, fmt.Errorf("%w: file_size_bytes must be positive", ErrInvalidResumeInput)
	}

	q := fmt.Sprintf(`
		INSERT INTO resumes (user_id, s3_key, original_filename, content_type, file_size_bytes, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING %s`, resumeColumns)

	row := r.db.QueryRow(ctx, q,
		res.UserID, res.S3Key, res.OriginalFilename, res.ContentType,
		res.FileSizeBytes, ResumeStatusUploaded)

	return scanResume(row)
}

// GetByID fetches a resume by primary key. Returns ErrResumeNotFound if no
// row matches.
func (r *ResumeRepository) GetByID(ctx context.Context, id string) (*Resume, error) {
	q := fmt.Sprintf(`SELECT %s FROM resumes WHERE id = $1`, resumeColumns)
	return scanResume(r.db.QueryRow(ctx, q, id))
}

// ListByUserID returns resumes for a user, most recent first. limit <= 0
// defaults to 50 to avoid unbounded scans on the caller's behalf.
func (r *ResumeRepository) ListByUserID(ctx context.Context, userID string, limit int) ([]*Resume, error) {
	if limit <= 0 {
		limit = 50
	}

	q := fmt.Sprintf(`
		SELECT %s FROM resumes
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`, resumeColumns)

	rows, err := r.db.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("repositories: list resumes: %w", err)
	}
	defer rows.Close()

	var out []*Resume
	for rows.Next() {
		res, err := scanResumeFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("repositories: scan resume row: %w", err)
		}
		out = append(out, res)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: list resumes: %w", err)
	}
	return out, nil
}

// UpdateStatus transitions a resume's processing status. When status is
// ResumeStatusParsed, parsedText should be provided; when
// ResumeStatusFailed, failureReason should be provided. Either may be nil
// otherwise. Returns ErrResumeNotFound if no row matches id.
func (r *ResumeRepository) UpdateStatus(ctx context.Context, id string, status ResumeStatus, parsedText, failureReason *string) error {
	if !status.valid() {
		return fmt.Errorf("%w: %q", ErrInvalidResumeStatus, status)
	}

	const q = `
		UPDATE resumes
		SET status = $1, parsed_text = $2, failure_reason = $3, updated_at = now()
		WHERE id = $4`

	tag, err := r.db.Exec(ctx, q, status, parsedText, failureReason, id)
	if err != nil {
		return fmt.Errorf("repositories: update resume status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrResumeNotFound
	}
	return nil
}

// Delete removes a resume row. The caller is responsible for deleting the
// corresponding S3 object beforehand or in the same transaction/workflow;
// this repository does not reach into S3.
func (r *ResumeRepository) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM resumes WHERE id = $1`

	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("repositories: delete resume: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrResumeNotFound
	}
	return nil
}

// row is the subset of pgx.Row/pgx.Rows behavior scanResume needs, letting
// it serve both QueryRow (single) and Query (iterated) call sites.
type row interface {
	Scan(dest ...any) error
}

func scanResume(rw pgx.Row) (*Resume, error) {
	res, err := scanResumeFromRows(rw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResumeNotFound
		}
		return nil, fmt.Errorf("repositories: query resume: %w", err)
	}
	return res, nil
}

func scanResumeFromRows(rw row) (*Resume, error) {
	res := &Resume{}
	var status string
	err := rw.Scan(&res.ID, &res.UserID, &res.S3Key, &res.OriginalFilename,
		&res.ContentType, &res.FileSizeBytes, &status, &res.ParsedText,
		&res.FailureReason, &res.CreatedAt, &res.UpdatedAt)
	if err != nil {
		return nil, err
	}
	res.Status = ResumeStatus(status)
	return res, nil
}