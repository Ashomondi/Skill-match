package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSavedJobsUnauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/saved-jobs", nil)
	w := httptest.NewRecorder()

	h := NewSavedJobsHandler(nil)
	h.HandleSavedJobs(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestDeleteSavedJobUnauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/saved-jobs/123", nil)
	w := httptest.NewRecorder()

	h := NewSavedJobsHandler(nil)
	h.HandleDeleteSavedJob(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
