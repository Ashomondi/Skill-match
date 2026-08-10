package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"
)

// Logging runs on every request and logs method, path, and duration.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// failureTracker counts recent failures per category to detect repeated
// failures worth alerting on, without needing an external monitoring service.
type failureTracker struct {
	mu      sync.Mutex
	counts  map[string][]time.Time
	window  time.Duration
	alertAt int
}

var tracker = &failureTracker{
	counts:  make(map[string][]time.Time),
	window:  5 * time.Minute,
	alertAt: 5,
}
func RecordFailure(category string) {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-tracker.window)

	recent := tracker.counts[category][:0]
	for _, t := range tracker.counts[category] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	recent = append(recent, now)
	tracker.counts[category] = recent

	if len(recent) == tracker.alertAt {
		log.Printf("ALERT: %d %q failures in the last %s — investigate immediately", len(recent), category, tracker.window)
	}
}