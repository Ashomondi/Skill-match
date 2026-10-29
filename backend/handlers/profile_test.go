package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"skill-match/backend/middleware"
	"skill-match/backend/models"
	"skill-match/backend/repositories"
	"skill-match/backend/services"
	"skill-match/backend/utils"
)

type handlerFakeProfileStore struct {
	profile     *models.UserProfile
	updateCalls int
	lastPatch   models.UserProfilePatch
	upserted    *models.UserProfile
}

func (f *handlerFakeProfileStore) GetByUserID(_ context.Context, _ string) (*models.UserProfile, error) {
	if f.profile == nil {
		return nil, repositories.ErrProfileNotFound
	}
	return f.profile, nil
}

func (f *handlerFakeProfileStore) Update(_ context.Context, _ string, patch models.UserProfilePatch) (*models.UserProfile, error) {
	f.updateCalls++
	f.lastPatch = patch
	if f.profile == nil {
		return nil, repositories.ErrProfileNotFound
	}
	out := *f.profile
	if patch.Bio != nil {
		out.Bio = *patch.Bio
	}
	if patch.Skills != nil {
		out.Skills = append([]string{}, (*patch.Skills)...)
	}
	if patch.Experience != nil {
		out.Experience = append([]string{}, (*patch.Experience)...)
	}
	if patch.ResumeURL != nil {
		out.ResumeURL = *patch.ResumeURL
	}
	return &out, nil
}

func (f *handlerFakeProfileStore) Upsert(_ context.Context, profile *models.UserProfile) (*models.UserProfile, error) {
	f.upserted = profile
	return profile, nil
}

const profileTestSecret = "profile-handler-test-secret"

func newTestProfileHandler() (*ProfileHandler, *handlerFakeProfileStore) {
	store := &handlerFakeProfileStore{}
	return NewProfileHandler(services.NewProfileService(store)), store
}

func profileBearerToken(t *testing.T, userID string) string {
	t.Helper()
	token, err := utils.NewJWTManager(profileTestSecret, time.Hour).GenerateToken(userID, "user@example.com")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return token
}

func profileHandler(h *ProfileHandler, method string) http.Handler {
	auth := middleware.Auth(utils.NewJWTManager(profileTestSecret, time.Hour))
	switch method {
	case http.MethodGet:
		return auth(http.HandlerFunc(h.Get))
	default:
		return auth(http.HandlerFunc(h.Update))
	}
}

func doProfileRequest(t *testing.T, handler http.Handler, userID, body string, authorized bool) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(http.MethodGet, "/api/profile", nil)
	} else {
		req = httptest.NewRequest(http.MethodPut, "/api/profile", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}
	if authorized {
		req.Header.Set("Authorization", "Bearer "+profileBearerToken(t, userID))
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

type profileEnvelope struct {
	Data struct {
		UserID     string     `json:"user_id"`
		Bio        string     `json:"bio"`
		Skills     []string   `json:"skills"`
		Experience []string   `json:"experience"`
		ResumeURL  string     `json:"resume_url"`
		UpdatedAt  *time.Time `json:"updated_at"`
	} `json:"data"`
}

func decodeProfileEnvelope(t *testing.T, rec *httptest.ResponseRecorder) profileEnvelope {
	t.Helper()
	var envelope profileEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v (body=%s)", err, rec.Body.String())
	}
	return envelope
}

func TestProfileGetRequiresAuth(t *testing.T) {
	h, _ := newTestProfileHandler()
	rec := doProfileRequest(t, profileHandler(h, http.MethodGet), "user-1", "", false)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestProfileGetReturnsDefaultWhenMissing(t *testing.T) {
	h, _ := newTestProfileHandler()
	rec := doProfileRequest(t, profileHandler(h, http.MethodGet), "user-1", "", true)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	envelope := decodeProfileEnvelope(t, rec)
	if envelope.Data.UserID != "user-1" {
		t.Fatalf("expected user-1, got %q", envelope.Data.UserID)
	}
	if envelope.Data.Skills == nil || envelope.Data.Experience == nil {
		t.Fatalf("expected [] lists, got %s", rec.Body.String())
	}
	if len(envelope.Data.Skills) != 0 || len(envelope.Data.Experience) != 0 {
		t.Fatalf("expected empty lists, got %s", rec.Body.String())
	}
	if envelope.Data.UpdatedAt != nil {
		t.Fatalf("expected updated_at to be omitted for a default profile, got %s", rec.Body.String())
	}
}

func TestProfileGetReturnsStoredProfile(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	store := &handlerFakeProfileStore{profile: &models.UserProfile{
		UserID:     "user-1",
		Bio:        "Backend engineer",
		Skills:     []string{"Go", "PostgreSQL"},
		Experience: []string{"Acme — Backend Engineer"},
		ResumeURL:  "https://example.com/cv.pdf",
		UpdatedAt:  now,
	}}
	h := NewProfileHandler(services.NewProfileService(store))

	rec := doProfileRequest(t, profileHandler(h, http.MethodGet), "user-1", "", true)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	envelope := decodeProfileEnvelope(t, rec)
	if envelope.Data.Bio != "Backend engineer" {
		t.Fatalf("unexpected bio: %q", envelope.Data.Bio)
	}
	if len(envelope.Data.Skills) != 2 || envelope.Data.Skills[0] != "Go" {
		t.Fatalf("unexpected skills: %+v", envelope.Data.Skills)
	}
	if envelope.Data.UpdatedAt == nil {
		t.Fatal("expected updated_at to be present")
	}
}

func TestProfileUpdateRequiresAuth(t *testing.T) {
	h, _ := newTestProfileHandler()
	rec := doProfileRequest(t, profileHandler(h, http.MethodPut), "user-1", `{"bio":"x"}`, false)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestProfileUpdateReturnsUpdatedProfile(t *testing.T) {
	store := &handlerFakeProfileStore{profile: &models.UserProfile{UserID: "user-1", Bio: "old"}}
	h := NewProfileHandler(services.NewProfileService(store))

	rec := doProfileRequest(t, profileHandler(h, http.MethodPut), "user-1",
		`{"bio":"new bio","skills":["Go","R&D"],"experience":["Acme"],"resume_url":"https://example.com/cv.pdf"}`, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	envelope := decodeProfileEnvelope(t, rec)
	if envelope.Data.Bio != "new bio" {
		t.Fatalf("unexpected bio: %q", envelope.Data.Bio)
	}
	if len(envelope.Data.Skills) != 2 || envelope.Data.Skills[1] != "R&D" {
		t.Fatalf("unexpected skills: %+v", envelope.Data.Skills)
	}
	if store.lastPatch.Bio == nil || *store.lastPatch.Bio != "new bio" {
		t.Fatalf("expected validated patch to reach the store, got %+v", store.lastPatch)
	}
}

func TestProfileUpdateCreatesWhenMissing(t *testing.T) {
	h, store := newTestProfileHandler()
	rec := doProfileRequest(t, profileHandler(h, http.MethodPut), "user-1", `{"bio":"new profile"}`, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.upserted == nil {
		t.Fatal("expected an upsert when the profile is missing")
	}
	if store.upserted.UserID != "user-1" || store.upserted.Bio != "new profile" {
		t.Fatalf("unexpected upserted profile: %+v", store.upserted)
	}
}

func TestProfileUpdateRejectsInvalidPayloads(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "empty object", body: `{}`},
		{name: "malformed json", body: `not-json`},
		{name: "unknown field", body: `{"unknown":"x"}`},
		{name: "trailing json", body: `{"bio":"a"}{"bio":"b"}`},
		{name: "bio too long", body: `{"bio":"` + strings.Repeat("a", 2001) + `"}`},
		{name: "invalid skill", body: `{"skills":["<script>"]}`},
		{name: "invalid resume url", body: `{"resume_url":"not-a-url"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, _ := newTestProfileHandler()
			rec := doProfileRequest(t, profileHandler(h, http.MethodPut), "user-1", tt.body, true)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}
