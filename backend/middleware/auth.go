package middleware

import "net/http"

// Auth is a placeholder until Issue 3 (JWT) lands. Currently a no-op passthrough
// so routes can be registered as "protected" now and gain real enforcement later
// without changing call sites.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}