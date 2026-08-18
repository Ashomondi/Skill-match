package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"skill-match/backend/repositories"
	"skill-match/backend/services"
)

type ApplicationHandler struct {
	service *services.ApplicationService
}

func NewApplicationHandler(service *services.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{service: service}
}

type CreateApplicationRequest struct {
	JobID  string                         `json:"job_id"`
	Status repositories.ApplicationStatus `json:"status"`
	Notes  string                         `json:"notes"`
}

type UpdateApplicationRequest struct {
	Status repositories.ApplicationStatus `json:"status"`
	Notes  string                         `json:"notes"`
}

// HandleCollection handles POST /api/applications and GET /api/applications
func (h *ApplicationHandler) HandleCollection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	switch r.Method {
	case http.MethodPost:
		var req CreateApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
			return
		}

		if strings.TrimSpace(req.JobID) == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "validation: job_id is required"})
			return
		}

		app, err := h.service.CreateApplication(r.Context(), userID, req.JobID, req.Status, req.Notes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to create application"})
			return
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(app)

	case http.MethodGet:
		apps, err := h.service.ListApplications(r.Context(), userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to list applications"})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"applications": apps,
			"total":        len(apps),
		})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
	}
}

// HandleResource handles GET, PUT, and DELETE for /api/applications/{id}
func (h *ApplicationHandler) HandleResource(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/applications/")
	id = strings.TrimSpace(id)
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "application ID is required"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		app, err := h.service.GetApplicationByID(r.Context(), id, userID)
		if err != nil {
			if err.Error() == "not_found" {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "application not found"})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to retrieve application"})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(app)

	case http.MethodPut:
		var req UpdateApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
			return
		}

		app, err := h.service.UpdateApplicationStatus(r.Context(), id, userID, req.Status, req.Notes)
		if err != nil {
			if strings.HasPrefix(err.Error(), "validation:") {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			if err.Error() == "not_found" {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "application not found or access denied"})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to update application"})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(app)

	case http.MethodDelete:
		err := h.service.DeleteApplication(r.Context(), id, userID)
		if err != nil {
			if err.Error() == "not_found" {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "application not found or access denied"})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to delete application"})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "application deleted successfully"})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
	}
}