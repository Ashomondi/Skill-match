package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"skill-match/backend/middleware"
	"skill-match/backend/models"
	"skill-match/backend/services"
	"skill-match/backend/utils"
)

// maxProfileBodyBytes caps the request body for profile updates. A profile is
// small, so 64 KiB is generous while keeping the handler bounded.
const maxProfileBodyBytes = 64 << 10

// ProfileHandler exposes the authenticated user's editable profile.
type ProfileHandler struct {
	service *services.ProfileService
}

func NewProfileHandler(service *services.ProfileService) *ProfileHandler {
	return &ProfileHandler{service: service}
}

type profileResponse struct {
	UserID     string     `json:"user_id"`
	Bio        string     `json:"bio"`
	Skills     []string   `json:"skills"`
	Experience []string   `json:"experience"`
	ResumeURL  string     `json:"resume_url"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

func toProfileResponse(profile *models.UserProfile) profileResponse {
	resp := profileResponse{
		UserID:     profile.UserID,
		Bio:        profile.Bio,
		Skills:     profile.Skills,
		Experience: profile.Experience,
		ResumeURL:  profile.ResumeURL,
	}
	if resp.Skills == nil {
		resp.Skills = []string{}
	}
	if resp.Experience == nil {
		resp.Experience = []string{}
	}
	if !profile.UpdatedAt.IsZero() {
		updated := profile.UpdatedAt
		resp.UpdatedAt = &updated
	}
	return resp
}

// Get handles GET /api/profile.
func (h *ProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}

	profile, err := h.service.Get(r.Context(), userID)
	if err != nil {
		utils.WriteRequestError(w, r, err)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, toProfileResponse(profile))
}

// Update handles PUT /api/profile with a partial payload.
func (h *ProfileHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}

	var patch models.UserProfilePatch
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxProfileBodyBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&patch); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid profile payload"})
		return
	}
	// Reject a second JSON value in the same body.
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid profile payload"})
		return
	}

	profile, err := h.service.Update(r.Context(), userID, patch)
	if err != nil {
		utils.WriteRequestError(w, r, err)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, toProfileResponse(profile))
}
