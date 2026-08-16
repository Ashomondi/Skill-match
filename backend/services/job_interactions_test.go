package services

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"skill-match/backend/models"
	"skill-match/backend/repositories"
)

type fakeJobInteractionRepo struct {
	byID      map[string]*models.JobInteraction
	byUser    map[string][]*models.JobInteraction
	createErr error
	getErr    error
	deleteErr error
	seq       int
}

func newFakeJobInteractionRepo() *fakeJobInteractionRepo {
	return &fakeJobInteractionRepo{
		byID:   map[string]*models.JobInteraction{},
		byUser: map[string][]*models.JobInteraction{},
	}
}

func (f *fakeJobInteractionRepo) Create(_ context.Context, in *models.JobInteraction) (*models.JobInteraction, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.seq++
	in.ID = fmt.Sprintf("ji-%d", f.seq)
	f.byID[in.ID] = in
	f.byUser[in.UserID] = append(f.byUser[in.UserID], in)
	return in, nil
}

func (f *fakeJobInteractionRepo) GetByID(_ context.Context, id string) (*models.JobInteraction, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if in, ok := f.byID[id]; ok {
		return in, nil
	}
	return nil, repositories.ErrJobInteractionNotFound
}

func (f *fakeJobInteractionRepo) ListByUserID(_ context.Context, userID string, _ int) ([]*models.JobInteraction, error) {
	return f.byUser[userID], nil
}

func (f *fakeJobInteractionRepo) Delete(_ context.Context, id string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	if _, ok := f.byID[id]; !ok {
		return repositories.ErrJobInteractionNotFound
	}
	delete(f.byID, id)
	return nil
}

func testJobInteractionService() (*JobInteractionService, *fakeJobInteractionRepo) {
	repo := newFakeJobInteractionRepo()
	return NewJobInteractionService(repo), repo
}

func TestRecordStoresInteraction(t *testing.T) {
	svc, repo := testJobInteractionService()

	created, err := svc.Record(context.Background(), "user-1", "job-1", models.InteractionView)
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	if created.UserID != "user-1" || created.JobID != "job-1" || created.Type != models.InteractionView {
		t.Fatalf("interaction fields not stored: %+v", created)
	}
	if len(repo.byUser["user-1"]) != 1 {
		t.Fatalf("expected 1 interaction, got %d", len(repo.byUser["user-1"]))
	}
}

func TestRecordRejectsMissingUserOrJob(t *testing.T) {
	svc, repo := testJobInteractionService()

	if _, err := svc.Record(context.Background(), "", "job-1", models.InteractionView); !errors.Is(err, ErrInvalidJobInteraction) {
		t.Fatalf("expected ErrInvalidJobInteraction for empty user, got %v", err)
	}
	if _, err := svc.Record(context.Background(), "user-1", "", models.InteractionView); !errors.Is(err, ErrInvalidJobInteraction) {
		t.Fatalf("expected ErrInvalidJobInteraction for empty job, got %v", err)
	}
	if len(repo.byUser) != 0 {
		t.Fatal("no interaction should be recorded for invalid input")
	}
}

func TestRecordRejectsInvalidType(t *testing.T) {
	svc, _ := testJobInteractionService()

	if _, err := svc.Record(context.Background(), "user-1", "job-1", models.InteractionType("click")); !errors.Is(err, ErrInvalidJobInteraction) {
		t.Fatalf("expected ErrInvalidJobInteraction for invalid type, got %v", err)
	}
}

func TestJobInteractionListScopesByUser(t *testing.T) {
	svc, _ := testJobInteractionService()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if _, err := svc.Record(ctx, "user-1", fmt.Sprintf("job-%d", i), models.InteractionView); err != nil {
			t.Fatalf("record for user-1: %v", err)
		}
	}
	if _, err := svc.Record(ctx, "user-2", "job-x", models.InteractionSave); err != nil {
		t.Fatalf("record for user-2: %v", err)
	}

	list, err := svc.List(ctx, "user-1", 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 interactions for user-1, got %d", len(list))
	}
	for _, in := range list {
		if in.UserID != "user-1" {
			t.Fatalf("list leaked another user's interaction: %+v", in)
		}
	}
}

func TestDeleteEnforcesOwnership(t *testing.T) {
	svc, _ := testJobInteractionService()
	ctx := context.Background()

	created, err := svc.Record(ctx, "user-1", "job-1", models.InteractionView)
	if err != nil {
		t.Fatalf("record: %v", err)
	}

	if err := svc.Delete(ctx, "user-2", created.ID); !errors.Is(err, ErrJobInteractionAccessDenied) {
		t.Fatalf("expected ErrJobInteractionAccessDenied, got %v", err)
	}

	if err := svc.Delete(ctx, "user-1", created.ID); err != nil {
		t.Fatalf("delete own: %v", err)
	}

	if err := svc.Delete(ctx, "user-1", created.ID); !errors.Is(err, ErrJobInteractionNotFound) {
		t.Fatalf("expected ErrJobInteractionNotFound after delete, got %v", err)
	}
}
