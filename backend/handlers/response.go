package handlers

import (
	"net/http"

	"skill-match/backend/utils"
)

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
