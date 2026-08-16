package routes

import (
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"skill-match/backend/clients"
	"skill-match/backend/handlers"
)

type RegisterFunc func(mux *http.ServeMux)

func RegisterAll(mux *http.ServeMux, registrars ...RegisterFunc) {
	for _, register := range registrars {
		register(mux)
	}
}

// RegisterHealth exposes GET /health. When a CockroachDB pool is provided it
// is pinged; a missing or unreachable database reports 503 (degraded).
func RegisterHealth(mux *http.ServeMux, pool *pgxpool.Pool) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if pool == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"degraded","database":"unavailable"}`))
			return
		}

		if err := clients.HealthCheck(r.Context(), pool); err != nil {
			log.Printf("health check failed: %v", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"degraded","database":"unavailable"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","database":"ok"}`))
	})
}

// RegisterAuth exposes the public authentication endpoints.
func RegisterAuth(mux *http.ServeMux, h *handlers.AuthHandler) {
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)
}
