package models

import "time"

// ConversationRole mirrors the CHECK constraint in migrations/003_memory.sql
// and the role field Bedrock expects on chat messages.
type ConversationRole string

const (
	ConversationRoleUser      ConversationRole = "user"
	ConversationRoleAssistant ConversationRole = "assistant"
	ConversationRoleSystem    ConversationRole = "system"
)

func (r ConversationRole) Valid() bool {
	switch r {
	case ConversationRoleUser, ConversationRoleAssistant, ConversationRoleSystem:
		return true
	default:
		return false
	}
}

// Conversation is a single turn in a user's chat history.
type Conversation struct {
	ID        string
	UserID    string
	Role      ConversationRole
	Content   string
	CreatedAt time.Time
}
