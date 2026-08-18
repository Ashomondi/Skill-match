package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"skill-match/backend/services"
)

type RecommendationHandler struct {
	recService *services.RecommendationService
}

func NewRecommendationHandler(recService *services.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{recService: recService}
}

func (h *RecommendationHandler) GetPersonalizedRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	// Retrieve authenticated user from header/context
	requestingUserID := r.Header.Get("X-User-ID")
	targetUserID := r.URL.Query().Get("user_id")

	if requestingUserID == "" || targetUserID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "missing user_id or authentication context"})
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}

	recommendations, err := h.recService.RecommendForUser(r.Context(), requestingUserID, targetUserID, limit)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if err.Error() == "services: unauthorized access to user profile data" {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "access denied: cannot view another user's recommendations"})
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to generate recommendations"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"recommendations": recommendations,
		"total":           len(recommendations),
	})
}
