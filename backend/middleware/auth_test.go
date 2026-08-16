package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"skill-match/backend/utils"
)

func newTestJWTManager(t *testing.T) *utils.JWTManager {
	t.Helper()
	return utils.NewJWTManager("unit-test-secret", time.Hour)
}

func TestAuthAllowsValidToken(t *testing.T) {
	manager := newTestJWTManager(t)
	token, err := manager.GenerateToken("user-123", "user@example.com")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	var gotUserID, gotEmail string
	handler := Auth(manager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = UserIDFromContext(r.Context())
		gotEmail = EmailFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if gotUserID != "user-123" {
		t.Errorf("expected user id user-123, got %q", gotUserID)
	}
	if gotEmail != "user@example.com" {
		t.Errorf("expected email user@example.com, got %q", gotEmail)
	}
}

func TestAuthRejectsMissingToken(t *testing.T) {
	manager := newTestJWTManager(t)
	handler := Auth(manager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without a token")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthRejectsInvalidToken(t *testing.T) {
	manager := newTestJWTManager(t)
	handler := Auth(manager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with an invalid token")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthRejectsWrongSigningSecret(t *testing.T) {
	issuer := utils.NewJWTManager("issuer-secret", time.Hour)
	token, err := issuer.GenerateToken("user-123", "user@example.com")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	handler := Auth(newTestJWTManager(t))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with a token signed by a different secret")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
