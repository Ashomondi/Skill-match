package services

import (
	"context"
	"errors"
	"testing"

	"skill-match/backend/models"
)

type fakeApplicationRepo struct {
	apps       []*models.Application
	byUser     map[string][]*models.Application
	createErr  error
	listErr    error
	getErr     error
	updateErr  error
	historyErr error
	created    []*models.Application
	updatedTo  []models.ApplicationStatus
}

func newFakeApplicationRepo() *fakeApplicationRepo {
	return &fakeApplicationRepo{byUser: map[string][]*models.Application{}}
}

func (f *fakeApplicationRepo) Create(_ context.Context, userID, jobID string) (*models.Application, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	a := &models.Application{ID: "app-" + string(rune('a'+len(f.created))), UserID: userID, JobID: jobID, Status: models.ApplicationSaved}
	f.created = append(f.created, a)
	f.byUser[userID] = append(f.byUser[userID], a)
	return a, nil
}

func (f *fakeApplicationRepo) GetByID(_ context.Context, userID, id string) (*models.Application, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	for _, a := range f.byUser[userID] {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, ErrApplicationNotFound
}

func (f *fakeApplicationRepo) UpdateStatus(_ context.Context, userID, id string, status models.ApplicationStatus) (*models.Application, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	a, err := f.GetByID(context.Background(), userID, id)
	if err != nil {
		return nil, err
	}
	a.Status = status
	f.updatedTo = append(f.updatedTo, status)
	return a, nil
}

func (f *fakeApplicationRepo) History(_ context.Context, _, _ string) ([]models.ApplicationStatusChange, error) {
	if f.historyErr != nil {
		return nil, f.historyErr
	}
	return nil, nil
}

func (f *fakeApplicationRepo) ListByUserID(_ context.Context, userID string) ([]*models.Application, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.byUser[userID], nil
}

func testApplicationService(repo *fakeApplicationRepo) *ApplicationService {
	return NewApplicationService(repo)
}

func TestApplicationListScopesByUser(t *testing.T) {
	repo := newFakeApplicationRepo()
	repo.byUser["user-1"] = []*models.Application{{ID: "a1", UserID: "user-1", JobID: "job-1", Status: models.ApplicationSaved}}
	repo.byUser["user-2"] = []*models.Application{{ID: "a2", UserID: "user-2", JobID: "job-2", Status: models.ApplicationSaved}}
	svc := testApplicationService(repo)

	list, err := svc.List(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].UserID != "user-1" || list[0].JobID != "job-1" {
		t.Fatalf("expected only user-1's application, got %+v", list)
	}
}

func TestApplicationListRejectsEmptyUser(t *testing.T) {
	svc := testApplicationService(newFakeApplicationRepo())

	if _, err := svc.List(context.Background(), "  "); !errors.Is(err, ErrApplicationInvalidInput) {
		t.Fatalf("expected ErrApplicationInvalidInput, got %v", err)
	}
}

func TestApplicationListReturnsJobDetails(t *testing.T) {
	repo := newFakeApplicationRepo()
	repo.byUser["user-1"] = []*models.Application{
		{ID: "a1", UserID: "user-1", JobID: "job-9", Status: models.ApplicationApplied,
			Job: &models.Job{ID: "job-9", Title: "Backend Engineer", Company: "Acme"}},
	}
	svc := testApplicationService(repo)

	list, err := svc.List(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].Job == nil || list[0].Job.Title != "Backend Engineer" || list[0].Job.Company != "Acme" {
		t.Fatalf("expected job details on the application, got %+v", list[0])
	}
}

func TestApplicationListSurfacesRepoError(t *testing.T) {
	repo := newFakeApplicationRepo()
	repo.listErr = errors.New("db down")
	svc := testApplicationService(repo)

	if _, err := svc.List(context.Background(), "user-1"); err == nil {
		t.Fatal("expected an error when the repository fails")
	}
}
