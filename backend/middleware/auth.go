package middleware

import (
	"context"
	"net/http"
	"strings"

	"skill-match/backend/utils"
)

type contextKey string

const claimsKey contextKey = "claims"

// Auth validates the Bearer token on a request and places the parsed
// claims into the request context for downstream handlers. Requests
// without a valid token are rejected with 401.
func Auth(jwtManager *utils.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || token == "" {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			claims, err := jwtManager.ValidateToken(token)
			if err != nil {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFromContext returns the authenticated user's claims from the
// request context, or nil when the request is unauthenticated.
func ClaimsFromContext(ctx context.Context) *utils.Claims {
	claims, _ := ctx.Value(claimsKey).(*utils.Claims)
	return claims
}
