package clients

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestVerifier(t *testing.T, status int, body string) *GoogleTokenVerifier {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("id_token"); got != "test-token" {
			t.Errorf("expected id_token query parameter, got %q", got)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	return &GoogleTokenVerifier{
		clientID:     "client-123",
		tokenInfoURL: server.URL,
		httpClient:   server.Client(),
		now:          func() time.Time { return time.Unix(1_700_000_000, 0) },
	}
}

func TestGoogleTokenVerifierVerifySuccess(t *testing.T) {
	verifier := newTestVerifier(t, http.StatusOK, `{
		"iss": "https://accounts.google.com",
		"aud": "client-123",
		"sub": "google-sub-1",
		"email": "Jane@Example.com",
		"email_verified": "true",
		"name": " Jane Doe ",
		"picture": "https://example.com/avatar.png",
		"exp": 1700003600
	}`)

	identity, err := verifier.Verify(context.Background(), "test-token")
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}

	if identity.Subject != "google-sub-1" {
		t.Errorf("Subject = %q, want google-sub-1", identity.Subject)
	}
	if identity.Email != "jane@example.com" {
		t.Errorf("Email = %q, want lowercased jane@example.com", identity.Email)
	}
	if !identity.EmailVerified {
		t.Error("EmailVerified = false, want true")
	}
	if identity.Name != "Jane Doe" {
		t.Errorf("Name = %q, want trimmed Jane Doe", identity.Name)
	}
}

func TestGoogleTokenVerifierRejectsBadTokens(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "wrong audience",
			body: `{"iss":"accounts.google.com","aud":"other-client","sub":"s","email":"a@b.com","email_verified":true,"exp":1700003600}`,
		},
		{
			name: "wrong issuer",
			body: `{"iss":"evil.example.com","aud":"client-123","sub":"s","email":"a@b.com","email_verified":true,"exp":1700003600}`,
		},
		{
			name: "unverified email",
			body: `{"iss":"accounts.google.com","aud":"client-123","sub":"s","email":"a@b.com","email_verified":"false","exp":1700003600}`,
		},
		{
			name: "missing email",
			body: `{"iss":"accounts.google.com","aud":"client-123","sub":"s","email_verified":true,"exp":1700003600}`,
		},
		{
			name: "expired",
			body: `{"iss":"accounts.google.com","aud":"client-123","sub":"s","email":"a@b.com","email_verified":true,"exp":1600000000}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			verifier := newTestVerifier(t, http.StatusOK, tc.body)

			_, err := verifier.Verify(context.Background(), "test-token")
			if !errors.Is(err, ErrGoogleTokenInvalid) {
				t.Fatalf("Verify error = %v, want ErrGoogleTokenInvalid", err)
			}
		})
	}
}

func TestGoogleTokenVerifierClassifiesResponses(t *testing.T) {
	t.Run("rejected token", func(t *testing.T) {
		verifier := newTestVerifier(t, http.StatusBadRequest, `{"error":"invalid_token"}`)
		_, err := verifier.Verify(context.Background(), "test-token")
		if !errors.Is(err, ErrGoogleTokenInvalid) {
			t.Fatalf("Verify error = %v, want ErrGoogleTokenInvalid", err)
		}
	})

	t.Run("verification service down", func(t *testing.T) {
		verifier := newTestVerifier(t, http.StatusInternalServerError, `oops`)
		_, err := verifier.Verify(context.Background(), "test-token")
		if !errors.Is(err, ErrGoogleVerificationUnavailable) {
			t.Fatalf("Verify error = %v, want ErrGoogleVerificationUnavailable", err)
		}
	})

	t.Run("empty token", func(t *testing.T) {
		verifier := newTestVerifier(t, http.StatusOK, `{}`)
		_, err := verifier.Verify(context.Background(), "   ")
		if !errors.Is(err, ErrGoogleTokenInvalid) {
			t.Fatalf("Verify error = %v, want ErrGoogleTokenInvalid", err)
		}
	})
}

func TestNewGoogleTokenVerifierRequiresClientID(t *testing.T) {
	if _, err := NewGoogleTokenVerifier("  ", nil); err == nil {
		t.Fatal("expected an error for an empty client ID")
	}
}
