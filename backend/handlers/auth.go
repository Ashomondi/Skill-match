package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"skill-match/backend/models"
	"skill-match/backend/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"fullName"`
}

type AuthResponse struct {
	Message string       `json:"message"`
	User    *models.User `json:"user,omitempty"`
	Token   string       `json:"token,omitempty"`
}

// NewAuthHandler creates a new authentication handler.
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles POST /api/auth/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var request AuthRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	user, token, err := h.authService.Register(
		r.Context(),
		request.Email,
		request.Password,
		request.FullName,
	)

	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidEmail):
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid email address",
			})

		case errors.Is(err, services.ErrInvalidPassword):
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "password must be at least 8 characters",
			})

		case errors.Is(err, services.ErrUserAlreadyExists):
			writeJSON(w, http.StatusConflict, map[string]string{
				"error": "user already exists",
			})

		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to create account",
			})
		}

		return
	}

	writeJSON(w, http.StatusCreated, AuthResponse{
		Message: "account created successfully",
		User:    user,
		Token:   token,
	})
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request AuthRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	user, token, err := h.authService.Login(
		r.Context(),
		request.Email,
		request.Password,
	)

	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "invalid email or password",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "login failed",
		})
		return
	}

	writeJSON(w, http.StatusOK, AuthResponse{
		Message: "login successful",
		User:    user,
		Token:   token,
	})
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	data interface{},
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return
	}
}
