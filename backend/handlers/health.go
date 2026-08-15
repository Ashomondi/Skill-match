package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"skill-match/backend/clients"
	"skill-match/backend/utils"
)

type HealthHandler struct {
	db       databasePinger
	s3Client storagePinger
}

type databasePinger interface {
	Ping(context.Context) error
}

type storagePinger interface {
	Ping(context.Context) error
}

func NewHealthHandler(db *pgxpool.Pool, s3Client *clients.S3Client) *HealthHandler {
	var database databasePinger
	if db != nil {
		database = db
	}
	var storage storagePinger
	if s3Client != nil {
		storage = s3Client
	}
	return &HealthHandler{db: database, s3Client: storage}
}

type dependencyStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type healthResponse struct {
	Status       string                      `json:"status"`
	Dependencies map[string]dependencyStatus `json:"dependencies"`
}

// Health reports overall service status and the reachability of critical
// dependencies. Returns 200 if healthy, 503 if any dependency is down —
// so load balancers/orchestrators can detect and route around failures.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	deps := map[string]dependencyStatus{
		"database": checkDatabase(ctx, h.db),
		"storage":  checkS3(ctx, h.s3Client),
	}

	overallStatus := http.StatusOK
	for _, dep := range deps {
		if dep.Status != "ok" {
			overallStatus = http.StatusServiceUnavailable
			break
		}
	}

	writeHealthJSON(w, overallStatus, healthResponse{
		Status:       statusText(overallStatus),
		Dependencies: deps,
	})
}

func checkDatabase(ctx context.Context, db databasePinger) dependencyStatus {
	if db == nil {
		return dependencyStatus{Status: "not configured"}
	}
	if err := db.Ping(ctx); err != nil {
		return dependencyStatus{Status: "down", Error: "database unreachable"}
	}
	return dependencyStatus{Status: "ok"}
}

func checkS3(ctx context.Context, s3Client storagePinger) dependencyStatus {
	if s3Client == nil {
		return dependencyStatus{Status: "not configured"}
	}
	if err := s3Client.Ping(ctx); err != nil {
		return dependencyStatus{Status: "down", Error: "storage unreachable"}
	}
	return dependencyStatus{Status: "ok"}
}

func statusText(code int) string {
	if code == http.StatusOK {
		return "healthy"
	}
	return "degraded"
}

func writeHealthJSON(w http.ResponseWriter, status int, data healthResponse) {
	utils.WriteJSON(w, status, data)
}
