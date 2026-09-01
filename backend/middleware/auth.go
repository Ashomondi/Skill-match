package middleware

import (
	"context"
	"net/http"
	"strings"

	"skill-match/backend/utils"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
	ClaimsKey contextKey = "claims"
)

// Auth middleware validates incoming JWT tokens and injects claims into request context.
func Auth(jwtManager *utils.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if authHeader == "" {
				utils.WriteRequestError(w, r, utils.NewAuthError("Authentication is required.", http.StatusUnauthorized))
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				utils.WriteRequestError(w, r, utils.NewAuthError("Invalid authorization header.", http.StatusUnauthorized))
				return
			}

			token := strings.TrimSpace(parts[1])
			if token == "" {
				utils.WriteRequestError(w, r, utils.NewAuthError("Authentication is required.", http.StatusUnauthorized))
				return
			}

			claims, err := jwtManager.ValidateToken(token)
			if err != nil {
				utils.WriteRequestError(w, r, utils.NewAuthError("Invalid or expired token.", http.StatusUnauthorized))
				return
			}

			if strings.TrimSpace(claims.UserID) == "" {
				utils.WriteRequestError(w, r, utils.NewAuthError("Invalid authentication claims.", http.StatusUnauthorized))
				return
			}

			ctx := context.WithValue(
				r.Context(),
				UserIDKey,
				claims.UserID,
			)

			ctx = context.WithValue(
				ctx,
				ClaimsKey,
				claims,
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID retrieves the authenticated user's ID from the request context.
func GetUserID(r *http.Request) (string, bool) {
	userID, ok := r.Context().
		Value(UserIDKey).(string)

	if !ok || strings.TrimSpace(userID) == "" {
		return "", false
	}

	return userID, true
}

// ClaimsFromContext retrieves the authenticated user's JWT claims from the context, or nil if unauthenticated.
func ClaimsFromContext(ctx context.Context) *utils.Claims {
	claims, _ := ctx.Value(ClaimsKey).(*utils.Claims)
	return claims
}
