//go:build integration

package repositories

import (
	"context"
	"errors"
	"os"
	"testing"

	"skill-match/backend/clients"
	"skill-match/backend/models"
)

// Issue #66 — application persistence. Fake-free: runs against a real
// CockroachDB. Requires TEST_DATABASE_URL and the migrations applied.
//
//	go test -tags integration ./repositories -run TestApplication -v

func TestApplicationCreateAndRetrieve(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	jobID := insertTestJob(t, pool)
	repo := NewApplicationRepository(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, &models.Application{
		UserID: user.ID,
		JobID:  &jobID,
		Status: models.ApplicationStatusApplied,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == "" || created.UserID != user.ID || created.JobID == nil || *created.JobID != jobID {
		t.Fatalf("application fields not persisted: %+v", created)
	}
	if created.AppliedAt.IsZero() {
		t.Fatal("expected applied_at to be set")
	}

	fetched, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if fetched.Status != models.ApplicationStatusApplied {
		t.Fatalf("expected status applied, got %s", fetched.Status)
	}
}

func TestApplicationAllowsNilJob(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	repo := NewApplicationRepository(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, &models.Application{
		UserID: user.ID,
		Status: models.ApplicationStatusInterview,
	})
	if err != nil {
		t.Fatalf("create without job: %v", err)
	}
	if created.JobID != nil {
		t.Fatalf("expected nil job id, got %v", *created.JobID)
	}
}

func TestApplicationUserIsolation(t *testing.T) {
	pool := connectTestPool(t)
	userA := createTestUser(t, pool)
	userB := createTestUser(t, pool)
	repo := NewApplicationRepository(pool)
	ctx := context.Background()

	for _, u := range []*models.User{userA, userB} {
		if _, err := repo.Create(ctx, &models.Application{UserID: u.ID, Status: models.ApplicationStatusApplied}); err != nil {
			t.Fatalf("create for %s: %v", u.ID, err)
		}
	}

	listA, err := repo.ListByUserID(ctx, userA.ID, 0)
	if err != nil {
		t.Fatalf("list A: %v", err)
	}
	if len(listA) != 1 || listA[0].UserID != userA.ID {
		t.Fatalf("expected only user A's application, got %+v", listA)
	}
}

func TestApplicationUpdateStatus(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	repo := NewApplicationRepository(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, &models.Application{UserID: user.ID, Status: models.ApplicationStatusApplied})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	updated, err := repo.UpdateStatus(ctx, created.ID, models.ApplicationStatusOffer)
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if updated.Status != models.ApplicationStatusOffer {
		t.Fatalf("expected status offer, got %s", updated.Status)
	}
	if !updated.UpdatedAt.After(updated.CreatedAt) {
		t.Fatalf("expected updated_at to move forward: %v -> %v", updated.CreatedAt, updated.UpdatedAt)
	}
}

func TestApplicationDelete(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	repo := NewApplicationRepository(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, &models.Application{UserID: user.ID, Status: models.ApplicationStatusApplied})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, created.ID); !errors.Is(err, ErrApplicationNotFound) {
		t.Fatalf("expected ErrApplicationNotFound after delete, got %v", err)
	}
}

func TestApplicationDatabaseFailure(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	pool, err := clients.NewPool(context.Background(), dsn, clients.PoolOptions{})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	pool.Close() // simulate an unavailable database

	repo := NewApplicationRepository(pool)
	ctx := context.Background()

	if _, err := repo.Create(ctx, &models.Application{UserID: "user-x", Status: models.ApplicationStatusApplied}); err == nil {
		t.Fatal("expected an error when the database is unavailable")
	}
	if _, err := repo.ListByUserID(ctx, "user-x", 0); err == nil {
		t.Fatal("expected an error when the database is unavailable")
	}
}
