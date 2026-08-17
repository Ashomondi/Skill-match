package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggingCapturesStatusAndRequestID(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	handler := Logging(Recovery(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("unexpected")
	})))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/test", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if recorder.Header().Get(requestIDHeader) == "" {
		t.Fatal("missing response request ID")
	}
	if output := logs.String(); !strings.Contains(output, `"status":500`) || !strings.Contains(output, `"request_id"`) {
		t.Fatalf("log missing request context: %s", output)
	}
}
