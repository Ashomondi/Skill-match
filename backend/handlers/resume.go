package handlers

import (
	"errors"
	"net/http"
	"strings"

	"skill-match/backend/services"
	"skill-match/backend/utils"
)

type ResumeHandler struct {
	resumeService *services.ResumeService
}

func NewResumeHandler(
	resumeService *services.ResumeService,
) *ResumeHandler {
	return &ResumeHandler{
		resumeService: resumeService,
	}
}

// Upload handles POST /api/resumes.
func (h *ResumeHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeResumeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	// Prevent unnecessarily large requests from reaching the application.
	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		utils.MaxResumeSize+1024,
	)

	if err := r.ParseMultipartForm(utils.MaxResumeSize); err != nil {
		writeResumeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid multipart form",
		})
		return
	}

	file, header, err := r.FormFile("resume")
	if err != nil {
		writeResumeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "resume file is required",
		})
		return
	}
	defer file.Close()

	// The authentication middleware places the user ID
	// into the request context.
	userIDValue := r.Context().Value("user_id")

	userID, ok := userIDValue.(string)
	if !ok || strings.TrimSpace(userID) == "" {
		writeResumeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
		return
	}

	resume, err := h.resumeService.Upload(
		r.Context(),
		services.UploadResumeInput{
			UserID:   userID,
			Filename: header.Filename,
			File:     file,
			Size:     header.Size,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrInvalidFileType):
			writeResumeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "only PDF and DOCX resumes are allowed",
			})

		case errors.Is(err, utils.ErrFileTooLarge):
			writeResumeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "resume must not exceed 5 MB",
			})

		case errors.Is(err, utils.ErrEmptyFile):
			writeResumeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "resume file cannot be empty",
			})

		default:
			writeResumeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to upload resume",
			})
		}

		return
	}

	writeResumeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":   "resume uploaded successfully",
		"resume_id": resume.ID,
		"status":    resume.Status,
		"resume":    resume,
	})
}

// Update handles PUT /api/resumes/{id}.
func (h *ResumeHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeResumeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	// Prevent unnecessarily large requests from reaching the application.
	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		utils.MaxResumeSize+1024,
	)

	// Get the authenticated user ID from the request context.
	userIDValue := r.Context().Value("user_id")

	userID, ok := userIDValue.(string)
	if !ok || strings.TrimSpace(userID) == "" {
		writeResumeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
		return
	}

	// Go's net/http ServeMux supports path parameters.
	resumeID := r.PathValue("id")

	if strings.TrimSpace(resumeID) == "" {
		writeResumeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "resume ID is required",
		})
		return
	}

	if err := r.ParseMultipartForm(utils.MaxResumeSize); err != nil {
		writeResumeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid multipart form",
		})
		return
	}

	file, header, err := r.FormFile("resume")
	if err != nil {
		writeResumeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "resume file is required",
		})
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
			writeResumeJSON(w, http.StatusNotFound, map[string]string{
				"error": "resume not found",
			})

		case errors.Is(err, services.ErrResumeUnauthorized):
			writeResumeJSON(w, http.StatusForbidden, map[string]string{
				"error": "you do not have permission to update this resume",
			})

		case errors.Is(err, utils.ErrInvalidFileType):
			writeResumeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "only PDF and DOCX resumes are allowed",
			})

		case errors.Is(err, utils.ErrFileTooLarge):
			writeResumeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "resume must not exceed 5 MB",
			})

		case errors.Is(err, utils.ErrEmptyFile):
			writeResumeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "resume file cannot be empty",
			})

		default:
			writeResumeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to update resume",
			})
		}

		return
	}

	writeResumeJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "resume updated successfully",
		"resume_id": updatedResume.ID,
		"version":   updatedResume.Version,
		"status":    updatedResume.Status,
		"resume":    updatedResume,
	})
}

// writeResumeJSON writes a JSON response.
func writeResumeJSON(
	w http.ResponseWriter,
	status int,
	data interface{},
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// We intentionally ignore encoding errors here because the
	// response headers have already been written.
	writeJSON(w, status, data)
}
