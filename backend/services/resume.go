package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"skill-match/backend/models"
	"skill-match/backend/utils"
)

var (
	ErrResumeUploadFailed = errors.New("resume upload failed")
	ErrResumeUpdateFailed = errors.New("resume update failed")
	ErrResumeNotFound     = errors.New("resume not found")
	ErrResumeUnauthorized = errors.New("resume does not belong to user")
)

// ResumeRepository defines the database operations required
// by the resume service.
//
// Sospeter's repository implementation must satisfy this interface.
type ResumeRepository interface {
	Create(ctx context.Context, resume *models.Resume) error
	GetByID(ctx context.Context, userID, resumeID string) (*models.Resume, error)
	Update(ctx context.Context, resume *models.Resume) error
}

// ObjectStorage defines the S3 operations required by the
// resume service.
type ObjectStorage interface {
	Upload(
		ctx context.Context,
		key string,
		body io.Reader,
		contentType string,
	) (string, error)

	Delete(ctx context.Context, key string) error
}

// ResumeService handles resume business logic.
type ResumeService struct {
	repository ResumeRepository
	storage    ObjectStorage
}

// NewResumeService creates a new ResumeService.
func NewResumeService(
	repository ResumeRepository,
	storage ObjectStorage,
) *ResumeService {
	return &ResumeService{
		repository: repository,
		storage:    storage,
	}
}

// UploadResumeInput contains the information needed to upload a resume.
type UploadResumeInput struct {
	UserID   string
	Filename string
	File     io.Reader
	Size     int64
}

// Upload uploads a new resume to S3 and stores its metadata
// in CockroachDB.
func (s *ResumeService) Upload(
	ctx context.Context,
	input UploadResumeInput,
) (*models.Resume, error) {

	// Read the file while preventing oversized files from
	// being loaded into memory.
	data, err := io.ReadAll(
		io.LimitReader(
			input.File,
			utils.MaxResumeSize+1,
		),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: failed to read file",
			ErrResumeUploadFailed,
		)
	}

	actualSize := int64(len(data))

	// Validate the file size reported by the multipart request.
	if input.Size > 0 && input.Size != actualSize {
		return nil, fmt.Errorf(
			"%w: invalid file size",
			ErrResumeUploadFailed,
		)
	}

	// Validate file type, size and content.
	if err := utils.ValidateResume(
		input.Filename,
		actualSize,
		bytes.NewReader(data),
	); err != nil {
		return nil, err
	}

	resumeID := uuid.NewString()

	extension := strings.ToLower(
		filepath.Ext(input.Filename),
	)

	s3Key := fmt.Sprintf(
		"resumes/%s/%s%s",
		input.UserID,
		resumeID,
		extension,
	)

	contentType := detectResumeContentType(extension)

	// Upload the file to S3 first.
	fileURL, err := s.storage.Upload(
		ctx,
		s3Key,
		bytes.NewReader(data),
		contentType,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: storage upload failed: %v",
			ErrResumeUploadFailed,
			err,
		)
	}

	now := time.Now().UTC()

	resume := &models.Resume{
		ID:        resumeID,
		UserID:    input.UserID,
		Filename:  input.Filename,
		FileURL:   fileURL,
		FileType:  extension,
		FileSize:  actualSize,
		Status:    "uploaded",
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save metadata to CockroachDB.
	if err := s.repository.Create(ctx, resume); err != nil {

		// Database failed after S3 succeeded.
		// Remove the uploaded object so we don't leave
		// an orphaned file in S3.
		_ = s.storage.Delete(ctx, s3Key)

		return nil, fmt.Errorf(
			"%w: database operation failed: %v",
			ErrResumeUploadFailed,
			err,
		)
	}

	return resume, nil
}

// UpdateResumeInput contains the information needed to replace
// an existing resume.
type UpdateResumeInput struct {
	UserID   string
	ResumeID string
	Filename string
	File     io.Reader
	Size     int64
}

// Update replaces an existing resume with a new file.
func (s *ResumeService) Update(
	ctx context.Context,
	input UpdateResumeInput,
) (*models.Resume, error) {

	// Retrieve the resume while scoping the lookup to the
	// authenticated user.
	existing, err := s.repository.GetByID(
		ctx,
		input.UserID,
		input.ResumeID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: %v",
			ErrResumeNotFound,
			err,
		)
	}

	if existing == nil {
		return nil, ErrResumeNotFound
	}

	// Extra ownership check.
	if existing.UserID != input.UserID {
		return nil, ErrResumeUnauthorized
	}

	// Read the replacement file.
	data, err := io.ReadAll(
		io.LimitReader(
			input.File,
			utils.MaxResumeSize+1,
		),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: failed to read replacement file",
			ErrResumeUpdateFailed,
		)
	}

	actualSize := int64(len(data))

	if input.Size > 0 && input.Size != actualSize {
		return nil, fmt.Errorf(
			"%w: invalid file size",
			ErrResumeUpdateFailed,
		)
	}

	// Validate the replacement before modifying S3.
	if err := utils.ValidateResume(
		input.Filename,
		actualSize,
		bytes.NewReader(data),
	); err != nil {
		return nil, err
	}

	newExtension := strings.ToLower(
		filepath.Ext(input.Filename),
	)

	contentType := detectResumeContentType(newExtension)

	// Increase the resume version.
	newVersion := existing.Version + 1

	// Create a new S3 object for the replacement.
	newS3Key := fmt.Sprintf(
		"resumes/%s/%s-v%d%s",
		input.UserID,
		existing.ID,
		newVersion,
		newExtension,
	)

	// Upload the new file first.
	newFileURL, err := s.storage.Upload(
		ctx,
		newS3Key,
		bytes.NewReader(data),
		contentType,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: storage upload failed: %v",
			ErrResumeUpdateFailed,
			err,
		)
	}

	updated := &models.Resume{
		ID:        existing.ID,
		UserID:    existing.UserID,
		Filename:  input.Filename,
		FileURL:   newFileURL,
		FileType:  newExtension,
		FileSize:  actualSize,
		Status:    "uploaded",
		Version:   newVersion,
		CreatedAt: existing.CreatedAt,
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.repository.Update(ctx, updated); err != nil {

		_ = s.storage.Delete(ctx, newS3Key)

		return nil, fmt.Errorf(
			"%w: database operation failed: %v",
			ErrResumeUpdateFailed,
			err,
		)
	}


	return updated, nil
}

func detectResumeContentType(extension string) string {
	switch strings.ToLower(extension) {
	case ".pdf":
		return "application/pdf"

	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"

	default:
		return "application/octet-stream"
	}
}