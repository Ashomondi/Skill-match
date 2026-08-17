package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"skill-match/backend/handlers"
)

func TestRegisterHealthMakesEndpointAvailable(t *testing.T) {
	mux := NewMux()
	RegisterHealth(mux, handlers.NewHealthHandler(nil, nil))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("content type = %q, want application/json", contentType)
	}
}
