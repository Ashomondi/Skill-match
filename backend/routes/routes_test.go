package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"skill-match/backend/handlers"
	"skill-match/backend/services"
	"skill-match/backend/utils"
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

func TestRegisterChatRequiresAuth(t *testing.T) {
	mux := NewMux()
	RegisterChat(mux, nil, utils.NewJWTManager("test-secret", time.Hour))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/chat", nil))

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRegisterChatAllowsAuthenticatedRequest(t *testing.T) {
	jwt := utils.NewJWTManager("test-secret", time.Hour)
	mux := NewMux()
	// The service is intentionally unconfigured: the request must be rejected
	// by the service layer (500), not by auth — proving the JWT user ID
	// reached the handler.
	RegisterChat(mux, handlers.NewChatHandler(&services.ChatService{}), jwt)

	token, err := jwt.GenerateToken("user-123", "user@example.com")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader(`{"message":"hello"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code == http.StatusUnauthorized {
		t.Fatal("authenticated chat request was rejected by the auth middleware")
	}
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d (service-level rejection)", w.Code, http.StatusInternalServerError)
	}
}

func TestRegisterTailorRequiresAuth(t *testing.T) {
	mux := NewMux()
	RegisterTailor(mux, nil, utils.NewJWTManager("test-secret", time.Hour))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/tailor", nil))

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRegisterTailorReachesHandler(t *testing.T) {
	jwt := utils.NewJWTManager("test-secret", time.Hour)
	mux := NewMux()
	// Zero-value AIService makes GenerateResponse fail fast (bedrock unset),
	// which the handler surfaces as 502 — proving the route is registered and
	// auth let the authenticated request through.
	RegisterTailor(mux, handlers.NewTailorHandler(&services.AIService{}), jwt)

	token, err := jwt.GenerateToken("user-123", "user@example.com")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	body := strings.NewReader(`{"resume_id":"00000000-0000-0000-0000-000000000001","job_title":"Backend Engineer"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/tailor", body)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code == http.StatusUnauthorized {
		t.Fatal("authenticated tailor request was rejected by the auth middleware")
	}
	if w.Code == http.StatusNotFound {
		t.Fatal("tailor route is not registered")
	}
	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d (AI unavailable)", w.Code, http.StatusBadGateway)
	}
}
