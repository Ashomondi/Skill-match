//go:build integration

package repositories

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Issue #72 — application list persistence. Fake-free: runs against a real
// CockroachDB. Requires TEST_DATABASE_URL and the migrations applied.
//
//	go test -tags integration ./repositories -run TestApplication -v

// insertTestApplicationJob seeds a minimal jobs row and returns its id,
// cleaning up afterwards (CASCADE removes any applications referencing it).
func insertTestApplicationJob(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO jobs (external_id, title, company) VALUES ($1, 'Integration Role', 'Acme Corp') RETURNING id`,
		"integration-job-"+time.Now().Format("150405"),
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert job: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM jobs WHERE id = $1`, id)
	})
	return id
}

func TestApplicationListByUserID(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	jobID := insertTestApplicationJob(t, pool)
	repo := NewApplicationRepository(pool)
	ctx := context.Background()

	if _, err := repo.Create(ctx, user.ID, jobID); err != nil {
		t.Fatalf("create application: %v", err)
	}

	list, err := repo.ListByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 application, got %d", len(list))
	}
	if list[0].UserID != user.ID || list[0].JobID != jobID {
		t.Fatalf("unexpected application: %+v", list[0])
	}
	if list[0].Job == nil || list[0].Job.Title != "Integration Role" || list[0].Job.Company != "Acme Corp" {
		t.Fatalf("expected job details to be joined, got %+v", list[0].Job)
	}
}

func TestApplicationListUserIsolation(t *testing.T) {
	pool := connectTestPool(t)
	userA := createTestUser(t, pool)
	userB := createTestUser(t, pool)
	jobID := insertTestApplicationJob(t, pool)
	repo := NewApplicationRepository(pool)
	ctx := context.Background()

	if _, err := repo.Create(ctx, userA.ID, jobID); err != nil {
		t.Fatalf("create for A: %v", err)
	}
	if _, err := repo.Create(ctx, userB.ID, jobID); err != nil {
		t.Fatalf("create for B: %v", err)
	}

	listA, err := repo.ListByUserID(ctx, userA.ID)
	if err != nil {
		t.Fatalf("list A: %v", err)
	}
	if len(listA) != 1 || listA[0].UserID != userA.ID {
		t.Fatalf("expected only user A's application, got %+v", listA)
	}
}

func TestApplicationListEmpty(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	repo := NewApplicationRepository(pool)

	list, err := repo.ListByUserID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected no applications, got %d", len(list))
	}
}

func TestApplicationListDatabaseFailure(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	pool.Close() // simulate an unavailable database

	repo := NewApplicationRepository(pool)
	if _, err := repo.ListByUserID(context.Background(), "user-x"); err == nil {
		t.Fatal("expected an error when the database is unavailable")
	}
}
