//go:build integration

package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"skill-match/backend/models"
	"skill-match/backend/utils"
)

// SM-02 — user profile persistence. Runs against a real PostgreSQL.
//
//	TEST_DATABASE_URL=... go test -tags integration ./repositories -run TestUserProfile -v

func TestUserProfileUpsertAndGet(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	repo := NewProfileRepository(pool)
	ctx := context.Background()

	_, err := repo.GetByUserID(ctx, user.ID)
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("expected ErrProfileNotFound for a new user, got %v", err)
	}
	appErr, ok := utils.AsAppError(err)
	if !ok {
		t.Fatalf("expected a canonical AppError, got %T", err)
	}
	if appErr.Category != utils.CategoryNotFound || appErr.StatusCode != 404 {
		t.Fatalf("unexpected AppError: category=%q status=%d", appErr.Category, appErr.StatusCode)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected pgx.ErrNoRows to stay in the chain, got %v", err)
	}

	created, err := repo.Upsert(ctx, &models.UserProfile{
		UserID:     user.ID,
		Bio:        "Backend engineer",
		Skills:     []string{"Go", "PostgreSQL"},
		Experience: []string{"Acme — Backend Engineer"},
		ResumeURL:  "https://example.com/cv.pdf",
	})
	if err != nil {
		t.Fatalf("upsert profile: %v", err)
	}
	if created.UserID != user.ID || created.Bio != "Backend engineer" {
		t.Fatalf("unexpected profile: %+v", created)
	}
	if !reflect.DeepEqual(created.Skills, []string{"Go", "PostgreSQL"}) {
		t.Fatalf("skills did not round-trip: %+v", created.Skills)
	}
	if !reflect.DeepEqual(created.Experience, []string{"Acme — Backend Engineer"}) {
		t.Fatalf("experience did not round-trip: %+v", created.Experience)
	}
	if created.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at to be set")
	}
	firstUpdated := created.UpdatedAt

	time.Sleep(50 * time.Millisecond)

	updated, err := repo.Upsert(ctx, &models.UserProfile{
		UserID:     user.ID,
		Bio:        "Senior backend engineer",
		Skills:     []string{"Go", "PostgreSQL", "Redis"},
		Experience: []string{"Acme — Backend Engineer", "Beta — Senior Engineer"},
		ResumeURL:  "https://example.com/cv-v2.pdf",
	})
	if err != nil {
		t.Fatalf("second upsert: %v", err)
	}
	if !updated.UpdatedAt.After(firstUpdated) {
		t.Fatalf("expected upsert to refresh updated_at: first=%s second=%s", firstUpdated, updated.UpdatedAt)
	}

	got, err := repo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}
	if got.Bio != "Senior backend engineer" || got.ResumeURL != "https://example.com/cv-v2.pdf" {
		t.Fatalf("unexpected persisted profile: %+v", got)
	}
	if !reflect.DeepEqual(got.Skills, []string{"Go", "PostgreSQL", "Redis"}) {
		t.Fatalf("expected parsed skills list, got %+v", got.Skills)
	}
	if !reflect.DeepEqual(got.Experience, []string{"Acme — Backend Engineer", "Beta — Senior Engineer"}) {
		t.Fatalf("expected parsed experience list, got %+v", got.Experience)
	}

	// The legacy name must stay usable by existing callers.
	legacy, err := repo.GetProfileByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("get profile via legacy method: %v", err)
	}
	if legacy.UserID != user.ID {
		t.Fatalf("unexpected legacy profile: %+v", legacy)
	}
}

func TestUserProfileSkillsSerializeDeserialize(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	repo := NewProfileRepository(pool)
	ctx := context.Background()

	empty, err := repo.Upsert(ctx, &models.UserProfile{UserID: user.ID})
	if err != nil {
		t.Fatalf("upsert empty profile: %v", err)
	}
	if empty.Skills == nil || empty.Experience == nil {
		t.Fatalf("expected non-nil empty lists, got skills=%v experience=%v", empty.Skills, empty.Experience)
	}
	raw, err := json.Marshal(empty)
	if err != nil {
		t.Fatalf("marshal profile: %v", err)
	}
	if !strings.Contains(string(raw), `"skills":[]`) || !strings.Contains(string(raw), `"experience":[]`) {
		t.Fatalf("expected empty lists to serialize as [], got %s", raw)
	}

	wantSkills := []string{"Go", "C++", "R&D", "日本語"}
	wantExperience := []string{"Acme / Platform", "Beta & Co."}
	saved, err := repo.Upsert(ctx, &models.UserProfile{
		UserID:     user.ID,
		Skills:     wantSkills,
		Experience: wantExperience,
	})
	if err != nil {
		t.Fatalf("upsert profile: %v", err)
	}
	if !reflect.DeepEqual(saved.Skills, wantSkills) {
		t.Fatalf("skills changed during write: got %+v want %+v", saved.Skills, wantSkills)
	}

	got, err := repo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}
	if !reflect.DeepEqual(got.Skills, wantSkills) {
		t.Fatalf("skills changed during read: got %+v want %+v", got.Skills, wantSkills)
	}
	if !reflect.DeepEqual(got.Experience, wantExperience) {
		t.Fatalf("experience changed during read: got %+v want %+v", got.Experience, wantExperience)
	}

	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal profile: %v", err)
	}
	var decoded models.UserProfile
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal profile: %v", err)
	}
	if !reflect.DeepEqual(decoded.Skills, wantSkills) || !reflect.DeepEqual(decoded.Experience, wantExperience) {
		t.Fatalf("JSON round-trip changed lists: %+v", decoded)
	}
}

func TestUserProfilePartialUpdate(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	repo := NewProfileRepository(pool)
	ctx := context.Background()

	base, err := repo.Upsert(ctx, &models.UserProfile{
		UserID:     user.ID,
		Bio:        "Original bio",
		Skills:     []string{"Go"},
		Experience: []string{"Acme"},
		ResumeURL:  "https://example.com/original.pdf",
	})
	if err != nil {
		t.Fatalf("upsert base profile: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	newSkills := []string{"Go", "Kubernetes"}
	updated, err := repo.Update(ctx, user.ID, models.UserProfilePatch{Skills: &newSkills})
	if err != nil {
		t.Fatalf("partial update skills: %v", err)
	}
	if !reflect.DeepEqual(updated.Skills, newSkills) {
		t.Fatalf("expected skills update, got %+v", updated.Skills)
	}
	if updated.Bio != "Original bio" || updated.ResumeURL != "https://example.com/original.pdf" {
		t.Fatalf("partial update clobbered untouched fields: %+v", updated)
	}
	if !reflect.DeepEqual(updated.Experience, []string{"Acme"}) {
		t.Fatalf("partial update clobbered experience: %+v", updated.Experience)
	}
	if !updated.UpdatedAt.After(base.UpdatedAt) {
		t.Fatalf("expected partial update to refresh updated_at: base=%s updated=%s", base.UpdatedAt, updated.UpdatedAt)
	}

	newBio := "Updated bio"
	emptyExperience := []string{}
	updated, err = repo.Update(ctx, user.ID, models.UserProfilePatch{
		Bio:        &newBio,
		Experience: &emptyExperience,
	})
	if err != nil {
		t.Fatalf("partial update bio/experience: %v", err)
	}
	if updated.Bio != newBio {
		t.Fatalf("expected bio update, got %q", updated.Bio)
	}
	if updated.Experience == nil || len(updated.Experience) != 0 {
		t.Fatalf("expected experience to be cleared to an empty list, got %+v", updated.Experience)
	}
	if !reflect.DeepEqual(updated.Skills, newSkills) {
		t.Fatalf("partial update clobbered skills: %+v", updated.Skills)
	}

	// Updating a profile that does not exist is a canonical not-found error.
	other := createTestUser(t, pool)
	_, err = repo.Update(ctx, other.ID, models.UserProfilePatch{Bio: &newBio})
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("expected ErrProfileNotFound for missing profile, got %v", err)
	}
	appErr, ok := utils.AsAppError(err)
	if !ok || appErr.Category != utils.CategoryNotFound {
		t.Fatalf("expected not-found AppError, got %#v", err)
	}
}

func TestUserProfileCascadeOnUserDelete(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	repo := NewProfileRepository(pool)
	ctx := context.Background()

	if _, err := repo.Upsert(ctx, &models.UserProfile{
		UserID: user.ID,
		Bio:    "Cascade me",
		Skills: []string{"Go"},
	}); err != nil {
		t.Fatalf("upsert profile: %v", err)
	}

	if err := NewUserRepository(pool).Delete(ctx, user.ID); err != nil {
		t.Fatalf("delete user: %v", err)
	}
	if n := countRows(t, pool, `SELECT count(*) FROM user_profiles WHERE user_id = $1`, user.ID); n != 0 {
		t.Fatalf("expected profile to cascade on user delete, got %d rows", n)
	}
	if _, err := repo.GetByUserID(ctx, user.ID); !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("expected ErrProfileNotFound after user delete, got %v", err)
	}
}

func TestUserProfileMissingUserMapsToNotFound(t *testing.T) {
	pool := connectTestPool(t)
	repo := NewProfileRepository(pool)
	ctx := context.Background()

	// Construct a valid but unused UUID so the INSERT hits the user FK.
	missingUser := "00000000-0000-0000-0000-000000000001"
	_, err := repo.Upsert(ctx, &models.UserProfile{
		UserID: missingUser,
		Bio:    "orphan",
	})
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("expected ErrProfileNotFound for missing user, got %v", err)
	}
	appErr, ok := utils.AsAppError(err)
	if !ok || appErr.Category != utils.CategoryNotFound {
		t.Fatalf("expected not-found AppError, got %#v", err)
	}
}
