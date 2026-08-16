package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"skill-match/backend/models"
	"skill-match/backend/repositories"
	"skill-match/backend/utils"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("password must be at least 8 characters")
)

// UserRepository defines the database operations required by authentication.
// repositories.UserRepository satisfies this interface.
type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
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

// Register creates a new user account and returns the persisted user and a
// freshly issued JWT (auto-login after signup). The returned user never
// carries the password hash.
func (s *AuthService) Register(
	ctx context.Context,
	email string,
	password string,
	fullName string,
) (*models.User, string, error) {

	email = strings.ToLower(strings.TrimSpace(email))

	if !strings.Contains(email, "@") {
		return nil, "", ErrInvalidEmail
	}

	if len(password) < 8 {
		return nil, "", ErrInvalidPassword
	}

	// Check whether the email is already registered.
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return nil, "", err
	}
	if err == nil && existingUser != nil {
		return nil, "", ErrUserAlreadyExists
	}

	// Hash the password before storing it.
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	now := time.Now()

	user := &models.User{
		Email:     email,
		Password:  hashedPassword,
		FullName:  strings.TrimSpace(fullName),
		CreatedAt: now,
		UpdatedAt: now,
	}

	created, err := s.userRepo.Create(ctx, user)
	if err != nil {
		if errors.Is(err, repositories.ErrUserEmailTaken) {
			return nil, "", ErrUserAlreadyExists
		}
		return nil, "", err
	}

	token, err := s.jwtManager.GenerateToken(created.ID, created.Email)
	if err != nil {
		return nil, "", err
	}

	// Never return the password hash.
	created.Password = ""

	return created, token, nil
}

// Login authenticates an existing user and returns the user plus a JWT. The
// returned user never carries the password hash.
func (s *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
) (*models.User, string, error) {

	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if user == nil || !user.IsActive {
		return nil, "", ErrInvalidCredentials
	}

	// Compare the supplied password against the stored bcrypt hash.
	if !utils.CheckPassword(password, user.Password) {
		return nil, "", ErrInvalidCredentials
	}

	token, err := s.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, "", err
	}

	user.Password = ""

	return user, token, nil
}
