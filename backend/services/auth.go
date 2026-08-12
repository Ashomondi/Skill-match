package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"skill-match/backend/repositories"
	"skill-match/backend/utils"

	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("password must be at least 8 characters")
)

// UserRepository defines the database operations required by authentication.
type UserRepository interface {
	Create(ctx context.Context, user *repositories.User) (*repositories.User, error)
	GetByEmail(ctx context.Context, email string) (*repositories.User, error)
}

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo   UserRepository
	jwtManager *utils.JWTManager
}

// NewAuthService creates a new authentication service.
func NewAuthService(
	userRepo UserRepository,
	jwtManager *utils.JWTManager,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register creates a new user account and returns the user plus an
// authenticated session token.
func (s *AuthService) Register(
	ctx context.Context,
	email string,
	password string,
	fullName string,
) (*repositories.User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	if !strings.Contains(email, "@") {
		return nil, "", ErrInvalidEmail
	}

	if len(password) < 8 {
		return nil, "", ErrInvalidPassword
	}

	// Check whether the email is already registered.
	existingUser, err := s.userRepo.GetByEmail(ctx, email)

	if err == nil && existingUser != nil {
		return nil, "", ErrUserAlreadyExists
	}

	// Hash the password before storing it.
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	now := time.Now()

	created, err := s.userRepo.Create(ctx, &repositories.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: hashedPassword,
		FullName:     strings.TrimSpace(fullName),
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return nil, "", err
	}

	token, err := s.jwtManager.GenerateToken(created.ID, created.Email)
	if err != nil {
		return nil, "", err
	}

	// Never return the password hash.
	created.PasswordHash = ""

	return created, token, nil
}

// Login authenticates an existing user and returns the user plus a JWT.
func (s *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
) (*repositories.User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	if user == nil {
		return nil, "", ErrInvalidCredentials
	}

	// Compare the supplied password against the stored bcrypt hash.
	if !utils.CheckPassword(password, user.PasswordHash) {
		return nil, "", ErrInvalidCredentials
	}

	token, err := s.jwtManager.GenerateToken(
		user.ID,
		user.Email,
	)
	if err != nil {
		return nil, "", err
	}

	// Never return the password hash.
	user.PasswordHash = ""

	return user, token, nil
}
