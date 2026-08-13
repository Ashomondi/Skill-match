package services

import (
	"context"
	"errors"
	"fmt"
	"log"
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
		return nil, fmt.Errorf(
			"%w: Bedrock client is not configured",
			ErrAIService,
		)
	}

	if s.conversations == nil {
		return nil, fmt.Errorf(
			"%w: conversation repository is not configured",
			ErrAIService,
		)
	}

	// Retrieve recent conversation history.
	history, err := s.conversations.ListRecentByUserID(
		ctx,
		input.UserID,
		20,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: failed to retrieve conversation history: %v",
			ErrAIService,
			err,
		)
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
			return nil, fmt.Errorf(
				"%w: resume repository is not configured",
				ErrAIService,
			)
		}

		resume, err := s.resumes.GetByID(
			ctx,
			input.ResumeID,
		)
		if err != nil {
			if errors.Is(err, repositories.ErrResumeNotFound) {
				return nil, fmt.Errorf(
					"%w: resume not found",
					ErrAIService,
				)
			}

			return nil, fmt.Errorf(
				"%w: failed to retrieve resume: %v",
				ErrAIService,
				err,
			)
		}

		if resume == nil {
			return nil, fmt.Errorf(
				"%w: resume not found",
				ErrAIService,
			)
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
		log.Printf("[ai] bedrock request failed for user=%s: %v", input.UserID, err)
		return nil, &utils.AppError{
			Category:   utils.CategoryInternal,
			UserMsg:    clients.ClassifyBedrockError(err),
			StatusCode: 502,
			Err:        err,
			Context:    map[string]string{"user_id": input.UserID},
		}
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

const maxAIMessageLength = 4000

func validateAIRequest(input AIRequest) error {
	if strings.TrimSpace(input.UserID) == "" {
		return utils.NewValidationError("User ID is required.", nil)
	}
	if strings.TrimSpace(input.Message) == "" {
		return utils.NewValidationError("Please enter a message.", nil)
	}
	if len(strings.TrimSpace(input.Message)) > maxAIMessageLength {
		return utils.NewValidationError(
			fmt.Sprintf("Your message is too long (max %d characters).", maxAIMessageLength),
			nil,
		)
	}
	return nil
}
