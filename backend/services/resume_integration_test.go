//go:build integration

package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"skill-match/backend/clients"
	"skill-match/backend/repositories"
)

// TestResumeStorageIntegration covers the full resume storage flow
// end-to-end — upload writes the object to local storage AND the metadata row
// to PostgreSQL, and the download URL round-trips the same bytes.
//
// It is build-tagged and env-guarded so it never runs in a normal `go test
// ./...` or in CI without infrastructure. Provide:
//
//	TEST_DATABASE_URL    postgres://...
//
// Run with:
//
//	go test -tags integration ./services -run TestResumeStorageIntegration -v
func TestResumeStorageIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	ctx := context.Background()

	pool, err := clients.NewPool(ctx, dsn, clients.PoolOptions{})
	if err != nil {
		t.Fatalf("connect to postgres: %v", err)
	}
	defer pool.Close()

	storage, err := clients.NewLocalFS(t.TempDir(), "/storage")
	if err != nil {
		t.Fatalf("init local storage: %v", err)
	}
	server := httptest.NewServer(storage.Handler())
	defer server.Close()

	userRepo := repositories.NewUserRepository(pool)
	user, err := userRepo.Create(ctx, &repositories.User{
		Email:        fmt.Sprintf("integration-%d@skillmatch.local", time.Now().UnixNano()),
		PasswordHash: "integration-placeholder-hash",
		FullName:     "Integration Tester",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() { _ = userRepo.Delete(ctx, user.ID) })

	svc := NewResumeService(repositories.NewResumeRepository(pool), storage)

	body := []byte("%PDF-1.4 integration end-to-end content")
	res, err := svc.Upload(ctx, user.ID, "", "resume.pdf", "application/pdf", body)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	t.Cleanup(func() { _ = svc.Delete(ctx, user.ID, res.ID) })

	if res.ID == "" {
		t.Fatal("expected an id on the created resume")
	}

	// Download via the storage URL and confirm the bytes round-trip.
	_, url, err := svc.DownloadURL(ctx, user.ID, res.ID, time.Minute)
	if err != nil {
		t.Fatalf("download url: %v", err)
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Get(server.URL + url)
	if err != nil {
		t.Fatalf("GET storage url: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from storage download, got %d", resp.StatusCode)
	}
	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(got) != string(body) {
		t.Fatalf("downloaded bytes mismatch: got %q want %q", got, body)
	}

	// The row must be owned by the user and listed back.
	list, err := svc.List(ctx, user.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].ID != res.ID {
		t.Fatalf("expected exactly the uploaded resume to be listed, got %d", len(list))
	}
}
