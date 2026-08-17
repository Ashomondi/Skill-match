package models

import "time"

// ConversationRole represents the role of a conversation message.
type ConversationRole string

const (
	ConversationRoleUser      ConversationRole = "user"
	ConversationRoleAssistant ConversationRole = "assistant"
	ConversationRoleSystem    ConversationRole = "system"
)

// Valid checks whether the conversation role is supported.
func (r ConversationRole) Valid() bool {
	switch r {
	case ConversationRoleUser,
		ConversationRoleAssistant,
		ConversationRoleSystem:
		return true
	default:
		return false
	}
}

type Conversation struct {
	ID        string
	UserID    string
	Role      ConversationRole
	Content   string
	CreatedAt time.Time
}
