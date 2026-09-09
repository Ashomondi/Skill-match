package clients

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLocalFSRoundTrip(t *testing.T) {
	dir := t.TempDir()
	fs, err := NewLocalFS(dir, "/storage")
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	key := fs.Key("user-123", "file-abc.txt")
	if key != filepath.Join("resumes", "user-123", "file-abc.txt") {
		t.Fatalf("unexpected key: %q", key)
	}

	payload := []byte("hello resume")
	if err := fs.Put(context.Background(), key, payload, "text/plain"); err != nil {
		t.Fatalf("put: %v", err)
	}

	got, err := fs.Read(key)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != string(payload) {
		t.Fatalf("round-trip mismatch: got %q want %q", got, payload)
	}

	url, err := fs.PresignDownload(context.Background(), key, time.Minute)
	if err != nil {
		t.Fatalf("presign: %v", err)
	}
	if url != "/storage/"+filepath.ToSlash(key) {
		t.Fatalf("unexpected download url: %q", url)
	}

	if err := fs.Delete(context.Background(), key); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := fs.Read(key); !os.IsNotExist(err) {
		t.Fatalf("expected file to be gone, got err=%v", err)
	}
	// Deleting again is idempotent.
	if err := fs.Delete(context.Background(), key); err != nil {
		t.Fatalf("delete idempotency: %v", err)
	}
}

func TestLocalFSHandlerServesStoredObject(t *testing.T) {
	dir := t.TempDir()
	fs, err := NewLocalFS(dir, "/storage")
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	key := fs.Key("u1", "f1.txt")
	if err := fs.Put(context.Background(), key, []byte("content"), "text/plain"); err != nil {
		t.Fatalf("put: %v", err)
	}

	handler := fs.Handler()
	req := httptest.NewRequest(http.MethodGet, "/storage/"+filepath.ToSlash(key), nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "content" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}

func TestLocalFSRejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	fs, err := NewLocalFS(dir, "/storage")
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// Escape via ".." must be rejected by Put and by the serving handler.
	if err := fs.Put(context.Background(), "../escape.txt", []byte("x"), "text/plain"); err == nil {
		t.Fatal("expected Put to reject a traversal key")
	}

	handler := fs.Handler()
	req := httptest.NewRequest(http.MethodGet, "/storage/../secret.txt", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatal("expected traversal request to be rejected")
	}
}
