package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"

	"skill-match/backend/models"
	"skill-match/backend/repositories"
	"skill-match/backend/utils"
)

// Profile payload limits. These are deliberately permissive enough for real CV
// content while keeping a single profile bounded.
const (
	MaxProfileBioLength         = 2000
	MaxProfileSkills            = 100
	MaxProfileSkillLength       = 50
	MaxProfileExperienceEntries = 50
	MaxProfileExperienceLength  = 500
	MaxProfileResumeURLLength   = 2048
)

// ProfileStore is the persistence subset the profile service needs.
// repositories.ProfileRepository satisfies this interface.
type ProfileStore interface {
	GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error)
	Update(ctx context.Context, userID string, patch models.UserProfilePatch) (*models.UserProfile, error)
	Upsert(ctx context.Context, profile *models.UserProfile) (*models.UserProfile, error)
}

// ProfileService owns profile validation and the "no profile yet" behaviour.
type ProfileService struct {
	store ProfileStore
}

// NewProfileService constructs a ProfileService.
func NewProfileService(store ProfileStore) *ProfileService {
	return &ProfileService{store: store}
}

// Get returns the user's profile. When the user has no stored profile yet it
// returns a default empty profile instead of an error, so the frontend can
// render an empty state.
func (s *ProfileService) Get(ctx context.Context, userID string) (*models.UserProfile, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, newProfileValidationError("user_id is required")
	}

	profile, err := s.store.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrProfileNotFound) {
			return defaultProfile(userID), nil
		}
		return nil, err
	}
	return normalizeProfile(profile), nil
}

// Update validates and applies a partial profile update. When the user has no
// profile row yet, the validated patch is persisted with an upsert so PUT is
// safe to call on a brand-new account.
func (s *ProfileService) Update(ctx context.Context, userID string, patch models.UserProfilePatch) (*models.UserProfile, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, newProfileValidationError("user_id is required")
	}

	normalized, err := normalizeProfilePatch(patch)
	if err != nil {
		return nil, err
	}

	profile, err := s.store.Update(ctx, userID, normalized)
	if err == nil {
		return normalizeProfile(profile), nil
	}
	if !errors.Is(err, repositories.ErrProfileNotFound) {
		return nil, err
	}

	created, err := s.store.Upsert(ctx, profileFromPatch(userID, normalized))
	if err != nil {
		return nil, err
	}
	return normalizeProfile(created), nil
}

func defaultProfile(userID string) *models.UserProfile {
	return &models.UserProfile{
		UserID:     userID,
		Skills:     []string{},
		Experience: []string{},
	}
}

func profileFromPatch(userID string, patch models.UserProfilePatch) *models.UserProfile {
	profile := defaultProfile(userID)
	if patch.Bio != nil {
		profile.Bio = *patch.Bio
	}
	if patch.Skills != nil {
		profile.Skills = *patch.Skills
	}
	if patch.Experience != nil {
		profile.Experience = *patch.Experience
	}
	if patch.ResumeURL != nil {
		profile.ResumeURL = *patch.ResumeURL
	}
	return profile
}

// normalizeProfile guarantees JSON-friendly, non-nil string lists.
func normalizeProfile(profile *models.UserProfile) *models.UserProfile {
	if profile.Skills == nil {
		profile.Skills = []string{}
	}
	if profile.Experience == nil {
		profile.Experience = []string{}
	}
	return profile
}

func normalizeProfilePatch(patch models.UserProfilePatch) (models.UserProfilePatch, error) {
	if patch.Empty() {
		return patch, newProfileValidationError("at least one profile field is required")
	}

	if patch.Bio != nil {
		bio := strings.TrimSpace(*patch.Bio)
		if utf8.RuneCountInString(bio) > MaxProfileBioLength {
			return patch, newProfileValidationError(
				fmt.Sprintf("bio must be at most %d characters", MaxProfileBioLength))
		}
		if hasDisallowedControl(bio, true) {
			return patch, newProfileValidationError("bio contains unsupported control characters")
		}
		patch.Bio = &bio
	}

	if patch.Skills != nil {
		skills, err := normalizeProfileTags(*patch.Skills, "skill", MaxProfileSkills, MaxProfileSkillLength, false)
		if err != nil {
			return patch, err
		}
		patch.Skills = &skills
	}

	if patch.Experience != nil {
		experience, err := normalizeProfileTags(*patch.Experience, "experience entry", MaxProfileExperienceEntries, MaxProfileExperienceLength, true)
		if err != nil {
			return patch, err
		}
		patch.Experience = &experience
	}

	if patch.ResumeURL != nil {
		raw := strings.TrimSpace(*patch.ResumeURL)
		if len(raw) > MaxProfileResumeURLLength {
			return patch, newProfileValidationError(
				fmt.Sprintf("resume_url must be at most %d characters", MaxProfileResumeURLLength))
		}
		if raw != "" {
			parsed, err := url.ParseRequestURI(raw)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
				return patch, newProfileValidationError("resume_url must be a valid http or https URL")
			}
		}
		patch.ResumeURL = &raw
	}

	return patch, nil
}

// normalizeProfileTags trims, validates, and de-duplicates (case-insensitively)
// an ordered list of tags. Empty input is allowed and yields an empty slice so
// callers can clear a list explicitly.
func normalizeProfileTags(values []string, label string, maxCount, maxLength int, singleLine bool) ([]string, error) {
	if len(values) > maxCount {
		return nil, newProfileValidationError(
			fmt.Sprintf("at most %d %ss are allowed", maxCount, label))
	}

	out := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, raw := range values {
		tag := strings.TrimSpace(raw)
		if tag == "" {
			return nil, newProfileValidationError(fmt.Sprintf("%s tags cannot be empty", label))
		}
		if utf8.RuneCountInString(tag) > maxLength {
			return nil, newProfileValidationError(
				fmt.Sprintf("%s tags must be at most %d characters", label, maxLength))
		}
		if singleLine && strings.ContainsAny(tag, "\r\n") {
			return nil, newProfileValidationError(
				fmt.Sprintf("%s entries must be single-line", label))
		}
		if hasDisallowedControl(tag, false) || strings.ContainsAny(tag, "<>") {
			return nil, newProfileValidationError(
				fmt.Sprintf("%s tags contain unsupported characters", label))
		}

		key := strings.ToLower(tag)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, tag)
	}
	return out, nil
}

// hasDisallowedControl reports whether s contains control characters. When
// withNewlines is true, tab/newline/carriage-return are permitted (for bio);
// otherwise every control character is rejected.
func hasDisallowedControl(s string, withNewlines bool) bool {
	for _, r := range s {
		if withNewlines && (r == '\n' || r == '\r' || r == '\t') {
			continue
		}
		if r < 32 || r == 127 {
			return true
		}
	}
	return false
}

func newProfileValidationError(message string) error {
	return utils.NewValidationError(message, map[string]string{"operation": "profile.validate"})
}
