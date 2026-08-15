package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type healthPinger struct{ err error }

func (p healthPinger) Ping(context.Context) error { return p.err }

func TestHealthReturnsHealthyWhenDependenciesRespond(t *testing.T) {
	h := &HealthHandler{db: healthPinger{}, s3Client: healthPinger{}}
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	h.Health(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var body healthResponse
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != "healthy" || body.Dependencies["database"].Status != "ok" || body.Dependencies["storage"].Status != "ok" {
		t.Fatalf("unexpected health response: %+v", body)
	}
}

func TestHealthReturnsServiceUnavailableForFailedDependency(t *testing.T) {
	h := &HealthHandler{db: healthPinger{err: errors.New("connection refused")}, s3Client: healthPinger{}}
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	h.Health(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
	var body healthResponse
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != "degraded" || body.Dependencies["database"].Error != "database unreachable" {
		t.Fatalf("unexpected health response: %+v", body)
	}
}

func TestHealthTreatsUnconfiguredDependencyAsUnavailable(t *testing.T) {
	h := &HealthHandler{}
	w := httptest.NewRecorder()
	h.Health(w, httptest.NewRequest(http.MethodGet, "/health", nil))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}
