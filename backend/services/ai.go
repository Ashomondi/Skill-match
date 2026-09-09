package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"skill-match/backend/clients"
	"skill-match/backend/repositories"
	"skill-match/backend/utils"
)

var (
	ErrAIInvalidInput = errors.New("invalid AI input")
	ErrAIService      = errors.New("AI service error")
)

type AIService struct {
	generator     clients.ModelGenerator
	conversations *repositories.ConversationRepository
	resumes       *repositories.ResumeRepository
}

type NewAIServiceInput struct {
	Generator     clients.ModelGenerator
	Conversations *repositories.ConversationRepository
	Resumes       *repositories.ResumeRepository
}

func NewAIService(input NewAIServiceInput) *AIService {
	return &AIService{
		generator:     input.Generator,
		conversations: input.Conversations,
		resumes:       input.Resumes,
	}
}

type AIRequest struct {
	UserID   string
	Message  string
	ResumeID string
}

type AIResponse struct {
	Message string
}

func (s *AIService) GenerateResponse(
	ctx context.Context,
	input AIRequest,
) (*AIResponse, error) {
	if err := validateAIRequest(input); err != nil {
		return nil, err
	}

	if s.generator == nil {
		return nil, utils.NewInternalError(ErrAIService, map[string]string{"operation": "generate_ai_response", "service": "generator"})
	}

	if s.conversations == nil {
		return nil, utils.NewInternalError(ErrAIService, map[string]string{"operation": "list_conversations", "resource": "conversation"})
	}

	// Retrieve recent conversation history.
	history, err := s.conversations.ListRecentByUserID(
		ctx,
		input.UserID,
		20,
	)
	if err != nil {
		return nil, utils.NewDatabaseError(err, map[string]string{"operation": "list_conversations", "resource": "conversation", "user_id": input.UserID})
	}

	// Build prompt using user message and conversation history.
	prompt := buildChatPrompt(
		input.Message,
		history,
	)

	// Add resume context when a resume ID was supplied.
	if strings.TrimSpace(input.ResumeID) != "" {
		if s.resumes == nil {
			return nil, utils.NewInternalError(ErrAIService, map[string]string{"operation": "get_resume", "resource": "resume"})
		}

		resume, err := s.resumes.GetByID(
			ctx,
			input.ResumeID,
		)
		if err != nil {
			if errors.Is(err, repositories.ErrResumeNotFound) {
				return nil, utils.NewNotFoundError("Resume not found.")
			}

			return nil, utils.NewDatabaseError(err, map[string]string{"operation": "get_resume", "resource": "resume", "resume_id": input.ResumeID})
		}

		if resume == nil {
			return nil, utils.NewNotFoundError("Resume not found.")
		}

		// Prevent user from using another user's resume as context.
		if resume.UserID != input.UserID {
			return nil, ErrResumeUnauthorized
		}

		// Only include extracted text when available.
		if resume.ParsedText != nil &&
			strings.TrimSpace(*resume.ParsedText) != "" {

			prompt += "\n\nResume context:\n"
			prompt += strings.TrimSpace(*resume.ParsedText)
		}
	}

	// Send the assembled context to the AI model.
	response, err := s.generator.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, utils.NewUpstreamError(clients.ClassifyGeminiError(err), err, map[string]string{
			"operation": "generate_content", "service": "gemini", "error_code": clients.GeminiErrorCode(err), "user_id": input.UserID,
		})
	}
	response = strings.TrimSpace(response)

	if response == "" {
		return nil, fmt.Errorf(
			"%w: the model returned an empty response",
			ErrAIService,
		)
	}

	return &AIResponse{
		Message: response,
	}, nil
}

func validateAIRequest(req AIRequest) error {
	if strings.TrimSpace(req.UserID) == "" {
		return fmt.Errorf("%w: user ID is required", ErrAIInvalidInput)
	}
	if strings.TrimSpace(req.Message) == "" {
		return fmt.Errorf("%w: message is required", ErrAIInvalidInput)
	}
	return nil
}
