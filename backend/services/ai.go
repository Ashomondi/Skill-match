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
	ErrAIInvalidInput     = errors.New("invalid AI input")
	ErrAIService          = errors.New("AI service error")
)

type AIService struct {
	bedrock       clients.BedrockGenerator
	conversations *repositories.ConversationRepository
	resumes       *repositories.ResumeRepository
}

type NewAIServiceInput struct {
	Bedrock       clients.BedrockGenerator
	Conversations *repositories.ConversationRepository
	Resumes       *repositories.ResumeRepository
}

func NewAIService(input NewAIServiceInput) *AIService {
	return &AIService{
		bedrock:       input.Bedrock,
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

	if s.bedrock == nil {
		return nil, utils.NewInternalError(ErrAIService, map[string]string{"operation": "generate_ai_response", "service": "bedrock"})
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

	// Build the initial prompt using the user's message
	// and previous conversation history.
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

		// Prevent a user from using another user's resume
		// as AI context.
		if resume.UserID != input.UserID {
			return nil, ErrResumeUnauthorized
		}

		// Only include extracted resume text when it exists.
		if resume.ParsedText != nil &&
			strings.TrimSpace(*resume.ParsedText) != "" {

			prompt += "\n\nResume context:\n"
			prompt += strings.TrimSpace(*resume.ParsedText)
		}
	}

	// Send the complete context to Amazon Bedrock.
	response, err := s.bedrock.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, utils.NewUpstreamError(clients.ClassifyBedrockError(err), err, map[string]string{
			"operation": "invoke_model", "service": "bedrock", "error_code": clients.BedrockErrorCode(err), "user_id": input.UserID,
		})
	}
	response = strings.TrimSpace(response)

	if response == "" {
		return nil, fmt.Errorf(
			"%w: Bedrock returned an empty response",
			ErrAIService,
		)
	}

	return &AIResponse{
		Message: response,
	}, nil
}