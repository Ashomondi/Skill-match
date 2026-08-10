package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

type SuccessResponse struct {
	Data any `json:"data"`
}

func WriteError(w http.ResponseWriter, err error) {
	appErr, ok := AsAppError(err)

	if !ok {
		log.Printf("[unclassified error] %v", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "Something went wrong. Please try again.",
			Code:  string(CategoryInternal),
		})
		return
	}

	log.Printf("[%s] %v | context=%v", appErr.Category, appErr.Err, appErr.Context)

	writeJSON(w, appErr.StatusCode, ErrorResponse{
		Error: appErr.UserMsg,
		Code:  string(appErr.Category),
	})
}

func WriteSuccess(w http.ResponseWriter, statusCode int, data any) {
	writeJSON(w, statusCode, SuccessResponse{Data: data})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write JSON response: %v", err)
	}
}