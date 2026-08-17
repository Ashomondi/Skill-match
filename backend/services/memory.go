package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"skill-match/backend/models"
	"skill-match/backend/repositories"
)

// Sentinel errors for memory operations.
var (
	ErrMemoryInvalidInput = errors.New("invalid memory input")
	ErrMemoryService      = errors.New("memory service error")
)

// MemoryService handles persistent conversation memory.
type MemoryService struct {
	conversations *repositories.ConversationRepository
}

// NewMemoryService creates a new MemoryService.
func NewMemoryService(
	conversations *repositories.ConversationRepository,
) *MemoryService {
	return &MemoryService{
		conversations: conversations,
	}
}

// StoreConversation stores a conversation turn for a user.
func (s *MemoryService) StoreConversation(
	ctx context.Context,
	conversation *models.Conversation,
) (*models.Conversation, error) {
	if conversation == nil {
		return nil, fmt.Errorf(
			"%w: conversation is required",
			ErrMemoryInvalidInput,
		)
	}

	if strings.TrimSpace(conversation.UserID) == "" {
		return nil, fmt.Errorf(
			"%w: user ID is required",
			ErrMemoryInvalidInput,
		)
	}

	if strings.TrimSpace(conversation.Content) == "" {
		return nil, fmt.Errorf(
			"%w: conversation content is required",
			ErrMemoryInvalidInput,
		)
	}

	if !conversation.Role.Valid() {
		return nil, fmt.Errorf(
			"%w: invalid conversation role",
			ErrMemoryInvalidInput,
		)
	}

	if s.conversations == nil {
		return nil, fmt.Errorf(
			"%w: conversation repository is not configured",
			ErrMemoryService,
		)
	}

	stored, err := s.conversations.Create(
		ctx,
		conversation,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: failed to store conversation: %v",
			ErrMemoryService,
			err,
		)
	}

	return stored, nil
}

// StoreConversationBatch stores multiple conversation turns.
func (s *MemoryService) StoreConversationBatch(
	ctx context.Context,
	conversations []*models.Conversation,
) ([]*models.Conversation, error) {
	if len(conversations) == 0 {
		return nil, fmt.Errorf(
			"%w: conversations are required",
			ErrMemoryInvalidInput,
		)
	}

	if s.conversations == nil {
		return nil, fmt.Errorf(
			"%w: conversation repository is not configured",
			ErrMemoryService,
		)
	}

	for _, conversation := range conversations {
		if conversation == nil {
			return nil, fmt.Errorf(
				"%w: conversation cannot be nil",
				ErrMemoryInvalidInput,
			)
		}

		if strings.TrimSpace(conversation.UserID) == "" {
			return nil, fmt.Errorf(
				"%w: user ID is required",
				ErrMemoryInvalidInput,
			)
		}

		if strings.TrimSpace(conversation.Content) == "" {
			return nil, fmt.Errorf(
				"%w: conversation content is required",
				ErrMemoryInvalidInput,
			)
		}

		if !conversation.Role.Valid() {
			return nil, fmt.Errorf(
				"%w: invalid conversation role",
				ErrMemoryInvalidInput,
			)
		}
	}

	stored, err := s.conversations.CreateBatch(
		ctx,
		conversations,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: failed to store conversation batch: %v",
			ErrMemoryService,
			err,
		)
	}

	return stored, nil
}

// RetrieveMemory retrieves recent conversation history for a user.
func (s *MemoryService) RetrieveMemory(
	ctx context.Context,
	userID string,
	limit int,
) ([]*models.Conversation, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf(
			"%w: user ID is required",
			ErrMemoryInvalidInput,
		)
	}

	if s.conversations == nil {
		return nil, fmt.Errorf(
			"%w: conversation repository is not configured",
			ErrMemoryService,
		)
	}

	history, err := s.conversations.ListRecentByUserID(
		ctx,
		userID,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: failed to retrieve memory: %v",
			ErrMemoryService,
			err,
		)
	}

	return history, nil
}

// DeleteMemory removes all stored conversation memory for a user.
func (s *MemoryService) DeleteMemory(
	ctx context.Context,
	userID string,
) error {
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf(
			"%w: user ID is required",
			ErrMemoryInvalidInput,
		)
	}

	if s.conversations == nil {
		return fmt.Errorf(
			"%w: conversation repository is not configured",
			ErrMemoryService,
		)
	}

	if err := s.conversations.DeleteByUserID(
		ctx,
		userID,
	); err != nil {
		return fmt.Errorf(
			"%w: failed to delete memory: %v",
			ErrMemoryService,
			err,
		)
	}

	return nil
}