package handlers

import (
	"errors"
	"io"
	"net/http"
	"time"

	"skill-match/backend/middleware"
	"skill-match/backend/models"
	"skill-match/backend/services"
	"skill-match/backend/utils"
)

// maxMultipartBytes caps the whole request body: file payload (5MB) plus a
// small allowance for multipart framing and the replaceId field.
const maxMultipartBytes = utils.MaxResumeFileSize + 1<<20

type ResumeHandler struct {
	resumeService *services.ResumeService
}

func NewResumeHandler(resumeService *services.ResumeService) *ResumeHandler {
	return &ResumeHandler{resumeService: resumeService}
}

type resumeResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Filename   string    `json:"filename"`
	Size       int64     `json:"size"`
	Status     string    `json:"status"`
	UploadedAt time.Time `json:"uploadedAt"`
	URL        string    `json:"url,omitempty"`
}

func toResumeResponse(r *models.Resume, url string) resumeResponse {
	return resumeResponse{
		ID:         r.ID,
		Name:       r.OriginalFilename,
		Filename:   r.OriginalFilename,
		Size:       r.FileSizeBytes,
		Status:     string(r.Status),
		UploadedAt: r.CreatedAt,
		URL:        url,
	}
}

// Create handles POST /api/resumes — multipart upload with a "resume" file
// part and an optional "replaceId" field.
func (h *ResumeHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartBytes)
	if err := r.ParseMultipartForm(maxMultipartBytes); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid multipart form or file too large"})
		return
	}

	file, header, err := r.FormFile("resume")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "resume file is required"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "could not read resume file"})
		return
	}

	resume, err := h.resumeService.Upload(
		r.Context(),
		userID,
		r.FormValue("replaceId"),
		header.Filename,
		header.Header.Get("Content-Type"),
		data,
	)
	if err != nil {
		writeResumeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toResumeResponse(resume, ""))
}

// List handles GET /api/resumes.
func (h *ResumeHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	resumes, err := h.resumeService.List(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load resumes"})
		return
	}

	out := make([]resumeResponse, 0, len(resumes))
	for _, res := range resumes {
		out = append(out, toResumeResponse(res, ""))
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"resumes": out})
}

// Get handles GET /api/resumes/{id} and includes a time-limited download URL.
func (h *ResumeHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	resume, url, err := h.resumeService.DownloadURL(r.Context(), userID, r.PathValue("id"), 15*time.Minute)
	if err != nil {
		writeResumeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toResumeResponse(resume, url))
}

// Delete handles DELETE /api/resumes/{id}.
func (h *ResumeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if err := h.resumeService.Delete(r.Context(), userID, r.PathValue("id")); err != nil {
		writeResumeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeResumeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrResumeNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "resume not found"})

	case errors.Is(err, services.ErrResumeAccessDenied):
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "you do not have access to this resume"})

	case errors.Is(err, services.ErrInvalidResume):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})

	case errors.Is(err, services.ErrStorageUnavailable):
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "object storage is not configured; resume uploads and downloads are unavailable"})

	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to process resume"})
	}
}
