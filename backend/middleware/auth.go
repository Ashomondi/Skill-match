package middleware

import (
	"context"
	"net/http"
	"strings"

	"skill-match/backend/utils"
)

type contextKey string

const (
	contextKeyUserID contextKey = "user_id"
	contextKeyEmail  contextKey = "email"
)

// Auth protects a handler behind a valid JWT. It validates the
// Authorization: Bearer <token> header, then injects the authenticated
// user's ID and email into the request context for downstream handlers and
// services.
func Auth(jwtManager *utils.JWTManager) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				writeUnauthorized(w)
				return
			}

			claims, err := jwtManager.ValidateToken(strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				writeUnauthorized(w)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, contextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, contextKeyEmail, claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext returns the authenticated user's ID injected by Auth,
// or "" when the request is not authenticated.
func UserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(contextKeyUserID).(string); ok {
		return v
	}
	return ""
}

// EmailFromContext returns the authenticated user's email injected by Auth,
// or "" when the request is not authenticated.
func EmailFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(contextKeyEmail).(string); ok {
		return v
	}
	return ""
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"unauthorized"}`))
}
