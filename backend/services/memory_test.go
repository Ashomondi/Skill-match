//go:build integration

package services

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"skill-match/backend/models"
	"skill-match/backend/repositories"
)

// Issue #41 — persistent memory workflow: information stored during one
// conversation can be recalled and used during a later conversation.
//
// Build-tagged and env-guarded. Provide TEST_DATABASE_URL (a CockroachDB dsn)
// and run with:
//
//	go test -tags integration ./services -run TestMemory -v

func memoryTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect to cockroachdb: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func memoryTestUser(t *testing.T, pool *pgxpool.Pool) *repositories.User {
	t.Helper()
	userRepo := repositories.NewUserRepository(pool)
	user, err := userRepo.Create(context.Background(), &repositories.User{
		Email:        fmt.Sprintf("mem-%d@skillmatch.local", time.Now().UnixNano()),
		PasswordHash: "integration-placeholder-hash",
		FullName:     "Memory Tester",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() { _ = userRepo.Delete(context.Background(), user.ID) })
	return user
}

// embeddingVector builds a 1024-dim vector (the fixed Titan dimension) whose
// value is dominated by the seed, so two vectors with similar seeds have a
// tiny cosine distance and are reliably recalled as nearest neighbors.
func embeddingVector(seed float32) []float32 {
	vec := make([]float32, repositories.EmbeddingDim)
	vec[0] = seed
	vec[1] = seed * 0.5
	vec[2] = seed * 0.25
	return vec
}

func TestMemoryRecallAcrossConversations(t *testing.T) {
	pool := memoryTestPool(t)
	user := memoryTestUser(t, pool)
	convRepo := repositories.NewConversationRepository(pool)
	embedRepo := repositories.NewEmbeddingRepository(pool)
	ctx := context.Background()

	// During the first conversation the assistant stores a turn and its
	// embedding ("memory written").
	turn, err := convRepo.Create(ctx, &models.Conversation{
		UserID:  user.ID,
		Role:    models.ConversationRoleUser,
		Content: "I specialize in Go and distributed systems",
	})
	if err != nil {
		t.Fatalf("create turn: %v", err)
	}

	_, err = embedRepo.Upsert(ctx, &repositories.Embedding{
		UserID:     user.ID,
		SourceType: repositories.EmbeddingSourceConversation,
		SourceID:   turn.ID,
		Vector:     embeddingVector(1.0),
	})
	if err != nil {
		t.Fatalf("upsert embedding: %v", err)
	}

	// During a later conversation a similar query must recall that memory.
	results, err := embedRepo.FindSimilar(ctx, embeddingVector(0.8), repositories.EmbeddingSourceConversation, user.ID, 5)
	if err != nil {
		t.Fatalf("find similar: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected the stored memory to be recalled")
	}
	if results[0].SourceID != turn.ID {
		t.Fatalf("expected nearest result to be the stored turn, got %+v", results[0])
	}
}

func TestMemoryRecallUserIsolation(t *testing.T) {
	pool := memoryTestPool(t)
	userA := memoryTestUser(t, pool)
	userB := memoryTestUser(t, pool)
	convRepo := repositories.NewConversationRepository(pool)
	embedRepo := repositories.NewEmbeddingRepository(pool)
	ctx := context.Background()

	store := func(user *repositories.User, content string) string {
		t.Helper()
		turn, err := convRepo.Create(ctx, &models.Conversation{UserID: user.ID, Role: models.ConversationRoleUser, Content: content})
		if err != nil {
			t.Fatalf("create turn for %s: %v", user.ID, err)
		}
		if _, err := embedRepo.Upsert(ctx, &repositories.Embedding{
			UserID:     user.ID,
			SourceType: repositories.EmbeddingSourceConversation,
			SourceID:   turn.ID,
			Vector:     embeddingVector(1.0),
		}); err != nil {
			t.Fatalf("upsert embedding for %s: %v", user.ID, err)
		}
		return turn.ID
	}

	turnA := store(userA, "memory of user A")
	store(userB, "memory of user B")

	// Querying user A's memory must not surface user B's embedding.
	results, err := embedRepo.FindSimilar(ctx, embeddingVector(0.9), repositories.EmbeddingSourceConversation, userA.ID, 10)
	if err != nil {
		t.Fatalf("find similar: %v", err)
	}
	if len(results) == 0 || results[0].SourceID != turnA {
		t.Fatalf("expected only user A's memory to be recalled, got %+v", results)
	}
	for _, r := range results {
		if r.UserID != userA.ID {
			t.Fatalf("recall leaked another user's embedding: %+v", r)
		}
	}
}

func TestMemoryRecallDatabaseFailure(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	pool.Close() // simulate an unavailable database

	embedRepo := repositories.NewEmbeddingRepository(pool)
	ctx := context.Background()

	if _, err := embedRepo.FindSimilar(ctx, embeddingVector(1.0), repositories.EmbeddingSourceConversation, "user-x", 5); err == nil {
		t.Fatal("expected an error when the database is unavailable")
	}
}
