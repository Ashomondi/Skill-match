package middleware

import (
	"context"
	"net/http"
	"strings"

	"skill-match/backend/utils"
)

type contextKey string

const claimsKey contextKey = "claims"

// Auth validates the JWT Authorization header and adds the authenticated
// user's claims to the request context.
func Auth(jwtManager *utils.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if authHeader == "" {
				utils.WriteError(w, utils.NewValidationError("Authorization header is required.", nil))
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				utils.WriteError(w, utils.NewValidationError("Invalid authorization header.", nil))
				return
			}

			token := strings.TrimSpace(parts[1])
			if token == "" {
				utils.WriteError(w, utils.NewValidationError("Token is required.", nil))
				return
			}

			claims, err := jwtManager.ValidateToken(token)
			if err != nil {
				utils.WriteError(w, utils.NewValidationError("Invalid or expired token.", nil))
				return
			}

			if strings.TrimSpace(claims.UserID) == "" {
				utils.WriteError(w, utils.NewValidationError("Invalid token claims.", nil))
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFromContext retrieves the authenticated user's JWT claims from
// the context, or nil if unauthenticated.
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