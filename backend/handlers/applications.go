package handlers

import (
	"net/http"

	"skill-match/backend/middleware"
	"skill-match/backend/models"
	"skill-match/backend/services"
)

type ApplicationHandler struct {
	svc *services.ApplicationService
}

func NewApplicationHandler(svc *services.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{svc: svc}
}

// List handles GET /api/applications — the authenticated user's applications,
// most recently updated first, enriched with job title/company.
func (h *ApplicationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}

	applications, err := h.svc.List(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load applications"})
		return
	}
	if applications == nil {
		applications = []*models.Application{}
	}

	// Flat {applications: [...]} — the frontend reads body.applications.
	writeJSON(w, http.StatusOK, map[string]interface{}{"applications": applications})
}
