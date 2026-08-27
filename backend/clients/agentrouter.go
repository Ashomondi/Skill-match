package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type AgentRouterClient struct {
	client                 *http.Client
	baseURL, apiKey, model string
}

func NewAgentRouterClient(baseURL, apiKey, model string) (*AgentRouterClient, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("agentrouter API key is required")
	}
	if strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("agentrouter model is required")
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://agentrouter.org/v1"
	}
	return &AgentRouterClient{client: &http.Client{Timeout: 60 * time.Second}, baseURL: strings.TrimRight(baseURL, "/"), apiKey: apiKey, model: model}, nil
}

func (c *AgentRouterClient) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	payload, err := json.Marshal(map[string]any{"model": c.model, "messages": []map[string]string{{"role": "user", "content": prompt}}})
	if err != nil {
		return "", fmt.Errorf("marshal AgentRouter request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create AgentRouter request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call AgentRouter: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 2<<20))
	if err != nil {
		return "", fmt.Errorf("read AgentRouter response: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("AgentRouter returned HTTP %d", res.StatusCode)
	}
	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		return "", fmt.Errorf("decode AgentRouter response: %w", err)
	}
	if len(decoded.Choices) == 0 || strings.TrimSpace(decoded.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("AgentRouter returned an empty response")
	}
	return decoded.Choices[0].Message.Content, nil
}
