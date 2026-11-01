package services

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"skill-match/backend/models"
	"skill-match/backend/repositories"
	"skill-match/backend/utils"
)

type fakeProfileStore struct {
	profile     *models.UserProfile
	getErr      error
	updateErr   error
	upsertErr   error
	updateCalls int
	lastPatch   models.UserProfilePatch
	upserted    *models.UserProfile
}

func (f *fakeProfileStore) GetByUserID(_ context.Context, _ string) (*models.UserProfile, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if f.profile == nil {
		return nil, repositories.ErrProfileNotFound
	}
	return f.profile, nil
}

func (f *fakeProfileStore) Update(_ context.Context, _ string, patch models.UserProfilePatch) (*models.UserProfile, error) {
	f.updateCalls++
	f.lastPatch = patch
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	if f.profile == nil {
		return nil, repositories.ErrProfileNotFound
	}
	return applyProfilePatch(f.profile, patch), nil
}

func (f *fakeProfileStore) Upsert(_ context.Context, profile *models.UserProfile) (*models.UserProfile, error) {
	f.upserted = profile
	if f.upsertErr != nil {
		return nil, f.upsertErr
	}
	return profile, nil
}

func applyProfilePatch(profile *models.UserProfile, patch models.UserProfilePatch) *models.UserProfile {
	out := *profile
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
	return &out
}

func profileStringPtr(v string) *string { return &v }

func profileSlicePtr(v []string) *[]string { return &v }

func requireProfileValidationError(t *testing.T, err error) {
	t.Helper()
	appErr, ok := utils.AsAppError(err)
	if !ok {
		t.Fatalf("expected an AppError, got %T: %v", err, err)
	}
	if appErr.StatusCode != http.StatusBadRequest || appErr.Category != utils.CategoryValidation {
		t.Fatalf("expected 400 validation error, got status=%d category=%q", appErr.StatusCode, appErr.Category)
	}
}

func TestProfileServiceGetReturnsDefaultWhenMissing(t *testing.T) {
	service := NewProfileService(&fakeProfileStore{})

	profile, err := service.Get(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.UserID != "user-1" || profile.Bio != "" || profile.ResumeURL != "" {
		t.Fatalf("unexpected default profile: %+v", profile)
	}
	if profile.Skills == nil || profile.Experience == nil {
		t.Fatalf("expected non-nil empty lists, got skills=%v experience=%v", profile.Skills, profile.Experience)
	}
	if len(profile.Skills) != 0 || len(profile.Experience) != 0 {
		t.Fatalf("expected empty lists, got %+v", profile)
	}
}

func TestProfileServiceGetNormalizesNilLists(t *testing.T) {
	store := &fakeProfileStore{profile: &models.UserProfile{UserID: "user-1", Bio: "hi"}}
	service := NewProfileService(store)

	profile, err := service.Get(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.Skills == nil || profile.Experience == nil {
		t.Fatalf("expected non-nil lists, got %+v", profile)
	}
}

func TestProfileServiceGetPropagatesStoreError(t *testing.T) {
	storeErr := errors.New("database unavailable")
	service := NewProfileService(&fakeProfileStore{getErr: storeErr})

	if _, err := service.Get(context.Background(), "user-1"); !errors.Is(err, storeErr) {
		t.Fatalf("expected store error, got %v", err)
	}
}

func TestProfileServiceUpdateValidation(t *testing.T) {
	tooManySkills := make([]string, MaxProfileSkills+1)
	for i := range tooManySkills {
		tooManySkills[i] = "skill"
	}
	tooManyExperience := make([]string, MaxProfileExperienceEntries+1)
	for i := range tooManyExperience {
		tooManyExperience[i] = "role"
	}

	tests := []struct {
		name  string
		patch models.UserProfilePatch
	}{
		{name: "empty patch", patch: models.UserProfilePatch{}},
		{name: "bio too long", patch: models.UserProfilePatch{Bio: profileStringPtr(strings.Repeat("a", MaxProfileBioLength+1))}},
		{name: "bio control character", patch: models.UserProfilePatch{Bio: profileStringPtr("bad\x00bio")}},
		{name: "skill empty", patch: models.UserProfilePatch{Skills: profileSlicePtr([]string{"   "})}},
		{name: "skill too long", patch: models.UserProfilePatch{Skills: profileSlicePtr([]string{strings.Repeat("x", MaxProfileSkillLength+1)})}},
		{name: "skill markup", patch: models.UserProfilePatch{Skills: profileSlicePtr([]string{"<script>"})}},
		{name: "too many skills", patch: models.UserProfilePatch{Skills: profileSlicePtr(tooManySkills)}},
		{name: "experience newline", patch: models.UserProfilePatch{Experience: profileSlicePtr([]string{"Acme\nEngineer"})}},
		{name: "experience too long", patch: models.UserProfilePatch{Experience: profileSlicePtr([]string{strings.Repeat("x", MaxProfileExperienceLength+1)})}},
		{name: "too many experience", patch: models.UserProfilePatch{Experience: profileSlicePtr(tooManyExperience)}},
		{name: "resume url not a url", patch: models.UserProfilePatch{ResumeURL: profileStringPtr("not-a-url")}},
		{name: "resume url scheme", patch: models.UserProfilePatch{ResumeURL: profileStringPtr("ftp://example.com/cv.pdf")}},
		{name: "resume url too long", patch: models.UserProfilePatch{ResumeURL: profileStringPtr("https://example.com/" + strings.Repeat("a", MaxProfileResumeURLLength))}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewProfileService(&fakeProfileStore{})
			_, err := service.Update(context.Background(), "user-1", tt.patch)
			requireProfileValidationError(t, err)
		})
	}
}

func TestProfileServiceUpdateNormalizesAndDeduplicates(t *testing.T) {
	store := &fakeProfileStore{profile: &models.UserProfile{UserID: "user-1"}}
	service := NewProfileService(store)

	profile, err := service.Update(context.Background(), "user-1", models.UserProfilePatch{
		Bio:        profileStringPtr("  Backend engineer  "),
		Skills:     profileSlicePtr([]string{"  Go  ", "go", "C++", "R&D", "日本語"}),
		Experience: profileSlicePtr([]string{"  Acme — Backend Engineer  "}),
		ResumeURL:  profileStringPtr("  https://example.com/cv.pdf  "),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if profile.Bio != "Backend engineer" {
		t.Fatalf("expected trimmed bio, got %q", profile.Bio)
	}
	wantSkills := []string{"Go", "C++", "R&D", "日本語"}
	if len(profile.Skills) != len(wantSkills) {
		t.Fatalf("expected %d skills, got %+v", len(wantSkills), profile.Skills)
	}
	for i, want := range wantSkills {
		if profile.Skills[i] != want {
			t.Fatalf("skill %d: got %q want %q", i, profile.Skills[i], want)
		}
	}
	if profile.Experience[0] != "Acme — Backend Engineer" {
		t.Fatalf("expected trimmed experience, got %q", profile.Experience[0])
	}
	if profile.ResumeURL != "https://example.com/cv.pdf" {
		t.Fatalf("expected trimmed resume url, got %q", profile.ResumeURL)
	}

	if store.lastPatch.Bio == nil || *store.lastPatch.Bio != "Backend engineer" {
		t.Fatalf("expected normalized patch to reach the store, got %+v", store.lastPatch)
	}
}

func TestProfileServiceUpdateCreatesWhenMissing(t *testing.T) {
	store := &fakeProfileStore{}
	service := NewProfileService(store)

	profile, err := service.Update(context.Background(), "user-1", models.UserProfilePatch{
		Bio:    profileStringPtr("New profile"),
		Skills: profileSlicePtr([]string{"Go"}),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.updateCalls != 1 {
		t.Fatalf("expected one update probe before upsert, got %d", store.updateCalls)
	}
	if store.upserted == nil {
		t.Fatal("expected an upsert when the profile is missing")
	}
	if store.upserted.UserID != "user-1" || store.upserted.Bio != "New profile" {
		t.Fatalf("unexpected upserted profile: %+v", store.upserted)
	}
	if len(store.upserted.Skills) != 1 || store.upserted.Skills[0] != "Go" {
		t.Fatalf("unexpected upserted skills: %+v", store.upserted.Skills)
	}
	if store.upserted.Experience == nil || len(store.upserted.Experience) != 0 {
		t.Fatalf("expected empty experience, got %+v", store.upserted.Experience)
	}
	if profile.Bio != "New profile" {
		t.Fatalf("unexpected returned profile: %+v", profile)
	}
}

func TestProfileServiceUpdatePropagatesStoreError(t *testing.T) {
	storeErr := errors.New("database unavailable")
	service := NewProfileService(&fakeProfileStore{updateErr: storeErr})

	if _, err := service.Update(context.Background(), "user-1", models.UserProfilePatch{
		Bio: profileStringPtr("hi"),
	}); !errors.Is(err, storeErr) {
		t.Fatalf("expected store error, got %v", err)
	}
}
