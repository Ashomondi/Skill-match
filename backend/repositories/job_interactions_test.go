//go:build integration

package repositories

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"skill-match/backend/clients"
	"skill-match/backend/models"
)

// Issue #56 — job interaction persistence. Fake-free: runs against a real
// CockroachDB. Requires TEST_DATABASE_URL and the migrations applied.
//
//	go test -tags integration ./repositories -run TestJobInteraction -v

// insertTestJob seeds a minimal jobs row (the job_interactions table has an
// FK to jobs) and returns its id, cleaning up afterwards.
func insertTestJob(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO jobs (title, company) VALUES ('Integration Job', 'Acme Corp') RETURNING id`,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert job: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM jobs WHERE id = $1`, id)
	})
	return id
}

func TestJobInteractionRecordAndRetrieve(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	jobID := insertTestJob(t, pool)
	repo := NewJobInteractionRepository(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, &models.JobInteraction{
		UserID: user.ID,
		JobID:  jobID,
		Type:   models.InteractionView,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == "" || created.UserID != user.ID || created.JobID != jobID {
		t.Fatalf("interaction fields not persisted: %+v", created)
	}

	list, err := repo.ListByUserID(ctx, user.ID, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].Type != models.InteractionView {
		t.Fatalf("expected the recorded interaction to be retrieved, got %+v", list)
	}
}

func TestJobInteractionUserIsolation(t *testing.T) {
	pool := connectTestPool(t)
	userA := createTestUser(t, pool)
	userB := createTestUser(t, pool)
	jobID := insertTestJob(t, pool)
	repo := NewJobInteractionRepository(pool)
	ctx := context.Background()

	for _, u := range []*models.User{userA, userB} {
		if _, err := repo.Create(ctx, &models.JobInteraction{
			UserID: u.ID, JobID: jobID, Type: models.InteractionSave,
		}); err != nil {
			t.Fatalf("create for %s: %v", u.ID, err)
		}
	}

	listA, err := repo.ListByUserID(ctx, userA.ID, 0)
	if err != nil {
		t.Fatalf("list A: %v", err)
	}
	if len(listA) != 1 || listA[0].UserID != userA.ID {
		t.Fatalf("expected only user A's interaction, got %+v", listA)
	}
}

func TestJobInteractionDelete(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	jobID := insertTestJob(t, pool)
	repo := NewJobInteractionRepository(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, &models.JobInteraction{
		UserID: user.ID, JobID: jobID, Type: models.InteractionView,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, created.ID); !errors.Is(err, ErrJobInteractionNotFound) {
		t.Fatalf("expected ErrJobInteractionNotFound after delete, got %v", err)
	}
}

func TestJobInteractionDatabaseFailure(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	pool, err := clients.NewPool(context.Background(), dsn, clients.PoolOptions{})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	pool.Close() // simulate an unavailable database

	repo := NewJobInteractionRepository(pool)
	ctx := context.Background()

	if _, err := repo.Create(ctx, &models.JobInteraction{
		UserID: "user-x", JobID: "job-x", Type: models.InteractionView,
	}); err == nil {
		t.Fatal("expected an error when the database is unavailable")
	}
	if _, err := repo.ListByUserID(ctx, "user-x", 0); err == nil {
		t.Fatal("expected an error when the database is unavailable")
	}
}
