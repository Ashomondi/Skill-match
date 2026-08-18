package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"skill-match/backend/utils"
)

func TestUnauthenticatedRequestRejected(t *testing.T) {
	jwtManager := utils.NewJWTManager("test-secret", 15*time.Minute)

	handler := Auth(jwtManager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/saved-jobs", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}