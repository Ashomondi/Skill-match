package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"skill-match/backend/utils"
)

type contextKey string

const userIDKey contextKey = "user_id"
func Auth(jwtManager *utils.JWTManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "authorization header is required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "invalid authorization header",
			})
			return
		}

		token := strings.TrimSpace(parts[1])

		if token == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "token is required",
			})
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
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "invalid token claims",
			})
			return
		}

		ctx := context.WithValue(
			r.Context(),
			userIDKey,
			claims.UserID,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(userIDKey).(string)

	if !ok || strings.TrimSpace(userID) == "" {
		return "", false
	}

	return userID, true
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