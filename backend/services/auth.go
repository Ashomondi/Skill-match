package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"skill-match/backend/models"
	"skill-match/backend/utils"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("password must be at least 8 characters")
)

// UserRepository defines the database operations required by authentication.
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo  UserRepository
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

// Register creates a new user account.
func (s *AuthService) Register(
	ctx context.Context,
	email string,
	password string,
) (*models.User, error) {

	email = strings.ToLower(strings.TrimSpace(email))

	if !strings.Contains(email, "@") {
		return nil, ErrInvalidEmail
	}

	if len(password) < 8 {
		return nil, ErrInvalidPassword
	}

	// Check whether the email is already registered.
	existingUser, err := s.userRepo.GetByEmail(ctx, email)

	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash the password before storing it.
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	user := &models.User{
		ID:        generateUserID(),
		Email:     email,
		Password:  hashedPassword,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Never return the password hash.
	user.Password = ""

	return user, nil
}

// Login authenticates an existing user and returns a JWT.
func (s *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
) (string, error) {

	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if user == nil {
		return "", ErrInvalidCredentials
	}

	// Compare the supplied password against the stored bcrypt hash.
	if !utils.CheckPassword(password, user.Password) {
		return "", ErrInvalidCredentials
	}

	token, err := s.jwtManager.GenerateToken(
		user.ID,
		user.Email,
	)

	if err != nil {
		return "", err
	}

	return token, nil
}

// generateUserID creates a simple unique user identifier.
func generateUserID() string {
	return time.Now().UTC().Format("20060102150405.000000000")
}