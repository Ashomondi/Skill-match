package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApplicationCollectionUnauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/applications", nil)
	w := httptest.NewRecorder()

	h := NewApplicationHandler(nil)
	h.HandleCollection(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestApplicationResourceUnauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/applications/app_123", nil)
	w := httptest.NewRecorder()

	h := NewApplicationHandler(nil)
	h.HandleResource(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}