package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"skill-match/backend/clients"
	"skill-match/backend/repositories"
	"skill-match/backend/utils"
)

type fakeUserRepo struct {
	byID      map[string]*repositories.User
	byEmail   map[string]*repositories.User
	byGoogle  map[string]*repositories.User
	createErr error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		byID:     map[string]*repositories.User{},
		byEmail:  map[string]*repositories.User{},
		byGoogle: map[string]*repositories.User{},
	}
}

func (f *fakeUserRepo) Create(_ context.Context, user *repositories.User) (*repositories.User, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	if _, exists := f.byEmail[user.Email]; exists {
		return nil, repositories.ErrUserEmailTaken
	}
	copied := *user
	if copied.ID == "" {
		copied.ID = "generated-id"
	}
	f.byID[copied.ID] = &copied
	f.byEmail[copied.Email] = &copied
	if copied.GoogleID != "" {
		f.byGoogle[copied.GoogleID] = &copied
	}
	return &copied, nil
}

func (f *fakeUserRepo) GetByEmail(_ context.Context, email string) (*repositories.User, error) {
	if user, ok := f.byEmail[email]; ok {
		return user, nil
	}
	return nil, repositories.ErrUserNotFound
}

func (f *fakeUserRepo) GetByGoogleID(_ context.Context, googleID string) (*repositories.User, error) {
	if user, ok := f.byGoogle[googleID]; ok {
		return user, nil
	}
	return nil, repositories.ErrUserNotFound
}

func (f *fakeUserRepo) LinkGoogleID(_ context.Context, id, googleID string) (*repositories.User, error) {
	user, ok := f.byID[id]
	if !ok {
		return nil, repositories.ErrUserNotFound
	}
	user.GoogleID = googleID
	f.byGoogle[googleID] = user
	return user, nil
}

type fakeGoogleVerifier struct {
	identity *clients.GoogleIdentity
	err      error
}

func (f *fakeGoogleVerifier) Verify(_ context.Context, _ string) (*clients.GoogleIdentity, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.identity, nil
}

func newTestAuthService(repo UserRepository) (*AuthService, *utils.JWTManager) {
	jwtManager := utils.NewJWTManager("test-secret", time.Hour)
	return NewAuthService(repo, jwtManager), jwtManager
}

func TestLoginWithGoogleCreatesAccount(t *testing.T) {
	repo := newFakeUserRepo()
	service, jwtManager := newTestAuthService(repo)
	service.SetGoogleVerifier(&fakeGoogleVerifier{identity: &clients.GoogleIdentity{
		Subject:       "google-sub",
		Email:         "jane@example.com",
		EmailVerified: true,
		Name:          "Jane Doe",
	}})

	user, token, err := service.LoginWithGoogle(context.Background(), "token")
	if err != nil {
		t.Fatalf("LoginWithGoogle returned error: %v", err)
	}

	if user.Email != "jane@example.com" || user.FullName != "Jane Doe" {
		t.Errorf("unexpected user: %+v", user)
	}
	if user.GoogleID != "google-sub" {
		t.Errorf("GoogleID = %q, want google-sub", user.GoogleID)
	}
	if user.PasswordHash != "" {
		t.Error("password hash must not be returned")
	}
	if _, err := jwtManager.ValidateToken(token); err != nil {
		t.Errorf("returned token is invalid: %v", err)
	}
	if len(repo.byEmail) != 1 {
		t.Errorf("expected 1 stored user, got %d", len(repo.byEmail))
	}
}

func TestLoginWithGoogleReturnsExistingLinkedAccount(t *testing.T) {
	repo := newFakeUserRepo()
	existing := &repositories.User{ID: "u1", Email: "jane@example.com", GoogleID: "google-sub", IsActive: true}
	repo.byID[existing.ID] = existing
	repo.byEmail[existing.Email] = existing
	repo.byGoogle[existing.GoogleID] = existing

	service, _ := newTestAuthService(repo)
	service.SetGoogleVerifier(&fakeGoogleVerifier{identity: &clients.GoogleIdentity{
		Subject:       "google-sub",
		Email:         "jane@example.com",
		EmailVerified: true,
	}})

	user, _, err := service.LoginWithGoogle(context.Background(), "token")
	if err != nil {
		t.Fatalf("LoginWithGoogle returned error: %v", err)
	}
	if user.ID != "u1" {
		t.Errorf("user.ID = %q, want u1", user.ID)
	}
}

func TestLoginWithGoogleLinksExistingEmailAccount(t *testing.T) {
	repo := newFakeUserRepo()
	existing := &repositories.User{ID: "u1", Email: "jane@example.com", PasswordHash: "hash", IsActive: true}
	repo.byID[existing.ID] = existing
	repo.byEmail[existing.Email] = existing

	service, _ := newTestAuthService(repo)
	service.SetGoogleVerifier(&fakeGoogleVerifier{identity: &clients.GoogleIdentity{
		Subject:       "google-sub",
		Email:         "jane@example.com",
		EmailVerified: true,
	}})

	user, _, err := service.LoginWithGoogle(context.Background(), "token")
	if err != nil {
		t.Fatalf("LoginWithGoogle returned error: %v", err)
	}
	if user.ID != "u1" {
		t.Errorf("user.ID = %q, want linked u1", user.ID)
	}
	if existing.GoogleID != "google-sub" {
		t.Errorf("existing account not linked, GoogleID = %q", existing.GoogleID)
	}
}

func TestLoginWithGoogleUnavailableWithoutVerifier(t *testing.T) {
	service, _ := newTestAuthService(newFakeUserRepo())

	_, _, err := service.LoginWithGoogle(context.Background(), "token")
	if !errors.Is(err, ErrGoogleAuthUnavailable) {
		t.Fatalf("error = %v, want ErrGoogleAuthUnavailable", err)
	}
}

func TestLoginWithGoogleRejectsInvalidToken(t *testing.T) {
	service, _ := newTestAuthService(newFakeUserRepo())
	service.SetGoogleVerifier(&fakeGoogleVerifier{err: clients.ErrGoogleTokenInvalid})

	_, _, err := service.LoginWithGoogle(context.Background(), "token")
	if !errors.Is(err, ErrInvalidGoogleToken) {
		t.Fatalf("error = %v, want ErrInvalidGoogleToken", err)
	}
}

func TestLoginWithGoogleMapsVerificationOutage(t *testing.T) {
	service, _ := newTestAuthService(newFakeUserRepo())
	service.SetGoogleVerifier(&fakeGoogleVerifier{err: clients.ErrGoogleVerificationUnavailable})

	_, _, err := service.LoginWithGoogle(context.Background(), "token")
	if !errors.Is(err, ErrGoogleAuthUnavailable) {
		t.Fatalf("error = %v, want ErrGoogleAuthUnavailable", err)
	}
}
