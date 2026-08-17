package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"time"

	"skill-match/backend/models"
	"skill-match/backend/repositories"
	"skill-match/backend/utils"
)

var (
<<<<<<< HEAD
	ErrResumeNotFound     = errors.New("resume not found")
	ErrResumeAccessDenied = errors.New("resume access denied")
	ErrInvalidResume      = errors.New("invalid resume")
)

// ObjectStorage is the subset of the S3 client the resume service needs.
// clients.S3Client satisfies this interface.
=======
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

>>>>>>> dev
type ObjectStorage interface {
	Put(ctx context.Context, key string, body []byte, contentType string) error
	PresignDownload(ctx context.Context, key string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
	Key(userID, fileID string) string
}

// ResumeRepository defines the persistence operations required by the resume
// service. repositories.ResumeRepository satisfies this interface.
type ResumeRepository interface {
	Create(ctx context.Context, res *models.Resume) (*models.Resume, error)
	GetByID(ctx context.Context, id string) (*models.Resume, error)
	ListByUserID(ctx context.Context, userID string, limit int) ([]*models.Resume, error)
	Delete(ctx context.Context, id string) error
}

// ResumeService coordinates resume storage (S3) and metadata persistence
// (CockroachDB). It enforces user ownership on every operation.
type ResumeService struct {
	repo    ResumeRepository
	storage ObjectStorage
}

func NewResumeService(repo ResumeRepository, storage ObjectStorage) *ResumeService {
	return &ResumeService{repo: repo, storage: storage}
}

// Upload validates and stores a resume: bytes go to S3 first, then the
// metadata row is persisted. If the DB insert fails after the S3 upload, the
// object is deleted so no orphan is left behind. When replaceID is set, the
// user's previous resume (and its object) is removed after the new one is
// safely stored.
func (s *ResumeService) Upload(ctx context.Context, userID, replaceID, filename, contentType string, data []byte) (*models.Resume, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user is required", ErrInvalidResume)
	}

	if err := utils.ValidateResumeFile(filename, contentType, int64(len(data))); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResume, err)
	}

	if replaceID != "" {
		if _, err := s.getOwned(ctx, userID, replaceID); err != nil {
			return nil, err
		}
	}

<<<<<<< HEAD
	fileID, err := utils.GenerateFileID(filename)
	if err != nil {
		return nil, err
	}
	key := s.storage.Key(userID, fileID)

	if err := s.storage.Put(ctx, key, data, contentType); err != nil {
		return nil, fmt.Errorf("upload resume to storage: %w", err)
	}

	res := &models.Resume{
		UserID:           userID,
		S3Key:            key,
		OriginalFilename: filename,
		ContentType:      contentType,
		FileSizeBytes:    int64(len(data)),
		Status:           models.ResumeStatusUploaded,
	}

	created, err := s.repo.Create(ctx, res)
	if err != nil {
		// Roll back the object so a failed row doesn't leak storage.
		_ = s.storage.Delete(ctx, key)
		return nil, err
	}

	if replaceID != "" {
		old, err := s.repo.GetByID(ctx, replaceID)
		if err == nil && old != nil {
			_ = s.storage.Delete(ctx, old.S3Key)
			_ = s.repo.Delete(ctx, replaceID)
		}
	}

	return created, nil
}

// List returns all resumes owned by userID, most recent first.
func (s *ResumeService) List(ctx context.Context, userID string) ([]*models.Resume, error) {
	return s.repo.ListByUserID(ctx, userID, 0)
}

// DownloadURL returns the resume and a time-limited presigned download URL.
func (s *ResumeService) DownloadURL(ctx context.Context, userID, id string, expiry time.Duration) (*models.Resume, string, error) {
	res, err := s.getOwned(ctx, userID, id)
	if err != nil {
		return nil, "", err
	}

	url, err := s.storage.PresignDownload(ctx, res.S3Key, expiry)
	if err != nil {
		return nil, "", fmt.Errorf("presign resume download: %w", err)
	}

	return res, url, nil
}

// Delete removes a resume the user owns: the S3 object first, then the row.
func (s *ResumeService) Delete(ctx context.Context, userID, id string) error {
	res, err := s.getOwned(ctx, userID, id)
	if err != nil {
		return err
	}

	if err := s.storage.Delete(ctx, res.S3Key); err != nil {
		return fmt.Errorf("delete resume from storage: %w", err)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}

// getOwned fetches a resume and enforces that it belongs to userID.
func (s *ResumeService) getOwned(ctx context.Context, userID, id string) (*models.Resume, error) {
	res, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrResumeNotFound) {
			return nil, ErrResumeNotFound
		}
		return nil, err
	}
	if res.UserID != userID {
		return nil, ErrResumeAccessDenied
	}
	return res, nil
=======
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
		return nil, utils.NewStorageError(err, map[string]string{
			"operation": "upload_resume", "service": "s3", "user_id": input.UserID,
		})
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

		return nil, utils.NewDatabaseError(err, map[string]string{
			"operation": "create_resume", "resource": "resume", "user_id": input.UserID,
		})
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
		if errors.Is(err, ErrResumeNotFound) {
			return nil, ErrResumeNotFound
		}
		return nil, utils.NewDatabaseError(err, map[string]string{
			"operation": "get_resume", "resource": "resume", "resume_id": input.ResumeID,
		})
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
		return nil, utils.NewStorageError(err, map[string]string{
			"operation": "replace_resume", "service": "s3", "resume_id": input.ResumeID,
		})
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

		return nil, utils.NewDatabaseError(err, map[string]string{
			"operation": "update_resume", "resource": "resume", "resume_id": input.ResumeID,
		})
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

		return utils.NewDatabaseError(err, map[string]string{
			"operation": "get_resume", "resource": "resume", "resume_id": resumeID,
		})
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
			return utils.NewStorageError(err, map[string]string{
				"operation": "delete_resume", "service": "s3", "resume_id": resumeID,
			})
		}
	}

	if err := s.repository.Delete(ctx, resumeID); err != nil {
		return utils.NewDatabaseError(err, map[string]string{
			"operation": "delete_resume", "resource": "resume", "resume_id": resumeID,
		})
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
>>>>>>> dev
}
