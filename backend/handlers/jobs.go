package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"skill-match/backend/repositories"
	"skill-match/backend/services"
)

type JobsHandler struct {
	jobService *services.JobService
}

type MatchJobsRequest struct {
	Skills   []string `json:"skills"`
	MinScore float64  `json:"min_score"`
	Limit    int      `json:"limit"`
}

func NewJobsHandler(jobService *services.JobService) *JobsHandler {
	return &JobsHandler{
		jobService: jobService,
	}
}

func (h *JobsHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "job ID is required"})
		return
	}
	job, err := h.jobService.GetJob(r.Context(), id)
	if err != nil {
		if errors.Is(err, repositories.ErrJobNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load job"})
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (h *JobsHandler) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "method not allowed",
		})
		return
	}

	q := r.URL.Query().Get("q")
	location := r.URL.Query().Get("location")
	company := r.URL.Query().Get("company")
	remoteStr := r.URL.Query().Get("remote")

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	var remote *bool
	if remoteStr != "" {
		val := strings.ToLower(remoteStr) == "true" || remoteStr == "1"
		remote = &val
	}

	filter := repositories.JobSearchFilter{
		Query:    q,
		Location: location,
		Company:  company,
		Remote:   remote,
		Limit:    limit,
		Offset:   offset,
	}

	result, err := h.jobService.SearchJobs(r.Context(), filter)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "failed to search jobs",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"jobs": result.Jobs,
		"pagination": map[string]int{
			"total":  result.Total,
			"limit":  filter.Limit,
			"offset": filter.Offset,
		},
	})
}

func (h *JobsHandler) Match(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	var req MatchJobsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON body"})
		return
	}

	filter := repositories.SemanticMatchFilter{
		UserSkills: req.Skills,
		MinScore:   req.MinScore,
		Limit:      req.Limit,
	}

	matches, err := h.jobService.MatchJobs(r.Context(), filter)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to match jobs"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"matches": matches,
		"total":   len(matches),
	})
}
