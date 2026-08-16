package services

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"skill-match/backend/models"
	"skill-match/backend/repositories"
)

type fakeApplicationRepo struct {
	byID        map[string]*models.Application
	byUser      map[string][]*models.Application
	createErr   error
	getErr      error
	updateErr   error
	deleteErr   error
	seq         int
	updatedStatus map[string]models.ApplicationStatus
}

func newFakeApplicationRepo() *fakeApplicationRepo {
	return &fakeApplicationRepo{
		byID:          map[string]*models.Application{},
		byUser:        map[string][]*models.Application{},
		updatedStatus: map[string]models.ApplicationStatus{},
	}
}

func (f *fakeApplicationRepo) Create(_ context.Context, a *models.Application) (*models.Application, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.seq++
	a.ID = fmt.Sprintf("app-%d", f.seq)
	f.byID[a.ID] = a
	f.byUser[a.UserID] = append(f.byUser[a.UserID], a)
	return a, nil
}

func (f *fakeApplicationRepo) GetByID(_ context.Context, id string) (*models.Application, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if a, ok := f.byID[id]; ok {
		return a, nil
	}
	return nil, repositories.ErrApplicationNotFound
}

func (f *fakeApplicationRepo) ListByUserID(_ context.Context, userID string, _ int) ([]*models.Application, error) {
	return f.byUser[userID], nil
}

func (f *fakeApplicationRepo) UpdateStatus(_ context.Context, id string, status models.ApplicationStatus) (*models.Application, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	a, ok := f.byID[id]
	if !ok {
		return nil, repositories.ErrApplicationNotFound
	}
	a.Status = status
	f.updatedStatus[id] = status
	return a, nil
}

func (f *fakeApplicationRepo) Delete(_ context.Context, id string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	if _, ok := f.byID[id]; !ok {
		return repositories.ErrApplicationNotFound
	}
	delete(f.byID, id)
	return nil
}

func testApplicationService() (*ApplicationService, *fakeApplicationRepo) {
	repo := newFakeApplicationRepo()
	return NewApplicationService(repo), repo
}

func TestApplyCreatesApplication(t *testing.T) {
	svc, repo := testApplicationService()
	jobID := "job-1"

	created, err := svc.Apply(context.Background(), "user-1", &jobID, models.ApplicationStatusApplied)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if created.UserID != "user-1" || created.JobID == nil || *created.JobID != jobID {
		t.Fatalf("application fields not stored: %+v", created)
	}
	if created.Status != models.ApplicationStatusApplied {
		t.Fatalf("expected status applied, got %s", created.Status)
	}
	if len(repo.byUser["user-1"]) != 1 {
		t.Fatalf("expected 1 application, got %d", len(repo.byUser["user-1"]))
	}
}

func TestApplyAllowsNilJob(t *testing.T) {
	svc, _ := testApplicationService()

	created, err := svc.Apply(context.Background(), "user-1", nil, models.ApplicationStatusInterview)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if created.JobID != nil {
		t.Fatalf("expected nil job id, got %v", *created.JobID)
	}
}

func TestApplyRejectsInvalidInput(t *testing.T) {
	svc, repo := testApplicationService()

	if _, err := svc.Apply(context.Background(), "", nil, models.ApplicationStatusApplied); !errors.Is(err, ErrInvalidApplication) {
		t.Fatalf("expected ErrInvalidApplication for empty user, got %v", err)
	}
	if _, err := svc.Apply(context.Background(), "user-1", nil, models.ApplicationStatus("ghosted")); !errors.Is(err, ErrInvalidApplication) {
		t.Fatalf("expected ErrInvalidApplication for invalid status, got %v", err)
	}
	if len(repo.byUser) != 0 {
		t.Fatal("no application should be created for invalid input")
	}
}

func TestApplicationListScopesByUser(t *testing.T) {
	svc, _ := testApplicationService()
	ctx := context.Background()

	if _, err := svc.Apply(ctx, "user-1", nil, models.ApplicationStatusApplied); err != nil {
		t.Fatalf("apply user-1: %v", err)
	}
	if _, err := svc.Apply(ctx, "user-2", nil, models.ApplicationStatusApplied); err != nil {
		t.Fatalf("apply user-2: %v", err)
	}

	list, err := svc.List(ctx, "user-1", 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].UserID != "user-1" {
		t.Fatalf("expected only user-1's applications, got %+v", list)
	}
}

func TestUpdateStatusEnforcesOwnershipAndApplies(t *testing.T) {
	svc, repo := testApplicationService()
	ctx := context.Background()

	created, err := svc.Apply(ctx, "user-1", nil, models.ApplicationStatusApplied)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if _, err := svc.UpdateStatus(ctx, "user-2", created.ID, models.ApplicationStatusInterview); !errors.Is(err, ErrApplicationAccessDenied) {
		t.Fatalf("expected ErrApplicationAccessDenied, got %v", err)
	}

	updated, err := svc.UpdateStatus(ctx, "user-1", created.ID, models.ApplicationStatusInterview)
	if err != nil {
		t.Fatalf("update own: %v", err)
	}
	if updated.Status != models.ApplicationStatusInterview {
		t.Fatalf("expected status interview, got %s", updated.Status)
	}
	if repo.updatedStatus[created.ID] != models.ApplicationStatusInterview {
		t.Fatalf("repo update not recorded: %v", repo.updatedStatus)
	}

	if _, err := svc.UpdateStatus(ctx, "user-1", created.ID, models.ApplicationStatus("ghosted")); !errors.Is(err, ErrInvalidApplication) {
		t.Fatalf("expected ErrInvalidApplication for invalid status, got %v", err)
	}
}

func TestApplicationDeleteEnforcesOwnership(t *testing.T) {
	svc, _ := testApplicationService()
	ctx := context.Background()

	created, err := svc.Apply(ctx, "user-1", nil, models.ApplicationStatusApplied)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if err := svc.Delete(ctx, "user-2", created.ID); !errors.Is(err, ErrApplicationAccessDenied) {
		t.Fatalf("expected ErrApplicationAccessDenied, got %v", err)
	}
	if err := svc.Delete(ctx, "user-1", created.ID); err != nil {
		t.Fatalf("delete own: %v", err)
	}
	if err := svc.Delete(ctx, "user-1", created.ID); !errors.Is(err, ErrApplicationNotFound) {
		t.Fatalf("expected ErrApplicationNotFound after delete, got %v", err)
	}
}
