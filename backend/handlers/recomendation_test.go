package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecommendationsMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/recommendations", nil)
	w := httptest.NewRecorder()

	h := NewRecommendationHandler(nil)
	h.GetPersonalizedRecommendations(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestRecommendationsMissingAuth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/recommendations?user_id=usr_123", nil)
	w := httptest.NewRecorder()

	h := NewRecommendationHandler(nil)
	h.GetPersonalizedRecommendations(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d for missing authentication, got %d", http.StatusUnauthorized, w.Code)
	}
}
