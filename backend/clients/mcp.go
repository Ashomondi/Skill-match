package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)


type MCPClient struct {
	httpClient *http.Client
	endpoint   string
	apiKey     string
}

func NewMCPClient(endpoint, apiKey string) (*MCPClient, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("mcp endpoint is required but was empty")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("mcp api key is required but was empty")
	}

	return &MCPClient{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		endpoint:   endpoint,
		apiKey:     apiKey,
	}, nil
}

type mcpRequest struct {
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
}

type mcpResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *mcpError       `json:"error,omitempty"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *MCPClient) Call(ctx context.Context, method string, params map[string]any) (json.RawMessage, error) {
	reqBody, err := json.Marshal(mcpRequest{Method: method, Params: params})
	if err != nil {
		return nil, fmt.Errorf("marshaling mcp request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("building mcp request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mcp request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading mcp response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mcp server returned status %d: %s", resp.StatusCode, string(body))
	}

	var mcpResp mcpResponse
	if err := json.Unmarshal(body, &mcpResp); err != nil {
		return nil, fmt.Errorf("parsing mcp response: %w", err)
	}

	if mcpResp.Error != nil {
		return nil, fmt.Errorf("mcp error %d: %s", mcpResp.Error.Code, mcpResp.Error.Message)
	}

	return mcpResp.Result, nil
}


func ClassifyMCPError(err error) string {
	if err == nil {
		return ""
	}
	return "We couldn't reach the memory service right now. Please try again shortly."
}


type MCPCaller interface {
	Call(ctx context.Context, method string, params map[string]any) (json.RawMessage, error)
}