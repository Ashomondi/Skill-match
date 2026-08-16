package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"skill-match/backend/models"
	"skill-match/backend/repositories"
	"skill-match/backend/services"
	"skill-match/backend/utils"
)

type handlerFakeUserRepo struct {
	usersByEmail map[string]*models.User
	seq          int
}

func newHandlerFakeUserRepo() *handlerFakeUserRepo {
	return &handlerFakeUserRepo{usersByEmail: map[string]*models.User{}}
}

func (f *handlerFakeUserRepo) Create(_ context.Context, u *models.User) (*models.User, error) {
	f.seq++
	id := "user-" + string(rune('a'+f.seq-1))

	stored := *u
	stored.ID = id
	stored.IsActive = true // DB column default
	f.usersByEmail[stored.Email] = &stored

	returned := stored
	return &returned, nil
}

func (f *handlerFakeUserRepo) GetByEmail(_ context.Context, email string) (*models.User, error) {
	if u, ok := f.usersByEmail[email]; ok {
		return u, nil
	}
	return nil, repositories.ErrUserNotFound
}

func newTestAuthHandler() *AuthHandler {
	return NewAuthHandler(services.NewAuthService(newHandlerFakeUserRepo(), utils.NewJWTManager("handler-test-secret", time.Hour)))
}

func doJSON(h func(http.ResponseWriter, *http.Request), body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

func TestRegisterHandlerSuccess(t *testing.T) {
	rec := doJSON(newTestAuthHandler().Register, `{"email":"jane@example.com","password":"password123","fullName":"Jane Doe"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp AuthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Token == "" || resp.User == nil || resp.User.Email != "jane@example.com" {
		t.Fatalf("expected token and user in response: %+v", resp)
	}
	if resp.User.Password != "" {
		t.Fatal("password must not leak in the response")
	}
}

func TestRegisterHandlerBadBody(t *testing.T) {
	rec := doJSON(newTestAuthHandler().Register, `{not-json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRegisterHandlerDuplicate(t *testing.T) {
	h := newTestAuthHandler()
	if rec := doJSON(h.Register, `{"email":"jane@example.com","password":"password123"}`); rec.Code != http.StatusCreated {
		t.Fatalf("setup register: %d", rec.Code)
	}
	rec := doJSON(h.Register, `{"email":"jane@example.com","password":"password123"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for duplicate, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLoginHandlerSuccess(t *testing.T) {
	h := newTestAuthHandler()
	if rec := doJSON(h.Register, `{"email":"jane@example.com","password":"password123","fullName":"Jane"}`); rec.Code != http.StatusCreated {
		t.Fatalf("setup register: %d", rec.Code)
	}

	rec := doJSON(h.Login, `{"email":"jane@example.com","password":"password123"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp AuthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Token == "" || resp.User == nil {
		t.Fatalf("expected token and user, got %+v", resp)
	}
}

func TestLoginHandlerWrongPassword(t *testing.T) {
	h := newTestAuthHandler()
	if rec := doJSON(h.Register, `{"email":"jane@example.com","password":"password123"}`); rec.Code != http.StatusCreated {
		t.Fatalf("setup register: %d", rec.Code)
	}

	rec := doJSON(h.Login, `{"email":"jane@example.com","password":"nope"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}
