package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"skill-match/backend/models"
	"skill-match/backend/utils"
)

var (
	ErrResumeUploadFailed = errors.New("resume upload failed")
)

type ResumeRepository interface {
	Create(ctx context.Context, resume *models.Resume) error
}

type ObjectStorage interface {
	Upload(
		ctx context.Context,
		key string,
		body io.Reader,
		contentType string,
	) (string, error)

	Delete(ctx context.Context, key string) error
}

type ResumeService struct {
	repository ResumeRepository
	storage    ObjectStorage
}

func NewResumeService(
	repository ResumeRepository,
	storage ObjectStorage,
) *ResumeService {
	return &ResumeService{
		repository: repository,
		storage:    storage,
	}
}

type UploadResumeInput struct {
	UserID   string
	Filename string
	File     io.Reader
	Size     int64
}

func (s *ResumeService) Upload(
	ctx context.Context,
	input UploadResumeInput,
) (*models.Resume, error) {

	// Read the complete file so validation and S3 upload
	// can safely use the same data.
	data, err := io.ReadAll(io.LimitReader(
		input.File,
		utils.MaxResumeSize+1,
	))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read file", ErrResumeUploadFailed)
	}

	actualSize := int64(len(data))

	if input.Size > 0 && input.Size != actualSize {
		return nil, fmt.Errorf("%w: invalid file size", ErrResumeUploadFailed)
	}

	// Validate file.
	if err := utils.ValidateResume(
		input.Filename,
		actualSize,
		bytes.NewReader(data),
	); err != nil {
		return nil, err
	}

	resumeID := uuid.NewString()

	extension := filepath.Ext(input.Filename)

	key := fmt.Sprintf(
		"resumes/%s/%s%s",
		input.UserID,
		resumeID,
		extension,
	)

	contentType := detectResumeContentType(extension)

	// Upload to S3 first.
	fileURL, err := s.storage.Upload(
		ctx,
		key,
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
		Status:    "processing",
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save metadata to CockroachDB.
	if err := s.repository.Create(ctx, resume); err != nil {

		// Compensating rollback:
		// S3 succeeded but DB failed, so remove the S3 object.
		_ = s.storage.Delete(ctx, key)

		return nil, fmt.Errorf(
			"%w: database operation failed: %v",
			ErrResumeUploadFailed,
			err,
		)
	}

	return resume, nil
}

func detectResumeContentType(extension string) string {
	switch extension {
	case ".pdf":
		return "application/pdf"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	default:
		return "application/octet-stream"
	}
}