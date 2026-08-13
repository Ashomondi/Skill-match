package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"skill-match/backend/repositories"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/smithy-go"
)

type BedrockClient struct {
	client  *bedrockruntime.Client
	modelID string
}

func NewBedrockClient(ctx context.Context, region, modelID string) (*BedrockClient, error) {
	if modelID == "" {
		return nil, fmt.Errorf("bedrock model id is required but was empty")
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	return &BedrockClient{
		client:  bedrockruntime.NewFromConfig(cfg),
		modelID: modelID,
	}, nil
}

type modelRequest struct {
	AnthropicVersion string    `json:"anthropic_version"`
	MaxTokens        int       `json:"max_tokens"`
	Messages         []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type modelResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

type embedRequest struct {
	InputText string `json:"inputText"`
}

type embedResponse struct {
	Embedding []float32 `json:"embedding"`
}

func (c *BedrockClient) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	body := modelRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        1024,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshaling bedrock request: %w", err)
	}

	result, err := c.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(c.modelID),
		ContentType: aws.String("application/json"),
		Body:        payload,
	})
	if err != nil {
		return "", fmt.Errorf("invoking bedrock model %s: %w", c.modelID, err)
	}

	var response modelResponse
	if err := json.Unmarshal(result.Body, &response); err != nil {
		return "", fmt.Errorf("parsing bedrock response: %w", err)
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("bedrock returned an empty response")
	}

	return response.Content[0].Text, nil
}

type BedrockGenerator interface {
	GenerateResponse(ctx context.Context, prompt string) (string, error)
}

func ClassifyBedrockError(err error) string {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "ThrottlingException":
			return "The AI service is busy right now. Please try again in a moment."
		case "ModelTimeoutException":
			return "The AI took too long to respond. Please try again."
		case "ValidationException":
			return "There was a problem with the request format."
		case "AccessDeniedException":
			return "AI service access is not configured correctly."
		default:
			return "The AI service encountered an error. Please try again."
		}
	}
	return "Something went wrong. Please try again."
}

func (c *BedrockClient) GenerateEmbedding(ctx context.Context, embedModelID, text string) ([]float32, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("text is required to generate an embedding")
	}

	body := embedRequest{InputText: text}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling embed request: %w", err)
	}

	result, err := c.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(embedModelID),
		ContentType: aws.String("application/json"),
		Body:        payload,
	})
	if err != nil {
		return nil, fmt.Errorf("invoking bedrock embedding model %s: %w", embedModelID, err)
	}

	var response embedResponse
	if err := json.Unmarshal(result.Body, &response); err != nil {
		return nil, fmt.Errorf("parsing embedding response: %w", err)
	}

	if len(response.Embedding) != repositories.EmbeddingDim {
		return nil, fmt.Errorf("unexpected embedding dimension: got %d, want %d", len(response.Embedding), repositories.EmbeddingDim)
	}

	return response.Embedding, nil
}
