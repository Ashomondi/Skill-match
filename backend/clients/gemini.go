package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const geminiAPIBase = "https://generativelanguage.googleapis.com/v1beta"

// GeminiError wraps a non-2xx response from the Gemini API so callers can
// classify failures (quota, invalid key, blocked prompt, etc.).
type GeminiError struct {
	Code    int
	Message string
}

func (e *GeminiError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("gemini api error (status %d): %s", e.Code, e.Message)
	}
	return fmt.Sprintf("gemini api error (status %d)", e.Code)
}

// GeminiClient generates model responses via the Gemini generateContent REST
// endpoint using an API key. It implements the generator interface consumed by
// the AI service.
type GeminiClient struct {
	apiKey      string
	modelID     string
	baseURL     string
	httpClient  *http.Client
	maxTokens   int
	temperature float64
}

// NewGeminiClient creates a GeminiClient for the given model. An API key and a
// non-empty model are required.
func NewGeminiClient(apiKey, modelID string) (*GeminiClient, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("gemini API key is required but was empty")
	}
	if strings.TrimSpace(modelID) == "" {
		return nil, fmt.Errorf("gemini model id is required but was empty")
	}
	return &GeminiClient{
		apiKey:      strings.TrimSpace(apiKey),
		modelID:     strings.TrimSpace(modelID),
		baseURL:     geminiAPIBase,
		httpClient:  &http.Client{Timeout: 90 * time.Second},
		maxTokens:   4096,
		temperature: 0.7,
	}, nil
}

// Model returns the configured model identifier.
func (c *GeminiClient) Model() string { return c.modelID }

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
	Config   geminiConfig    `json:"generationConfig"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiConfig struct {
	MaxOutputTokens int     `json:"maxOutputTokens"`
	Temperature     float64 `json:"temperature"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Role  string       `json:"role"`
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// GenerateResponse sends a single prompt and returns the model's text reply.
func (c *GeminiClient) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if strings.TrimSpace(prompt) == "" {
		return "", fmt.Errorf("prompt is required to generate a response")
	}

	body, err := json.Marshal(geminiRequest{
		Contents: []geminiContent{
			{Role: "user", Parts: []geminiPart{{Text: prompt}}},
		},
		Config: geminiConfig{
			MaxOutputTokens: c.maxTokens,
			Temperature:     c.temperature,
		},
	})
	if err != nil {
		return "", fmt.Errorf("marshaling gemini request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, c.modelID, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("building gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling gemini model %s: %w", c.modelID, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("reading gemini response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := classifyHTTPError(resp.StatusCode, raw)
		return "", &GeminiError{Code: resp.StatusCode, Message: message}
	}

	var parsed geminiResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("parsing gemini response: %w", err)
	}

	if parsed.Error != nil {
		return "", &GeminiError{Code: parsed.Error.Code, Message: parsed.Error.Message}
	}

	if len(parsed.Candidates) == 0 {
		return "", fmt.Errorf("gemini returned no candidates")
	}

	var text strings.Builder
	for _, part := range parsed.Candidates[0].Content.Parts {
		text.WriteString(part.Text)
	}
	return text.String(), nil
}

// classifyHTTPError extracts a friendly message from a non-2xx body when it
// carries the standard Gemini error envelope; otherwise falls back to the
// status text.
func classifyHTTPError(status int, raw []byte) string {
	var envelope struct {
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(raw, &envelope) == nil && envelope.Error != nil && envelope.Error.Message != "" {
		return envelope.Error.Message
	}
	return http.StatusText(status)
}

// ClassifyGeminiError returns a user-facing message for a Gemini failure.
func ClassifyGeminiError(err error) string {
	var gErr *GeminiError
	if errors.As(err, &gErr) {
		switch gErr.Code {
		case http.StatusUnauthorized, http.StatusForbidden:
			return "The AI service key is invalid or unauthorized."
		case http.StatusTooManyRequests:
			return "The AI service is busy right now. Please try again in a moment."
		case http.StatusBadRequest:
			return "There was a problem with the AI request."
		case http.StatusNotFound:
			return "The configured AI model is not available."
		case http.StatusServiceUnavailable:
			return "The AI service is temporarily unavailable. Please try again."
		default:
			if gErr.Message != "" {
				return gErr.Message
			}
			return "The AI service encountered an error. Please try again."
		}
	}
	return "Something went wrong. Please try again."
}

// GeminiErrorCode returns a stable code for a Gemini failure, or "unknown".
func GeminiErrorCode(err error) string {
	var gErr *GeminiError
	if errors.As(err, &gErr) {
		return fmt.Sprintf("%d", gErr.Code)
	}
	return "unknown"
}

// ModelGenerator is the subset of AI generation the service layer depends on.
type ModelGenerator interface {
	GenerateResponse(ctx context.Context, prompt string) (string, error)
}
