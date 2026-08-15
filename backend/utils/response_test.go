package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteRequestErrorHidesInternalCause(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/resumes", nil)

	WriteRequestError(recorder, request, NewDatabaseError(errors.New("password=secret SELECT * FROM users"), map[string]string{"operation": "create_resume"}))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if got := recorder.Body.String(); strings.Contains(got, "secret") || strings.Contains(got, "SELECT") {
		t.Fatalf("response exposed internal cause: %s", got)
	}
	var body ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != string(CategoryDatabase) {
		t.Fatalf("code = %q, want %q", body.Error.Code, CategoryDatabase)
	}
}
