package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"skill-match/backend/clients"
	"skill-match/backend/repositories"
	"skill-match/backend/utils"
)

var (
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrInvalidEmail          = errors.New("invalid email")
	ErrInvalidPassword       = errors.New("password must be at least 8 characters")
	ErrInvalidGoogleToken    = errors.New("invalid google id token")
	ErrGoogleAuthUnavailable = errors.New("google authentication unavailable")
)

// UserRepository defines the database operations required by authentication.
type UserRepository interface {
	Create(ctx context.Context, user *repositories.User) (*repositories.User, error)
	GetByEmail(ctx context.Context, email string) (*repositories.User, error)
	GetByGoogleID(ctx context.Context, googleID string) (*repositories.User, error)
	LinkGoogleID(ctx context.Context, id, googleID string) (*repositories.User, error)
}

// GoogleIdentityVerifier verifies a Google ID token and returns its identity.
type GoogleIdentityVerifier interface {
	Verify(ctx context.Context, idToken string) (*clients.GoogleIdentity, error)
}

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo       UserRepository
	jwtManager     *utils.JWTManager
	googleVerifier GoogleIdentityVerifier
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

// SetGoogleVerifier enables sign-in with Google by supplying a token verifier.
// When no verifier is set, LoginWithGoogle returns ErrGoogleAuthUnavailable.
func (s *AuthService) SetGoogleVerifier(verifier GoogleIdentityVerifier) {
	s.googleVerifier = verifier
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

	return s.issueToken(created)
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

	return s.issueToken(user)
}

// LoginWithGoogle verifies a Google ID token, then finds, links, or creates the
// matching local account and returns it with a fresh application JWT.
//
// Accounts are matched by Google subject first. If no account is linked yet,
// an existing password account with the same verified email is linked instead
// so users can sign in either way.
func (s *AuthService) LoginWithGoogle(ctx context.Context, idToken string) (*repositories.User, string, error) {
	if s.googleVerifier == nil {
		return nil, "", ErrGoogleAuthUnavailable
	}

	identity, err := s.googleVerifier.Verify(ctx, idToken)
	if err != nil {
		if errors.Is(err, clients.ErrGoogleVerificationUnavailable) {
			return nil, "", ErrGoogleAuthUnavailable
		}
		return nil, "", ErrInvalidGoogleToken
	}

	email := strings.ToLower(strings.TrimSpace(identity.Email))
	if strings.TrimSpace(identity.Subject) == "" || email == "" || !identity.EmailVerified {
		return nil, "", ErrInvalidGoogleToken
	}

	// 1. An account already linked to this Google subject.
	user, err := s.userRepo.GetByGoogleID(ctx, identity.Subject)
	switch {
	case err == nil && user != nil:
		return s.issueToken(user)
	case err != nil && !errors.Is(err, repositories.ErrUserNotFound):
		return nil, "", err
	}

	// 2. A password account with the same (Google-verified) email: link it.
	existing, err := s.userRepo.GetByEmail(ctx, email)
	switch {
	case err == nil && existing != nil:
		linked, linkErr := s.userRepo.LinkGoogleID(ctx, existing.ID, identity.Subject)
		if linkErr != nil {
			return nil, "", linkErr
		}
		return s.issueToken(linked)
	case err != nil && !errors.Is(err, repositories.ErrUserNotFound):
		return nil, "", err
	}

	// 3. A brand new Google-only account.
	now := time.Now()

	created, err := s.userRepo.Create(ctx, &repositories.User{
		ID:        uuid.NewString(),
		Email:     email,
		GoogleID:  identity.Subject,
		FullName:  strings.TrimSpace(identity.Name),
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return nil, "", err
	}

	return s.issueToken(created)
}

// issueToken mints a JWT for user and strips the password hash before the
// user object is returned to callers.
func (s *AuthService) issueToken(user *repositories.User) (*repositories.User, string, error) {
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, "", err
	}

	// Never return the password hash.
	user.PasswordHash = ""

	return user, token, nil
}
