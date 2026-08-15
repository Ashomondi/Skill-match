package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"skill-match/backend/utils"
)

func TestRecoveryReturnsSafeJSON(t *testing.T) {
	handler := Recovery(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("database password=secret")
	}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/test", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if recorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("content type = %q", recorder.Header().Get("Content-Type"))
	}
	var body utils.ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != string(utils.CategoryInternal) {
		t.Fatalf("code = %q, want %q", body.Error.Code, utils.CategoryInternal)
	}
}
