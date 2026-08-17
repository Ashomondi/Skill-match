package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/jobs/search", nil)
	w := httptest.NewRecorder()

	h := NewJobsHandler(nil)
	h.Search(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}