package utils

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

type SuccessResponse struct {
	Data any `json:"data"`
}

func WriteError(w http.ResponseWriter, err error) {
	WriteRequestError(w, nil, err)
}

func WriteRequestError(w http.ResponseWriter, r *http.Request, err error) {
	appErr, ok := AsAppError(err)
	if !ok {
		appErr = NewInternalError(err, nil)
	}

	attrs := []any{"category", appErr.Category, "status", appErr.StatusCode}
	if appErr.Err != nil {
		attrs = append(attrs, "error_type", fmt.Sprintf("%T", appErr.Err))
	}
	if r != nil {
		attrs = append(attrs, "request_id", RequestID(r.Context()), "method", r.Method, "path", r.URL.Path)
	}
	for key, value := range appErr.Context {
		if safeLogKey(key) {
			attrs = append(attrs, key, value)
		}
	}
	// Error text is deliberately excluded: database/AWS errors may contain
	// queries, object keys, endpoints, or other sensitive implementation data.
	slog.Error("request failed", attrs...)

	WriteJSON(w, appErr.StatusCode, ErrorResponse{Error: ErrorBody{
		Message: appErr.UserMsg,
		Code:    string(appErr.Category),
	}})
}

func WriteSuccess(w http.ResponseWriter, statusCode int, data any) {
	WriteJSON(w, statusCode, SuccessResponse{Data: data})
}

func WriteJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("failed to encode JSON response", "error_type", "json_encode")
	}
}

func safeLogKey(key string) bool {
	switch key {
	case "operation", "resource", "service", "error_code", "user_id", "resume_id":
		return true
	default:
		return false
	}
}

