//go:build integration

package repositories

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"skill-match/backend/models"
)

// Issue #41 — conversation persistence workflow.
//
// Build-tagged and env-guarded so it never runs in a normal `go test ./...`
// or CI without infrastructure. Provide TEST_DATABASE_URL (a PostgreSQL dsn)
// and run with:
//
//	go test -tags integration ./repositories -run TestConversation -v

func connectTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect to postgres: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func createTestUser(t *testing.T, pool *pgxpool.Pool) *User {
	t.Helper()
	userRepo := NewUserRepository(pool)
	user, err := userRepo.Create(context.Background(), &User{
		Email:        fmt.Sprintf("it-%d@skillmatch.local", time.Now().UnixNano()),
		PasswordHash: "integration-placeholder-hash",
		FullName:     "Integration Tester",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() { _ = userRepo.Delete(context.Background(), user.ID) })
	return user
}

func TestConversationPersistence(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	repo := NewConversationRepository(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, &models.Conversation{
		UserID:  user.ID,
		Role:    models.ConversationRoleUser,
		Content: "I prefer remote backend roles",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == "" || created.UserID != user.ID {
		t.Fatalf("expected persisted turn with id and user, got %+v", created)
	}

	history, err := repo.ListRecentByUserID(ctx, user.ID, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(history) != 1 || history[0].Content != "I prefer remote backend roles" {
		t.Fatalf("expected the stored turn to be retrievable, got %+v", history)
	}
}

func TestConversationBatchInsert(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	repo := NewConversationRepository(pool)
	ctx := context.Background()

	turns := []*models.Conversation{
		{UserID: user.ID, Role: models.ConversationRoleUser, Content: "first"},
		{UserID: user.ID, Role: models.ConversationRoleAssistant, Content: "second"},
		{UserID: user.ID, Role: models.ConversationRoleUser, Content: "third"},
	}
	inserted, err := repo.CreateBatch(ctx, turns)
	if err != nil {
		t.Fatalf("batch insert: %v", err)
	}
	if len(inserted) != 3 {
		t.Fatalf("expected 3 inserted turns, got %d", len(inserted))
	}

	history, err := repo.ListRecentByUserID(ctx, user.ID, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	got := map[string]bool{}
	for _, c := range history {
		got[c.Content] = true
	}
	for _, want := range []string{"first", "second", "third"} {
		if !got[want] {
			t.Errorf("missing turn %q in history %v", want, history)
		}
	}
}

func TestConversationChronologicalRetrieval(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	repo := NewConversationRepository(pool)
	ctx := context.Background()

	for _, content := range []string{"first", "second", "third"} {
		if _, err := repo.Create(ctx, &models.Conversation{
			UserID:  user.ID,
			Role:    models.ConversationRoleUser,
			Content: content,
		}); err != nil {
			t.Fatalf("create %q: %v", content, err)
		}
		time.Sleep(15 * time.Millisecond) // distinct created_at timestamps
	}

	history, err := repo.ListRecentByUserID(ctx, user.ID, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	got := make([]string, len(history))
	for i, c := range history {
		got[i] = c.Content
	}
	want := []string{"first", "second", "third"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("expected chronological order %v, got %v", want, got)
	}
}

func TestConversationUserIsolation(t *testing.T) {
	pool := connectTestPool(t)
	userA := createTestUser(t, pool)
	userB := createTestUser(t, pool)
	repo := NewConversationRepository(pool)
	ctx := context.Background()

	for _, content := range []string{"a-one", "a-two"} {
		if _, err := repo.Create(ctx, &models.Conversation{UserID: userA.ID, Role: models.ConversationRoleUser, Content: content}); err != nil {
			t.Fatalf("create for A: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if _, err := repo.Create(ctx, &models.Conversation{UserID: userB.ID, Role: models.ConversationRoleUser, Content: "b-one"}); err != nil {
		t.Fatalf("create for B: %v", err)
	}

	historyA, err := repo.ListRecentByUserID(ctx, userA.ID, 0)
	if err != nil {
		t.Fatalf("list A: %v", err)
	}
	if len(historyA) != 2 {
		t.Fatalf("expected 2 turns for user A, got %d", len(historyA))
	}
	for _, c := range historyA {
		if c.UserID != userA.ID {
			t.Errorf("user A history leaked another user's turn: %+v", c)
		}
	}

	historyB, err := repo.ListRecentByUserID(ctx, userB.ID, 0)
	if err != nil {
		t.Fatalf("list B: %v", err)
	}
	if len(historyB) != 1 || historyB[0].Content != "b-one" {
		t.Fatalf("expected only B's turn, got %+v", historyB)
	}
}

func TestConversationHistoryLimit(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	repo := NewConversationRepository(pool)
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		content := fmt.Sprintf("turn-%d", i)
		if _, err := repo.Create(ctx, &models.Conversation{UserID: user.ID, Role: models.ConversationRoleUser, Content: content}); err != nil {
			t.Fatalf("create %q: %v", content, err)
		}
		time.Sleep(15 * time.Millisecond)
	}

	history, err := repo.ListRecentByUserID(ctx, user.ID, 2)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected 2 most recent turns, got %d", len(history))
	}
	if history[0].Content != "turn-2" || history[1].Content != "turn-3" {
		t.Fatalf("expected the two most recent turns, got %v", history)
	}
}

func TestConversationDatabaseFailure(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	pool.Close() // simulate an unavailable database

	repo := NewConversationRepository(pool)
	ctx := context.Background()

	if _, err := repo.Create(ctx, &models.Conversation{
		UserID:  "user-x",
		Role:    models.ConversationRoleUser,
		Content: "should fail",
	}); err == nil {
		t.Fatal("expected an error when the database is unavailable")
	}

	if _, err := repo.ListRecentByUserID(ctx, "user-x", 0); err == nil {
		t.Fatal("expected an error when the database is unavailable")
	}
}
