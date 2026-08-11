package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMCPClient_Call(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("missing or wrong auth header")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mcpResponse{Result: json.RawMessage(`{"ok":true}`)})
	}))
	defer server.Close()

	client, err := NewMCPClient(server.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	result, err := client.Call(context.Background(), "test.method", nil)
	if err != nil {
		t.Fatalf("unexpected error calling mcp: %v", err)
	}
	if string(result) != `{"ok":true}` {
		t.Errorf("unexpected result: %s", result)
	}
}