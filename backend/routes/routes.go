package routes

import (
	"net/http"

	"skill-match/backend/handlers"
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

func RegisterResumeRoutes(
	mux *http.ServeMux,
	resumeHandler *handlers.ResumeHandler,
) {
	mux.HandleFunc("POST /api/resumes", resumeHandler.Upload)
	mux.HandleFunc("GET /api/resumes", resumeHandler.List)
	mux.HandleFunc("GET /api/resumes/{id}", resumeHandler.Get)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}