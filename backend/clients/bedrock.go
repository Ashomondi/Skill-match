package clients

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
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