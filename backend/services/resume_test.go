package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"skill-match/backend/models"
	"skill-match/backend/repositories"
)

// fakeStorage is an in-memory ObjectStorage.
type fakeStorage struct {
	objects    map[string][]byte
	deleted    []string
	putErr     error
	deleteErr  error
	presignErr error
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{objects: map[string][]byte{}}
}

func (f *fakeStorage) Put(_ context.Context, key string, body []byte, _ string) error {
	if f.putErr != nil {
		return f.putErr
	}
	f.objects[key] = body
	return nil
}

func (f *fakeStorage) PresignDownload(_ context.Context, key string, _ time.Duration) (string, error) {
	if f.presignErr != nil {
		return "", f.presignErr
	}
	return "https://presigned/" + key, nil
}

func (f *fakeStorage) Delete(_ context.Context, key string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deleted = append(f.deleted, key)
	delete(f.objects, key)
	return nil
}

func (f *fakeStorage) Key(userID, fileID string) string {
	return "resumes/" + userID + "/" + fileID
}

// fakeResumeRepo is an in-memory ResumeRepository.
type fakeResumeRepo struct {
	byID      map[string]*models.Resume
	byUser    map[string][]*models.Resume
	createErr error
	getErr    error
	deleteErr error
	seq       int
}

func newFakeResumeRepo() *fakeResumeRepo {
	return &fakeResumeRepo{
		byID:   map[string]*models.Resume{},
		byUser: map[string][]*models.Resume{},
	}
}

func (f *fakeResumeRepo) Create(_ context.Context, r *models.Resume) (*models.Resume, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.seq++
	r.ID = fmt.Sprintf("res-%d", f.seq)
	f.byID[r.ID] = r
	f.byUser[r.UserID] = append(f.byUser[r.UserID], r)
	return r, nil
}

func (f *fakeResumeRepo) GetByID(_ context.Context, id string) (*models.Resume, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if r, ok := f.byID[id]; ok {
		return r, nil
	}
	return nil, repositories.ErrResumeNotFound
}

func (f *fakeResumeRepo) ListByUserID(_ context.Context, userID string, _ int) ([]*models.Resume, error) {
	return f.byUser[userID], nil
}

func (f *fakeResumeRepo) Delete(_ context.Context, id string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	if _, ok := f.byID[id]; !ok {
		return repositories.ErrResumeNotFound
	}
	delete(f.byID, id)
	return nil
}

const (
	testUserID = "user-1"
	pdfBody    = "%PDF-1.4 fake resume content"
	pdfCT      = "application/pdf"
)

func testResumeService() (*ResumeService, *fakeStorage, *fakeResumeRepo) {
	storage := newFakeStorage()
	repo := newFakeResumeRepo()
	return NewResumeService(repo, storage), storage, repo
}

func TestUploadStoresObjectAndRow(t *testing.T) {
	svc, storage, repo := testResumeService()

	res, err := svc.Upload(context.Background(), testUserID, "", "resume.pdf", pdfCT, []byte(pdfBody))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	if res.UserID != testUserID {
		t.Errorf("expected user id %s, got %s", testUserID, res.UserID)
	}
	if res.Status != models.ResumeStatusUploaded {
		t.Errorf("expected status uploaded, got %s", res.Status)
	}
	if res.S3Key == "" {
		t.Error("expected s3 key to be set")
	}
	if len(storage.objects[res.S3Key]) != len(pdfBody) {
		t.Error("expected object bytes to be stored under the resume s3 key")
	}
	if len(repo.byID) != 1 {
		t.Errorf("expected 1 row, got %d", len(repo.byID))
	}
}

func TestUploadRejectsInvalidFile(t *testing.T) {
	svc, storage, repo := testResumeService()

	_, err := svc.Upload(context.Background(), testUserID, "", "resume.exe", "application/octet-stream", []byte("x"))
	if !errors.Is(err, ErrInvalidResume) {
		t.Fatalf("expected ErrInvalidResume, got %v", err)
	}
	if len(storage.objects) != 0 {
		t.Error("no object should be stored for an invalid file")
	}
	if len(repo.byID) != 0 {
		t.Error("no row should be created for an invalid file")
	}
}

func TestUploadRollsBackObjectWhenCreateFails(t *testing.T) {
	svc, storage, repo := testResumeService()
	repo.createErr = errors.New("db down")

	_, err := svc.Upload(context.Background(), testUserID, "", "resume.pdf", pdfCT, []byte(pdfBody))
	if err == nil {
		t.Fatal("expected error when repo.Create fails")
	}
	if len(storage.objects) != 0 {
		t.Error("expected the s3 object to be rolled back after a failed insert")
	}
	if len(repo.byID) != 0 {
		t.Error("no row should exist after a failed create")
	}
}

func TestUploadReplaceRemovesOldResume(t *testing.T) {
	svc, storage, repo := testResumeService()

	old, err := svc.Upload(context.Background(), testUserID, "", "old.pdf", pdfCT, []byte("%PDF-1.4 old"))
	if err != nil {
		t.Fatalf("upload old: %v", err)
	}

	newRes, err := svc.Upload(context.Background(), testUserID, old.ID, "new.pdf", pdfCT, []byte("%PDF-1.4 new"))
	if err != nil {
		t.Fatalf("upload new: %v", err)
	}

	if newRes.ID == old.ID {
		t.Error("expected a new resume row")
	}
	if _, ok := storage.objects[old.S3Key]; ok {
		t.Error("expected old s3 object to be deleted")
	}
	if _, err := repo.GetByID(context.Background(), old.ID); !errors.Is(err, repositories.ErrResumeNotFound) {
		t.Error("expected old resume row to be removed")
	}
}

func TestUploadRejectsReplacingOthersResume(t *testing.T) {
	svc, storage, _ := testResumeService()

	other, err := svc.Upload(context.Background(), "user-other", "", "other.pdf", pdfCT, []byte("%PDF-1.4 other"))
	if err != nil {
		t.Fatalf("upload other: %v", err)
	}

	before := len(storage.objects)
	_, err = svc.Upload(context.Background(), testUserID, other.ID, "mine.pdf", pdfCT, []byte("%PDF-1.4 mine"))
	if !errors.Is(err, ErrResumeAccessDenied) {
		t.Fatalf("expected ErrResumeAccessDenied, got %v", err)
	}
	if len(storage.objects) != before {
		t.Error("no object should be stored when replace is denied")
	}
}

func TestDeleteRemovesObjectAndRow(t *testing.T) {
	svc, storage, repo := testResumeService()

	res, err := svc.Upload(context.Background(), testUserID, "", "resume.pdf", pdfCT, []byte(pdfBody))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	if err := svc.Delete(context.Background(), testUserID, res.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(storage.objects) != 0 {
		t.Error("expected s3 object to be deleted")
	}
	if _, err := repo.GetByID(context.Background(), res.ID); !errors.Is(err, repositories.ErrResumeNotFound) {
		t.Error("expected row to be deleted")
	}
}

func TestDeleteDeniesOtherUsersResume(t *testing.T) {
	svc, storage, _ := testResumeService()

	res, err := svc.Upload(context.Background(), "user-other", "", "other.pdf", pdfCT, []byte("%PDF-1.4 other"))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	err = svc.Delete(context.Background(), testUserID, res.ID)
	if !errors.Is(err, ErrResumeAccessDenied) {
		t.Fatalf("expected ErrResumeAccessDenied, got %v", err)
	}
	if len(storage.objects) == 0 {
		t.Error("object should not be deleted for a denied request")
	}
}

func TestDownloadURLPresignsOwnedResume(t *testing.T) {
	svc, _, _ := testResumeService()

	res, err := svc.Upload(context.Background(), testUserID, "", "resume.pdf", pdfCT, []byte(pdfBody))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	_, url, err := svc.DownloadURL(context.Background(), testUserID, res.ID, time.Minute)
	if err != nil {
		t.Fatalf("download url: %v", err)
	}
	if url == "" {
		t.Error("expected a presigned url")
	}
}

func TestListScopesByUser(t *testing.T) {
	svc, _, _ := testResumeService()

	_, err := svc.Upload(context.Background(), testUserID, "", "a.pdf", pdfCT, []byte("%PDF-1.4 a"))
	if err != nil {
		t.Fatalf("upload a: %v", err)
	}
	_, err = svc.Upload(context.Background(), "user-other", "", "b.pdf", pdfCT, []byte("%PDF-1.4 b"))
	if err != nil {
		t.Fatalf("upload b: %v", err)
	}

	list, err := svc.List(context.Background(), testUserID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].UserID != testUserID {
		t.Errorf("expected only user-1 resumes, got %d", len(list))
	}
}

func TestUploadWithoutStorageFailsCleanly(t *testing.T) {
	svc := NewResumeService(newFakeResumeRepo(), nil)

	_, err := svc.Upload(context.Background(), testUserID, "", "resume.pdf", pdfCT, []byte(pdfBody))
	if !errors.Is(err, ErrStorageUnavailable) {
		t.Fatalf("expected ErrStorageUnavailable, got %v", err)
	}
}

func TestUploadWithTypedNilStorageFailsCleanly(t *testing.T) {
	// Reproduces production wiring: a nil *clients.S3Client boxed into the
	// ObjectStorage interface is not a nil interface; the guard must catch it.
	var storage *fakeStorage
	svc := NewResumeService(newFakeResumeRepo(), storage)

	_, err := svc.Upload(context.Background(), testUserID, "", "resume.pdf", pdfCT, []byte(pdfBody))
	if !errors.Is(err, ErrStorageUnavailable) {
		t.Fatalf("expected ErrStorageUnavailable, got %v", err)
	}
}

func TestDownloadURLWithoutStorageFailsCleanly(t *testing.T) {
	repo := newFakeResumeRepo()
	repo.byID["res-1"] = &models.Resume{ID: "res-1", UserID: testUserID, S3Key: "resumes/user-1/x"}
	repo.byUser[testUserID] = []*models.Resume{repo.byID["res-1"]}
	svc := NewResumeService(repo, nil)

	_, _, err := svc.DownloadURL(context.Background(), testUserID, "res-1", time.Minute)
	if !errors.Is(err, ErrStorageUnavailable) {
		t.Fatalf("expected ErrStorageUnavailable, got %v", err)
	}
}

func TestDeleteWithoutStorageFailsCleanly(t *testing.T) {
	repo := newFakeResumeRepo()
	repo.byID["res-1"] = &models.Resume{ID: "res-1", UserID: testUserID, S3Key: "resumes/user-1/x"}
	repo.byUser[testUserID] = []*models.Resume{repo.byID["res-1"]}
	svc := NewResumeService(repo, nil)

	err := svc.Delete(context.Background(), testUserID, "res-1")
	if !errors.Is(err, ErrStorageUnavailable) {
		t.Fatalf("expected ErrStorageUnavailable, got %v", err)
	}
}
