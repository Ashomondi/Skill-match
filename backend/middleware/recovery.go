package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"skill-match/backend/utils"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("panic recovered",
					"request_id", utils.RequestID(r.Context()),
					"method", r.Method,
					"path", r.URL.Path,
					"panic_type", "unexpected_panic",
					"stack", string(debug.Stack()),
				)
				utils.WriteJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: utils.ErrorBody{
					Message: "Something went wrong on our end. Please try again.",
					Code:    string(utils.CategoryInternal),
				}})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
