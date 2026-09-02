package services

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"skill-match/backend/models"
	"skill-match/backend/repositories"
	"skill-match/backend/utils"
)

var (
	ErrResumeNotFound     = errors.New("resume not found")
	ErrResumeAccessDenied = errors.New("resume access denied")
	ErrInvalidResume      = errors.New("invalid resume")
	ErrStorageUnavailable = errors.New("object storage is not configured")
	ErrResumeUnauthorized = ErrResumeAccessDenied // alias used by AI/recommendation services
)

// ObjectStorage is the subset of the S3 client the resume service needs.
// clients.S3Client satisfies this interface.
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
// (PostgreSQL). It enforces user ownership on every operation.
type ResumeService struct {
	repo    ResumeRepository
	storage ObjectStorage
}

func NewResumeService(repo ResumeRepository, storage ObjectStorage) *ResumeService {
	return &ResumeService{repo: repo, storage: storage}
}

// storageAvailable reports whether object storage is configured. It also
// guards against a typed-nil storage value (e.g. a nil *clients.S3Client
// boxed into the ObjectStorage interface), which a plain `== nil` check
// misses.
func storageAvailable(storage ObjectStorage) bool {
	if storage == nil {
		return false
	}
	v := reflect.ValueOf(storage)
	return !(v.Kind() == reflect.Ptr && v.IsNil())
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

	if err := utils.ValidateResumeFile(filename, contentType, int64(len(data)), data); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResume, err)
	}

	if replaceID != "" {
		if _, err := s.getOwned(ctx, userID, replaceID); err != nil {
			return nil, err
		}
	}

	if !storageAvailable(s.storage) {
		return nil, ErrStorageUnavailable
	}

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

	if !storageAvailable(s.storage) {
		return nil, "", ErrStorageUnavailable
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

	if !storageAvailable(s.storage) {
		return ErrStorageUnavailable
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
}
