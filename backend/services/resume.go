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
	ErrResumeDeleteFailed = errors.New("resume delete failed")
	ErrResumeNotFound     = errors.New("resume not found")
	ErrResumeUnauthorized = errors.New("resume does not belong to user")
)

type ResumeRepository interface {
	Create(ctx context.Context, resume *models.Resume) error

	GetByID(ctx context.Context, resumeID string) (*models.Resume, error)

	Update(ctx context.Context, resume *models.Resume) error

	Delete(ctx context.Context, resumeID string) error
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

	if input.Size > 0 && input.Size != actualSize {
		return nil, fmt.Errorf(
			"%w: invalid file size",
			ErrResumeUploadFailed,
		)
	}

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

	if err := s.repository.Create(ctx, resume); err != nil {
		_ = s.storage.Delete(ctx, s3Key)

		return nil, fmt.Errorf(
			"%w: database operation failed: %v",
			ErrResumeUploadFailed,
			err,
		)
	}

	return resume, nil
}

type UpdateResumeInput struct {
	UserID   string
	ResumeID string
	Filename string
	File     io.Reader
	Size     int64
}

func (s *ResumeService) Update(
	ctx context.Context,
	input UpdateResumeInput,
) (*models.Resume, error) {

	existing, err := s.repository.GetByID(
		ctx,
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

	if existing.UserID != input.UserID {
		return nil, ErrResumeUnauthorized
	}

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

	newVersion := existing.Version + 1

	newS3Key := fmt.Sprintf(
		"resumes/%s/%s-v%d%s",
		input.UserID,
		existing.ID,
		newVersion,
		newExtension,
	)

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

func (s *ResumeService) Delete(
	ctx context.Context,
	userID string,
	resumeID string,
) error {

	resume, err := s.repository.GetByID(
		ctx,
		resumeID,
	)
	if err != nil {
		if errors.Is(err, ErrResumeNotFound) {
			return ErrResumeNotFound
		}

		return fmt.Errorf(
			"%w: failed to find resume: %v",
			ErrResumeDeleteFailed,
			err,
		)
	}

	if resume == nil {
		return ErrResumeNotFound
	}

	if resume.UserID != userID {
		return ErrResumeUnauthorized
	}

	s3Key := buildResumeS3Key(resume)

	if s3Key != "" {
		if err := s.storage.Delete(ctx, s3Key); err != nil {
			return fmt.Errorf(
				"%w: failed to delete resume file: %v",
				ErrResumeDeleteFailed,
				err,
			)
		}
	}

	if err := s.repository.Delete(ctx, resumeID); err != nil {
		return fmt.Errorf(
			"%w: failed to delete resume metadata: %v",
			ErrResumeDeleteFailed,
			err,
		)
	}

	return nil
}

func buildResumeS3Key(resume *models.Resume) string {
	if resume == nil {
		return ""
	}

	if strings.TrimSpace(resume.UserID) == "" ||
		strings.TrimSpace(resume.ID) == "" {
		return ""
	}

	extension := strings.ToLower(
		resume.FileType,
	)

	if extension == "" {
		extension = strings.ToLower(
			filepath.Ext(resume.Filename),
		)
	}

	if resume.Version <= 1 {
		return fmt.Sprintf(
			"resumes/%s/%s%s",
			resume.UserID,
			resume.ID,
			extension,
		)
	}

	return fmt.Sprintf(
		"resumes/%s/%s-v%d%s",
		resume.UserID,
		resume.ID,
		resume.Version,
		extension,
	)
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
