package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"skill-match/backend/models"
)

// Sentinel errors for chat operations.
var (
	ErrChatInvalidInput = errors.New("invalid chat input")
	ErrChatService      = errors.New("chat service error")
)

// ChatService coordinates the chat workflow between memory and AI.
type ChatService struct {
	ai     *AIService
	memory *MemoryService
}

// NewChatService creates a new ChatService.
func NewChatService(
	ai *AIService,
	memory *MemoryService,
) *ChatService {
	return &ChatService{
		ai:     ai,
		memory: memory,
	}
}

// ChatRequest contains the information required to send a chat message.
type ChatRequest struct {
	UserID   string
	Message  string
	ResumeID string
}

// ChatResponse contains the assistant's response.
type ChatResponse struct {
	Message string `json:"message"`
}

// SendMessage stores the user's message, invokes the AI service,
// stores the assistant response, and returns it.
func (s *ChatService) SendMessage(
	ctx context.Context,
	input ChatRequest,
) (*ChatResponse, error) {

	if strings.TrimSpace(input.UserID) == "" {
		return nil, fmt.Errorf(
			"%w: user ID is required",
			ErrChatInvalidInput,
		)
	}

	if strings.TrimSpace(input.Message) == "" {
		return nil, fmt.Errorf(
			"%w: message is required",
			ErrChatInvalidInput,
		)
	}

	if s.ai == nil {
		return nil, fmt.Errorf(
			"%w: AI service is not configured",
			ErrChatService,
		)
	}

	if s.memory == nil {
		return nil, fmt.Errorf(
			"%w: memory service is not configured",
			ErrChatService,
		)
	}

	// Store the user's message.
	_, err := s.memory.StoreConversation(
		ctx,
		&models.Conversation{
			UserID:  input.UserID,
			Role:    models.ConversationRoleUser,
			Content: strings.TrimSpace(input.Message),
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: failed to store user message: %w",
			ErrChatService,
			err,
		)
	}

	// Generate the AI response.
	aiResponse, err := s.ai.GenerateResponse(
		ctx,
		AIRequest{
			UserID:   input.UserID,
			Message:  input.Message,
			ResumeID: input.ResumeID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: failed to generate AI response: %w",
			ErrChatService,
			err,
		)
	}

	if aiResponse == nil ||
		strings.TrimSpace(aiResponse.Message) == "" {
		return nil, fmt.Errorf(
			"%w: AI returned an empty response",
			ErrChatService,
		)
	}

	// Store the assistant response.
	_, err = s.memory.StoreConversation(
		ctx,
		&models.Conversation{
			UserID:  input.UserID,
			Role:    models.ConversationRoleAssistant,
			Content: strings.TrimSpace(aiResponse.Message),
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: failed to store assistant response: %w",
			ErrChatService,
			err,
		)
	}

	return &ChatResponse{
		Message: strings.TrimSpace(aiResponse.Message),
	}, nil
}

// buildChatPrompt builds the context sent to Amazon Bedrock.
func buildChatPrompt(
	message string,
	history []*models.Conversation,
) string {

	var builder strings.Builder

	builder.WriteString(
		"You are Skill-match, an AI assistant that helps users with job searching, resumes and career-related questions.\n\n",
	)

	if len(history) > 0 {
		builder.WriteString("Previous conversation:\n")

		for _, turn := range history {
			if turn == nil {
				continue
			}

			builder.WriteString(string(turn.Role))
			builder.WriteString(": ")
			builder.WriteString(strings.TrimSpace(turn.Content))
			builder.WriteString("\n")
		}

		builder.WriteString("\n")
	}

	builder.WriteString("Current user message:\n")
	builder.WriteString(strings.TrimSpace(message))
	builder.WriteString(
		"\n\nProvide a helpful, accurate and concise response.",
	)

	return builder.String()
}
