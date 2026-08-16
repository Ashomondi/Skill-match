package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"skill-match/backend/models"
	"skill-match/backend/repositories"
	"skill-match/backend/utils"
)

type fakeUserRepo struct {
	usersByEmail map[string]*models.User
	createErr    error
	getErr       error
	seq          int
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{usersByEmail: map[string]*models.User{}}
}

func (f *fakeUserRepo) Create(_ context.Context, u *models.User) (*models.User, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.seq++
	id := "user-" + string(rune('a'+f.seq-1))

	stored := *u
	stored.ID = id
	stored.IsActive = true // DB column default
	f.usersByEmail[stored.Email] = &stored

	returned := stored
	return &returned, nil
}

func (f *fakeUserRepo) GetByEmail(_ context.Context, email string) (*models.User, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if u, ok := f.usersByEmail[email]; ok {
		return u, nil
	}
	return nil, repositories.ErrUserNotFound
}

func testAuthService(repo *fakeUserRepo) *AuthService {
	return NewAuthService(repo, utils.NewJWTManager("unit-test-secret", time.Hour))
}

func TestRegisterCreatesUserAndReturnsToken(t *testing.T) {
	repo := newFakeUserRepo()
	svc := testAuthService(repo)

	user, token, err := svc.Register(context.Background(), "Jane@Example.com", "password123", "Jane Doe")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if user.Email != "jane@example.com" {
		t.Fatalf("expected lowercased email, got %q", user.Email)
	}
	if user.FullName != "Jane Doe" {
		t.Fatalf("expected full name, got %q", user.FullName)
	}
	if user.Password != "" {
		t.Fatal("password hash must never be returned")
	}
	if token == "" {
		t.Fatal("expected a token")
	}

	claims, err := svc.jwtManager.ValidateToken(token)
	if err != nil {
		t.Fatalf("token should validate: %v", err)
	}
	if claims.UserID != user.ID {
		t.Fatalf("expected token user id to match, got %q", claims.UserID)
	}
}

func TestRegisterRejectsInvalidInput(t *testing.T) {
	svc := testAuthService(newFakeUserRepo())
	ctx := context.Background()

	if _, _, err := svc.Register(ctx, "not-an-email", "password123", "N"); !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
	if _, _, err := svc.Register(ctx, "a@b.com", "short", "N"); !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestRegisterRejectsDuplicateEmail(t *testing.T) {
	repo := newFakeUserRepo()
	svc := testAuthService(repo)
	ctx := context.Background()

	if _, _, err := svc.Register(ctx, "jane@example.com", "password123", "Jane"); err != nil {
		t.Fatalf("first register: %v", err)
	}
	if _, _, err := svc.Register(ctx, "JANE@example.com", "password456", "Jane Two"); !errors.Is(err, ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestLoginReturnsUserAndToken(t *testing.T) {
	repo := newFakeUserRepo()
	svc := testAuthService(repo)
	ctx := context.Background()

	if _, _, err := svc.Register(ctx, "jane@example.com", "password123", "Jane"); err != nil {
		t.Fatalf("register: %v", err)
	}

	user, token, err := svc.Login(ctx, "jane@example.com", "password123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if user.Email != "jane@example.com" || user.Password != "" {
		t.Fatalf("unexpected login user: %+v", user)
	}
	if token == "" {
		t.Fatal("expected a token")
	}
}

func TestLoginRejectsBadCredentials(t *testing.T) {
	repo := newFakeUserRepo()
	svc := testAuthService(repo)
	ctx := context.Background()

	if _, _, err := svc.Register(ctx, "jane@example.com", "password123", "Jane"); err != nil {
		t.Fatalf("register: %v", err)
	}

	if _, _, err := svc.Login(ctx, "jane@example.com", "wrongpassword"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for wrong password, got %v", err)
	}
	if _, _, err := svc.Login(ctx, "nobody@example.com", "password123"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for unknown email, got %v", err)
	}
}

func TestLoginRejectsInactiveUser(t *testing.T) {
	repo := newFakeUserRepo()
	svc := testAuthService(repo)
	ctx := context.Background()

	if _, _, err := svc.Register(ctx, "jane@example.com", "password123", "Jane"); err != nil {
		t.Fatalf("register: %v", err)
	}
	repo.usersByEmail["jane@example.com"].IsActive = false // DB soft-disable

	if _, _, err := svc.Login(ctx, "jane@example.com", "password123"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for inactive user, got %v", err)
	}
}
