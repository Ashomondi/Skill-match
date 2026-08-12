package services

import (
	"skill-match/backend/repositories"
	"strings"
)

func buildChatPrompt(
	message string,
	history []*repositories.Conversation,
) string {
	var builder strings.Builder
	builder.WriteString(
		"You are Skill-match, an AI assistant that helps users with job searching, resumes and career-related questions.\n\n",
	)
	if len(history) > 0 {
		builder.WriteString("previous conversation:\n")
		for _, turn := range history {
			if turn == nil {
				continue
			}
			builder.WriteString(
				string(turn.Role),
			)
			builder.WriteString(": ")
			builder.WriteString(
				strings.TrimSpace(turn.Content),
			)
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
