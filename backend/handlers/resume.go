package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"skill-match/backend/middleware"
	// "skill-match/backend/models"
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

<<<<<<< HEAD
// Create handles POST /api/resumes — multipart upload with a "resume" file
// part and an optional "replaceId" field.
func (h *ResumeHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartBytes)
	if err := r.ParseMultipartForm(maxMultipartBytes); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid multipart form or file too large"})
=======
// Upload handles POST /api/resumes.
func (h *ResumeHandler) Upload(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		writeResumeJSON(
			w,
			http.StatusMethodNotAllowed,
			map[string]string{
				"error": "method not allowed",
			},
		)
		return
	}

	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		utils.MaxResumeSize+1024,
	)

	if err := r.ParseMultipartForm(
		utils.MaxResumeSize,
	); err != nil {
		writeResumeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "invalid multipart form",
			},
		)
>>>>>>> dev
		return
	}

	file, header, err := r.FormFile("resume")
	if err != nil {
<<<<<<< HEAD
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "resume file is required"})
=======
		writeResumeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "resume file is required",
			},
		)
>>>>>>> dev
		return
	}
	defer file.Close()

<<<<<<< HEAD
	data, err := io.ReadAll(file)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "could not read resume file"})
=======
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeResumeJSON(
			w,
			http.StatusUnauthorized,
			map[string]string{
				"error": "authentication required",
			},
		)
>>>>>>> dev
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
<<<<<<< HEAD
		writeResumeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toResumeResponse(resume, ""))
}

// List handles GET /api/resumes.
func (h *ResumeHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
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
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
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
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
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

	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to process resume"})
	}
=======
		switch {
		case errors.Is(err, utils.ErrInvalidFileType):
			writeResumeJSON(
				w,
				http.StatusBadRequest,
				map[string]string{
					"error": "only PDF and DOCX resumes are allowed",
				},
			)

		case errors.Is(err, utils.ErrFileTooLarge):
			writeResumeJSON(
				w,
				http.StatusBadRequest,
				map[string]string{
					"error": "resume must not exceed 5 MB",
				},
			)

		case errors.Is(err, utils.ErrEmptyFile):
			writeResumeJSON(
				w,
				http.StatusBadRequest,
				map[string]string{
					"error": "resume file cannot be empty",
				},
			)

		default:
			utils.WriteRequestError(w, r, err)
		}

		return
	}

	writeResumeJSON(
		w,
		http.StatusCreated,
		map[string]interface{}{
			"message":   "resume uploaded successfully",
			"resume_id": resume.ID,
			"status":    resume.Status,
			"resume":    resume,
		},
	)
}

// Update handles PUT /api/resumes/{id}.
func (h *ResumeHandler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPut {
		writeResumeJSON(
			w,
			http.StatusMethodNotAllowed,
			map[string]string{
				"error": "method not allowed",
			},
		)
		return
	}

	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		utils.MaxResumeSize+1024,
	)

	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeResumeJSON(
			w,
			http.StatusUnauthorized,
			map[string]string{
				"error": "authentication required",
			},
		)
		return
	}

	resumeID := r.PathValue("id")

	if strings.TrimSpace(resumeID) == "" {
		writeResumeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "resume ID is required",
			},
		)
		return
	}

	if err := r.ParseMultipartForm(
		utils.MaxResumeSize,
	); err != nil {
		writeResumeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "invalid multipart form",
			},
		)
		return
	}

	file, header, err := r.FormFile("resume")
	if err != nil {
		writeResumeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "resume file is required",
			},
		)
		return
	}
	defer file.Close()

	updatedResume, err := h.resumeService.Update(
		r.Context(),
		services.UpdateResumeInput{
			UserID:   userID,
			ResumeID: resumeID,
			Filename: header.Filename,
			File:     file,
			Size:     header.Size,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrResumeNotFound):
			writeResumeJSON(
				w,
				http.StatusNotFound,
				map[string]string{
					"error": "resume not found",
				},
			)

		case errors.Is(err, services.ErrResumeUnauthorized):
			writeResumeJSON(
				w,
				http.StatusForbidden,
				map[string]string{
					"error": "you do not have permission to update this resume",
				},
			)

		case errors.Is(err, utils.ErrInvalidFileType):
			writeResumeJSON(
				w,
				http.StatusBadRequest,
				map[string]string{
					"error": "only PDF and DOCX resumes are allowed",
				},
			)

		case errors.Is(err, utils.ErrFileTooLarge):
			writeResumeJSON(
				w,
				http.StatusBadRequest,
				map[string]string{
					"error": "resume must not exceed 5 MB",
				},
			)

		case errors.Is(err, utils.ErrEmptyFile):
			writeResumeJSON(
				w,
				http.StatusBadRequest,
				map[string]string{
					"error": "resume file cannot be empty",
				},
			)

		default:
			utils.WriteRequestError(w, r, err)
		}

		return
	}

	writeResumeJSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"message":   "resume updated successfully",
			"resume_id": updatedResume.ID,
			"version":   updatedResume.Version,
			"status":    updatedResume.Status,
			"resume":    updatedResume,
		},
	)
}

// Delete handles DELETE /api/resumes/{id}.
func (h *ResumeHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodDelete {
		writeResumeJSON(
			w,
			http.StatusMethodNotAllowed,
			map[string]string{
				"error": "method not allowed",
			},
		)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeResumeJSON(
			w,
			http.StatusUnauthorized,
			map[string]string{
				"error": "authentication required",
			},
		)
		return
	}

	// Get resume ID from:
	// DELETE /api/resumes/{id}
	resumeID := strings.TrimSpace(
		r.PathValue("id"),
	)

	if resumeID == "" {
		writeResumeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "resume ID is required",
			},
		)
		return
	}

	err := h.resumeService.Delete(
		r.Context(),
		userID,
		resumeID,
	)

	if err != nil {
		switch {
		case errors.Is(err, services.ErrResumeNotFound):
			writeResumeJSON(
				w,
				http.StatusNotFound,
				map[string]string{
					"error": "resume not found",
				},
			)

		case errors.Is(err, services.ErrResumeUnauthorized):
			writeResumeJSON(
				w,
				http.StatusForbidden,
				map[string]string{
					"error": "you do not have permission to delete this resume",
				},
			)

		default:
			utils.WriteRequestError(w, r, err)
		}

		return
	}

	writeResumeJSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"message":   "resume deleted successfully",
			"resume_id": resumeID,
		},
	)
}

// writeResumeJSON writes a JSON response.
func writeResumeJSON(
	w http.ResponseWriter,
	status int,
	data interface{},
) {
	if errorPayload, ok := data.(map[string]string); ok {
		if message, exists := errorPayload["error"]; exists {
			category := utils.CategoryValidation
			switch status {
			case http.StatusUnauthorized, http.StatusForbidden:
				category = utils.CategoryAuth
			case http.StatusNotFound:
				category = utils.CategoryNotFound
			case http.StatusInternalServerError:
				category = utils.CategoryInternal
			}
			utils.WriteJSON(w, status, utils.ErrorResponse{Error: utils.ErrorBody{Message: message, Code: string(category)}})
			return
		}
	}
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
>>>>>>> dev
}
