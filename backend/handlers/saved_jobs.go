package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"skill-match/backend/middleware"
	"skill-match/backend/models"
	"skill-match/backend/services"
	"skill-match/backend/utils"
)

type SavedJobsHandler struct{ service *services.SavedJobService }

func NewSavedJobsHandler(service *services.SavedJobService) *SavedJobsHandler {
	return &SavedJobsHandler{service: service}
}

type saveJobRequest struct {
	JobID string `json:"job_id"`
}

func (h *SavedJobsHandler) Save(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeSavedJobError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var input saveJobRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil || strings.TrimSpace(input.JobID) == "" {
		writeSavedJobError(w, http.StatusBadRequest, "a valid job_id is required")
		return
	}
	saved, err := h.service.Save(r.Context(), userID, input.JobID)
	if err != nil {
		h.writeError(w, err)
		return
	}
	utils.WriteSuccess(w, http.StatusCreated, saved)
}

func (h *SavedJobsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeSavedJobError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	items, err := h.service.List(r.Context(), userID)
	if err != nil {
		h.writeError(w, err)
		return
	}
	if items == nil {
		items = make([]*models.SavedJob, 0)
	}
	utils.WriteSuccess(w, http.StatusOK, items)
}

func (h *SavedJobsHandler) Remove(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeSavedJobError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if err := h.service.Remove(r.Context(), userID, r.PathValue("job_id")); err != nil {
		h.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SavedJobsHandler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrSavedJobInvalidInput):
		writeSavedJobError(w, http.StatusBadRequest, "a valid job ID is required")
	case errors.Is(err, services.ErrSavedJobDuplicate):
		writeSavedJobError(w, http.StatusConflict, "job is already saved")
	case errors.Is(err, services.ErrSavedJobNotFound):
		writeSavedJobError(w, http.StatusNotFound, "job or saved job not found")
	default:
		writeSavedJobError(w, http.StatusInternalServerError, "failed to manage saved jobs")
	}
}
// func (h *SavedJobsHandler) HandleSavedJobs(w http.ResponseWriter, r *http.Request) { h.List(w, r) }
// func (h *SavedJobsHandler) HandleDeleteSavedJob(w http.ResponseWriter, r *http.Request) {
// 	h.Remove(w, r)
// }

func writeSavedJobError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
