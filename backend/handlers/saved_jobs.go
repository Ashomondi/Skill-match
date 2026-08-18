package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"skill-match/backend/services"
)

type SavedJobsHandler struct {
	service *services.SavedJobService
}

func NewSavedJobsHandler(service *services.SavedJobService) *SavedJobsHandler {
	return &SavedJobsHandler{service: service}
}

type SaveJobRequest struct {
	JobID string `json:"job_id"`
}

func (h *SavedJobsHandler) HandleSavedJobs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Authentication check
	userID, ok := requestUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	switch r.Method {
	case http.MethodPost:
		var req SaveJobRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.JobID) == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "job_id is required"})
			return
		}

		sj, err := h.service.SaveJob(r.Context(), userID, req.JobID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to save job"})
			return
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(sj)

	case http.MethodGet:
		jobs, err := h.service.GetSavedJobs(r.Context(), userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to retrieve saved jobs"})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"saved_jobs": jobs,
			"total":      len(jobs),
		})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
	}
}

func (h *SavedJobsHandler) HandleDeleteSavedJob(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	userID, ok := requestUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	// Extract ID from path: /api/saved-jobs/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/saved-jobs/")
	id := strings.TrimSpace(path)
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "saved job ID required"})
		return
	}

	err := h.service.DeleteSavedJob(r.Context(), id, userID)
	if err != nil {
		if err.Error() == "not_found" {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "saved job not found or not owned by user"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to delete saved job"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "saved job deleted successfully"})
}
