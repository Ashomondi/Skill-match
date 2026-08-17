package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"skill-match/backend/middleware"
	"skill-match/backend/services"
	"skill-match/backend/utils"
)

// ChatHandler handles authenticated chat requests.
type ChatHandler struct {
	chatService *services.ChatService
}

// NewChatHandler creates a new ChatHandler.
func NewChatHandler(
	chatService *services.ChatService,
) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

type chatRequest struct {
	Message  string `json:"message"`
	ResumeID string `json:"resumeId,omitempty"`
}

type chatResponse struct {
	Message string `json:"message"`
}

// Chat handles POST /api/chat.
func (h *ChatHandler) Chat(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		writeJSON(
			w,
			http.StatusMethodNotAllowed,
			map[string]string{
				"error": "method not allowed",
			},
		)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeJSON(
			w,
			http.StatusUnauthorized,
			map[string]string{
				"error": "authentication required",
			},
		)
		return
	}

	var request chatRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "invalid request body",
			},
		)
		return
	}

	if strings.TrimSpace(request.Message) == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "message is required",
			},
		)
		return
	}

	response, err := h.chatService.SendMessage(
		r.Context(),
		services.ChatRequest{
			UserID:   userID,
			Message:  request.Message,
			ResumeID: request.ResumeID,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrChatInvalidInput):
			writeJSON(
				w,
				http.StatusBadRequest,
				map[string]string{
					"error": "invalid chat request",
				},
			)

		case errors.Is(err, services.ErrAIInvalidInput):
			writeJSON(
				w,
				http.StatusBadRequest,
				map[string]string{
					"error": "invalid AI request",
				},
			)

		default:
			utils.WriteRequestError(w, r, err)
		}

		return
	}

	writeJSON(
		w,
		http.StatusOK,
		chatResponse{
			Message: response.Message,
		},
	)
}
