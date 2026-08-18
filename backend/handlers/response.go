package handlers

import (
	"net/http"
	"strings"

	"skill-match/backend/middleware"
	"skill-match/backend/utils"
)

// requestUserID resolves the authenticated user, preferring the JWT context
// injected by middleware.Auth and falling back to the X-User-ID header (used
// by some handlers and their unit tests).
func requestUserID(r *http.Request) (string, bool) {
	if uid, ok := middleware.GetUserID(r); ok {
		return uid, true
	}
	if uid := strings.TrimSpace(r.Header.Get("X-User-ID")); uid != "" {
		return uid, true
	}
	return "", false
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	if errorPayload, ok := data.(map[string]string); ok {
		if message, exists := errorPayload["error"]; exists {
			category := utils.CategoryValidation
			switch status {
			case http.StatusUnauthorized, http.StatusForbidden:
				category = utils.CategoryAuth
			case http.StatusNotFound:
				category = utils.CategoryNotFound
			case http.StatusConflict:
				category = utils.CategoryConflict
			case http.StatusInternalServerError:
				category = utils.CategoryInternal
			case http.StatusServiceUnavailable, http.StatusBadGateway:
				category = utils.CategoryUpstream
			}
			utils.WriteJSON(w, status, utils.ErrorResponse{Error: utils.ErrorBody{
				Message: message,
				Code:    string(category),
			}})
			return
		}
	}
	utils.WriteJSON(w, status, data)
}
