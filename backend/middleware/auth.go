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
				utils.WriteError(w, utils.NewValidationError("Authorization header is required.", nil))
				return
			}

			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || strings.TrimSpace(token) == "" {
				utils.WriteError(w, utils.NewValidationError("Invalid authorization header.", nil))
				return
			}

			claims, err := jwtManager.ValidateToken(token)
			if err != nil {
				utils.WriteError(w, utils.NewValidationError("Invalid or expired token.", nil))
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

// GetUserID is a convenience wrapper for the common case of just needing
// the authenticated user's ID.
func GetUserID(ctx context.Context) (string, bool) {
	claims := ClaimsFromContext(ctx)
	if claims == nil || strings.TrimSpace(claims.UserID) == "" {
		return "", false
	}
	return claims.UserID, true
}