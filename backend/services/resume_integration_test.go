//go:build integration

package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"skill-match/backend/clients"
	"skill-match/backend/models"
	"skill-match/backend/repositories"
)

// TestResumeStorageIntegration covers Issue #21: the full resume storage
// flow end-to-end — upload writes the object to S3 AND the metadata row to
// CockroachDB, and the presigned download round-trips the same bytes.
//
// It is build-tagged and env-guarded so it never runs in a normal `go test
// ./...` or in CI without infrastructure. Provide:
//
//	TEST_DATABASE_URL    postgres://... (cockroachdb dsn)
//	TEST_S3_ENDPOINT     e.g. http://localhost:9000
//	TEST_S3_BUCKET       bucket name
//	TEST_S3_ACCESS_KEY   access key
//	TEST_S3_SECRET_KEY   secret key
//
// Run with:
//
//	go test -tags integration ./services -run TestResumeStorageIntegration -v
func TestResumeStorageIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	endpoint := os.Getenv("TEST_S3_ENDPOINT")
	bucket := os.Getenv("TEST_S3_BUCKET")
	if dsn == "" || endpoint == "" || bucket == "" {
		t.Skip("TEST_DATABASE_URL / TEST_S3_* not set; skipping integration test")
	}

	ctx := context.Background()

	pool, err := clients.NewPool(ctx, dsn, clients.PoolOptions{})
	if err != nil {
		t.Fatalf("connect to cockroachdb: %v", err)
	}
	defer pool.Close()

	s3Client, err := clients.NewS3Client(ctx, clients.S3Config{
		Region:         "us-east-1",
		Bucket:         bucket,
		Endpoint:       endpoint,
		AccessKey:      os.Getenv("TEST_S3_ACCESS_KEY"),
		SecretKey:      os.Getenv("TEST_S3_SECRET_KEY"),
		ForcePathStyle: true,
	})
	if err != nil {
		t.Fatalf("init s3 client: %v", err)
	}

	userRepo := repositories.NewUserRepository(pool)
	user, err := userRepo.Create(ctx, &models.User{
		Email:    fmt.Sprintf("integration-%d@skillmatch.local", time.Now().UnixNano()),
		Password: "integration-placeholder-hash",
		FullName: "Integration Tester",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() { _ = userRepo.Delete(ctx, user.ID) })

	svc := NewResumeService(repositories.NewResumeRepository(pool), s3Client)

	body := []byte("%PDF-1.4 integration end-to-end content")
	res, err := svc.Upload(ctx, user.ID, "", "resume.pdf", "application/pdf", body)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	t.Cleanup(func() { _ = svc.Delete(ctx, user.ID, res.ID) })

	if res.S3Key == "" || res.ID == "" {
		t.Fatal("expected s3 key and id on the created resume")
	}

	// Download via the presigned URL and confirm the bytes round-trip.
	_, url, err := svc.DownloadURL(ctx, user.ID, res.ID, time.Minute)
	if err != nil {
		t.Fatalf("download url: %v", err)
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Get(url)
	if err != nil {
		t.Fatalf("GET presigned url: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from presigned download, got %d", resp.StatusCode)
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
