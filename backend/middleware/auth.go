package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"skill-match/backend/utils"
)

type contextKey string

const (
	userIDKey contextKey = "userID"
	claimsKey contextKey = "claims"
)

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
				writeJSON(w, http.StatusUnauthorized, map[string]string{
					"error": "invalid or expired token",
				})
				return
			}

			if strings.TrimSpace(claims.UserID) == "" {
				utils.WriteError(w, utils.NewValidationError("Invalid token claims.", nil))
				return
			}

			ctx := context.WithValue(
				r.Context(),
				userIDKey,
				claims.UserID,
			)

			ctx = context.WithValue(
				ctx,
				claimsKey,
				claims,
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID retrieves the authenticated user's ID from the request context.
func GetUserID(r *http.Request) (string, bool) {
	userID, ok := r.Context().
		Value(userIDKey).(string)

	if !ok || strings.TrimSpace(userID) == "" {
		return "", false
	}

	return userID, true
}

// ClaimsFromContext retrieves JWT claims from the context.
// ClaimsFromContext retrieves the authenticated user's JWT claims from
// the context, or nil if unauthenticated.
func ClaimsFromContext(ctx context.Context) *utils.Claims {
	claims, _ := ctx.Value(claimsKey).(*utils.Claims)
	return claims
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	data interface{},
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}
