package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"

	"skill-match/backend/utils"
)

const requestIDHeader = "X-Request-ID"

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(data)
	r.bytes += n
	return n, err
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := r.Header.Get(requestIDHeader)
		if _, err := uuid.Parse(requestID); err != nil {
			requestID = uuid.NewString()
		}
		w.Header().Set(requestIDHeader, requestID)
		r = r.WithContext(utils.WithRequestID(r.Context(), requestID))
		recorder := &responseRecorder{ResponseWriter: w}

		next.ServeHTTP(recorder, r)
		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		slog.Info("request completed",
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"bytes", recorder.bytes,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

type failureTracker struct {
	mu      sync.Mutex
	counts  map[string][]time.Time
	window  time.Duration
	alertAt int
}

var tracker = &failureTracker{counts: make(map[string][]time.Time), window: 5 * time.Minute, alertAt: 5}

func RecordFailure(category string) {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-tracker.window)
	recent := tracker.counts[category][:0]
	for _, recordedAt := range tracker.counts[category] {
		if recordedAt.After(cutoff) {
			recent = append(recent, recordedAt)
		}
	}
	recent = append(recent, now)
	tracker.counts[category] = recent
	if len(recent) == tracker.alertAt {
		slog.Warn("repeated failures detected", "category", category, "count", len(recent), "window", tracker.window.String())
	}
}
