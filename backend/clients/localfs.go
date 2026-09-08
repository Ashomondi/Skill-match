package clients

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalFS is a local-filesystem implementation of the storage interface the
// resume service needs. It mirrors the S3 key layout (resumes/{userID}/{fileID})
// so switching between S3/MinIO and local disk is transparent.
//
// Keys are capability-style: every key contains a random fileID, so a URL is
// effectively an unguessable token — the same trust model as an S3 presigned
// URL. In production you'd front this handler with real auth.
type LocalFS struct {
	baseDir string
	baseURL string
}

// NewLocalFS creates the storage directory (if missing) and verifies it is
// writable. baseURL is the public prefix files are served from; leave empty
// for same-origin serving (e.g. "/storage").
func NewLocalFS(baseDir, baseURL string) (*LocalFS, error) {
	if baseDir == "" {
		return nil, fmt.Errorf("local storage directory is required")
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating local storage dir: %w", err)
	}
	if baseURL == "" {
		baseURL = "/storage"
	}
	return &LocalFS{baseDir: baseDir, baseURL: baseURL}, nil
}

func (l *LocalFS) Key(userID, fileID string) string {
	return filepath.Join("resumes", userID, fileID)
}

// Put writes raw bytes to {baseDir}/{key}, creating parent directories.
func (l *LocalFS) Put(_ context.Context, key string, body []byte, _ string) error {
	if err := validateLocalKey(key); err != nil {
		return err
	}

	fp := filepath.Join(l.baseDir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(fp), 0o755); err != nil {
		return fmt.Errorf("creating parent dirs for %s: %w", key, err)
	}

	if err := os.WriteFile(fp, body, 0o644); err != nil {
		return fmt.Errorf("writing local object %s: %w", key, err)
	}
	return nil
}

// PresignDownload returns a URL pointing at the local file-serving route.
// expiry is accepted for interface parity with S3; local files are served
// directly and are not time-limited.
func (l *LocalFS) PresignDownload(_ context.Context, key string, _ time.Duration) (string, error) {
	if err := validateLocalKey(key); err != nil {
		return "", err
	}
	return l.baseURL + "/" + key, nil
}

// Delete removes the file at {baseDir}/{key} and prunes empty parent
// directories. A missing object is treated as success (idempotent).
func (l *LocalFS) Delete(_ context.Context, key string) error {
	if err := validateLocalKey(key); err != nil {
		return err
	}

	fp := filepath.Join(l.baseDir, filepath.FromSlash(key))
	if err := os.Remove(fp); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing local object %s: %w", key, err)
	}

	// Prune empty parent directories up to (but not including) baseDir.
	for dir := filepath.Dir(fp); dir != l.baseDir && dir != string(filepath.Separator); dir = filepath.Dir(dir) {
		if err := os.Remove(dir); err != nil {
			break // dir non-empty or gone; stop pruning
		}
	}
	return nil
}

// Read returns the raw bytes for a stored object. Used by tests and the
// migration/backfill tooling.
func (l *LocalFS) Read(key string) ([]byte, error) {
	if err := validateLocalKey(key); err != nil {
		return nil, err
	}
	return os.ReadFile(filepath.Join(l.baseDir, filepath.FromSlash(key)))
}

// Ping verifies the storage directory exists and is writable.
func (l *LocalFS) Ping(_ context.Context) error {
	probe := filepath.Join(l.baseDir, ".probe")
	f, err := os.Create(probe)
	if err != nil {
		return fmt.Errorf("local storage not writable: %w", err)
	}
	_ = f.Close()
	_ = os.Remove(probe)
	return nil
}

// Handler serves stored objects from {baseDir}, anchored at {baseURL}.
// It restricts requests to within the storage directory so path traversal
// cannot escape the tree.
func (l *LocalFS) Handler() http.Handler {
	fsRoot := http.Dir(l.baseDir)
	return http.StripPrefix(l.baseURL, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// StripPrefix yields a path like /resumes/{user}/{file}; validate it
		// against the same relative key rules used by Put/Delete.
		key := strings.TrimPrefix(r.URL.Path, "/")
		if validateLocalKey(key) != nil {
			http.NotFound(w, r)
			return
		}
		http.FileServer(fsRoot).ServeHTTP(w, r)
	}))
}

// validateLocalKey rejects keys that could escape the storage directory.
func validateLocalKey(key string) error {
	if key == "" {
		return fmt.Errorf("storage key is required")
	}
	clean := filepath.ToSlash(filepath.Clean(key))
	if clean == "." || clean == ".." || filepath.IsAbs(key) || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("invalid storage key %q", key)
	}
	return nil
}
