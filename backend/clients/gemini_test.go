package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewGeminiClientValidates(t *testing.T) {
	if _, err := NewGeminiClient("", "gemini-2.0-flash"); err == nil {
		t.Fatal("expected an error when the API key is empty")
	}
	if _, err := NewGeminiClient("key", ""); err == nil {
		t.Fatal("expected an error when the model is empty")
	}
}

func TestGeminiClientGenerateResponse(t *testing.T) {
	var gotBody struct {
		Contents []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"contents"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") != "secret-key" {
			t.Errorf("expected api key in query, got %q", r.URL.Query().Get("key"))
		}
		if !strings.HasSuffix(r.URL.Path, "/models/gemini-2.0-flash:generateContent") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"candidates": []map[string]any{
				{"content": map[string]any{
					"role":  "model",
					"parts": []map[string]any{{"text": "Tailored CV content"}},
				}},
			},
		})
	}))
	defer server.Close()

	client, err := NewGeminiClient("secret-key", "gemini-2.0-flash")
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	client.baseURL = server.URL
	client.httpClient = server.Client()

	text, err := client.GenerateResponse(context.Background(), "tailor this CV")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if text != "Tailored CV content" {
		t.Fatalf("unexpected reply: %q", text)
	}
	if len(gotBody.Contents) != 1 || len(gotBody.Contents[0].Parts) != 1 {
		t.Fatal("expected a single content part in the request")
	}
	if gotBody.Contents[0].Parts[0].Text != "tailor this CV" {
		t.Fatalf("unexpected prompt: %q", gotBody.Contents[0].Parts[0].Text)
	}
}

func TestGeminiClientSurfacesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": 429, "message": "quota exceeded", "status": "RESOURCE_EXHAUSTED"},
		})
	}))
	defer server.Close()

	client, err := NewGeminiClient("secret-key", "gemini-2.0-flash")
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	client.baseURL = server.URL
	client.httpClient = server.Client()

	_, err = client.GenerateResponse(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected an error from the 429 response")
	}
	msg := ClassifyGeminiError(err)
	if !strings.Contains(msg, "busy") {
		t.Fatalf("expected a busy-friendly message, got %q", msg)
	}
	if code := GeminiErrorCode(err); code != "429" {
		t.Fatalf("expected error code 429, got %q", code)
	}
}
