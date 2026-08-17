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

func RegisterHealth(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", healthHandler)
}

func RegisterAuth(mux *http.ServeMux, h *handlers.AuthHandler) {
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)
}

func RegisterResume(mux *http.ServeMux, h *handlers.ResumeHandler) {
	mux.HandleFunc("POST /api/resumes", h.Upload)
	mux.HandleFunc("PUT /api/resumes/{id}", h.Update)
	mux.HandleFunc("DELETE /api/resumes/{id}", h.Delete)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func RegisterChat(mux *http.ServeMux, h *handlers.ChatHandler) {
	mux.HandleFunc("POST /api/chat", h.Chat)
}

func RegisterJobs(mux *http.ServeMux, h *handlers.JobsHandler, jwt *utils.JWTManager) {
	mux.Handle(
		"GET /api/jobs/search",
		middleware.Auth(jwt)(
			http.HandlerFunc(h.Search),
		),
	)
}