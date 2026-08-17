package routes

import (
	"net/http"

	"skill-match/backend/handlers"
	"skill-match/backend/middleware"
	"skill-match/backend/utils"
)

type RegisterFunc func(mux *http.ServeMux)

func RegisterAll(mux *http.ServeMux, registrars ...RegisterFunc) {
	for _, register := range registrars {
		register(mux)
	}
}

func RegisterHealth(mux *http.ServeMux, healthHandler *handlers.HealthHandler) {
	mux.HandleFunc("GET /health", healthHandler.Health)
}

func RegisterAuth(mux *http.ServeMux, h *handlers.AuthHandler) {
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)
}

// RegisterResumes exposes the resume endpoints, all protected by auth.
func RegisterResumes(mux *http.ServeMux, h *handlers.ResumeHandler, jwt *utils.JWTManager) {
	auth := middleware.Auth(jwt)
	mux.Handle("GET /api/resumes", auth(http.HandlerFunc(h.List)))
	mux.Handle("POST /api/resumes", auth(http.HandlerFunc(h.Create)))
	mux.Handle("GET /api/resumes/{id}", auth(http.HandlerFunc(h.Get)))
	mux.Handle("DELETE /api/resumes/{id}", auth(http.HandlerFunc(h.Delete)))
}

func RegisterChat(mux *http.ServeMux, h *handlers.ChatHandler) {
	mux.HandleFunc("POST /api/chat", h.Chat)
}

func RegisterSavedJobs(mux *http.ServeMux, h *handlers.SavedJobsHandler, jwt *utils.JWTManager) {
	mux.Handle("POST /api/saved-jobs", middleware.Auth(jwt)(http.HandlerFunc(h.Save)))
	mux.Handle("GET /api/saved-jobs", middleware.Auth(jwt)(http.HandlerFunc(h.List)))
	mux.Handle("DELETE /api/saved-jobs/{job_id}", middleware.Auth(jwt)(http.HandlerFunc(h.Remove)))
}

// RegisterApplications exposes the application endpoints, all protected by auth.
func RegisterApplications(mux *http.ServeMux, h *handlers.ApplicationHandler, jwt *utils.JWTManager) {
	mux.Handle("GET /api/applications", middleware.Auth(jwt)(http.HandlerFunc(h.List)))
}
