package handlers

import (
	"encoding/json"
	"errors"
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
func (h *ApplicationHandler) HandleCollection(w http.ResponseWriter, r *http.Request) { h.List(w, r) }
func (h *ApplicationHandler) HandleResource(w http.ResponseWriter, r *http.Request)   { h.List(w, r) }

type createApplicationRequest struct {
	JobID string `json:"job_id"`
}
type updateApplicationRequest struct {
	Status models.ApplicationStatus `json:"status"`
}

func (h *ApplicationHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}
	var input createApplicationRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&input); err != nil || input.JobID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "a valid job_id is required"})
		return
	}
	application, err := h.svc.Create(r.Context(), userID, input.JobID)
	if err != nil {
		writeApplicationError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, application)
}

func (h *ApplicationHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}
	var input updateApplicationRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&input); err != nil || !input.Status.Valid() {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "a valid status is required"})
		return
	}
	application, err := h.svc.UpdateStatus(r.Context(), userID, r.PathValue("id"), input.Status)
	if err != nil {
		writeApplicationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, application)
}

func writeApplicationError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := "failed to manage application"
	switch {
	case errors.Is(err, services.ErrApplicationInvalidInput), errors.Is(err, services.ErrApplicationInvalidStatus):
		status = http.StatusBadRequest
		message = "invalid application request"
	case errors.Is(err, services.ErrApplicationDuplicate):
		status = http.StatusConflict
		message = "application already exists"
	case errors.Is(err, services.ErrApplicationNotFound):
		status = http.StatusNotFound
		message = "application or job not found"
	case errors.Is(err, services.ErrApplicationInvalidTransition):
		status = http.StatusConflict
		message = "invalid application status transition"
	}
	writeJSON(w, status, map[string]string{"error": message})
}

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
