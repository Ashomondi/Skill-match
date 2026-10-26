package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"skill-match/backend/models"
	"skill-match/backend/utils"
)

// Sentinel errors for profile operations. Callers may errors.Is against these
// or use utils.AsAppError for the canonical category/status.
var (
	ErrProfileNotFound     = errors.New("repositories: user profile not found")
	ErrProfileConflict     = errors.New("repositories: user profile conflict")
	ErrInvalidProfileInput = errors.New("repositories: invalid user profile input")
)

// UserProfile and UserProfilePatch are retained as aliases so existing callers
// continue to compile; the canonical definitions live in models.
type UserProfile = models.UserProfile
type UserProfilePatch = models.UserProfilePatch

// ProfileRepository provides persistence operations for user profiles.
type ProfileRepository struct {
	db *pgxpool.Pool
}

func NewProfileRepository(db *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{db: db}
}

const profileColumns = `user_id, COALESCE(bio, ''), skills, experience,
	COALESCE(resume_url, ''), updated_at`

// GetByUserID returns a user's profile. A missing row is reported as a
// canonical not-found AppError, while preserving pgx.ErrNoRows in the chain so
// callers that treat "no profile yet" as empty can keep doing so.
func (r *ProfileRepository) GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, newProfileValidationError("user_id is required")
	}

	q := fmt.Sprintf(`SELECT %s FROM user_profiles WHERE user_id = $1`, profileColumns)
	profile, err := scanUserProfile(r.db.QueryRow(ctx, q, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newProfileNotFoundError(userID, err)
		}
		return nil, newProfileDatabaseError("profile.get", userID, err)
	}
	return profile, nil
}

// GetProfileByUserID is the original name for GetByUserID.
//
// Deprecated: use GetByUserID.
func (r *ProfileRepository) GetProfileByUserID(ctx context.Context, userID string) (*models.UserProfile, error) {
	return r.GetByUserID(ctx, userID)
}

// Upsert inserts a profile or replaces it when the user already has one. The
// returned row is the persisted state, including the refreshed updated_at.
func (r *ProfileRepository) Upsert(ctx context.Context, profile *models.UserProfile) (*models.UserProfile, error) {
	if profile == nil {
		return nil, newProfileValidationError("profile is required")
	}
	userID := strings.TrimSpace(profile.UserID)
	if userID == "" {
		return nil, newProfileValidationError("user_id is required")
	}

	q := fmt.Sprintf(`
		INSERT INTO user_profiles (user_id, bio, skills, experience, resume_url, updated_at)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (user_id) DO UPDATE
		SET bio = EXCLUDED.bio,
		    skills = EXCLUDED.skills,
		    experience = EXCLUDED.experience,
		    resume_url = EXCLUDED.resume_url,
		    updated_at = now()
		RETURNING %s`, profileColumns)

	out, err := scanUserProfile(r.db.QueryRow(ctx, q,
		userID, profile.Bio, normalizeStrings(profile.Skills),
		normalizeStrings(profile.Experience), profile.ResumeURL))
	if err != nil {
		switch {
		case isUniqueViolation(err):
			return nil, newProfileConflictError(userID, err)
		case isForeignKeyViolation(err):
			// The user row does not exist (or was deleted concurrently).
			return nil, newProfileNotFoundError(userID, err)
		default:
			return nil, newProfileDatabaseError("profile.upsert", userID, err)
		}
	}
	return out, nil
}

// Update applies a partial update. Nil patch fields are left untouched; this
// allows a caller to change skills without clobbering bio, for example. An
// empty patch returns the current profile without bumping updated_at.
func (r *ProfileRepository) Update(ctx context.Context, userID string, patch models.UserProfilePatch) (*models.UserProfile, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, newProfileValidationError("user_id is required")
	}
	if patch.Empty() {
		return r.GetByUserID(ctx, userID)
	}

	var (
		bio        any
		skills     any
		experience any
		resumeURL  any
	)
	if patch.Bio != nil {
		bio = *patch.Bio
	}
	if patch.Skills != nil {
		skills = normalizeStrings(*patch.Skills)
	}
	if patch.Experience != nil {
		experience = normalizeStrings(*patch.Experience)
	}
	if patch.ResumeURL != nil {
		resumeURL = *patch.ResumeURL
	}

	q := fmt.Sprintf(`
		UPDATE user_profiles
		SET bio = COALESCE($2::text, bio),
		    skills = COALESCE($3::text[], skills),
		    experience = COALESCE($4::text[], experience),
		    resume_url = COALESCE($5::text, resume_url),
		    updated_at = now()
		WHERE user_id = $1
		RETURNING %s`, profileColumns)

	out, err := scanUserProfile(r.db.QueryRow(ctx, q,
		userID, bio, skills, experience, resumeURL))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newProfileNotFoundError(userID, err)
		}
		return nil, newProfileDatabaseError("profile.update", userID, err)
	}
	return out, nil
}

func scanUserProfile(row pgx.Row) (*models.UserProfile, error) {
	profile := &models.UserProfile{}
	err := row.Scan(&profile.UserID, &profile.Bio, &profile.Skills,
		&profile.Experience, &profile.ResumeURL, &profile.UpdatedAt)
	if err != nil {
		return nil, err
	}
	profile.Skills = normalizeStrings(profile.Skills)
	profile.Experience = normalizeStrings(profile.Experience)
	return profile, nil
}

// normalizeStrings turns a nil slice into an empty one so API responses always
// serialize skills/experience as [] rather than null, and writes never violate
// the NOT NULL array columns.
func normalizeStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}

func newProfileValidationError(message string) error {
	appErr := utils.NewValidationError(message, map[string]string{"operation": "profile.validate"})
	appErr.Err = fmt.Errorf("%w: %s", ErrInvalidProfileInput, message)
	return appErr
}

func newProfileNotFoundError(userID string, cause error) error {
	appErr := utils.NewNotFoundError("Profile not found.")
	appErr.Err = fmt.Errorf("%w: user %s: %w", ErrProfileNotFound, userID, cause)
	appErr.Context = map[string]string{"operation": "profile", "user_id": userID}
	return appErr
}

func newProfileConflictError(userID string, cause error) error {
	appErr := utils.NewConflictError("A profile with those details already exists.")
	appErr.Err = fmt.Errorf("%w: user %s: %w", ErrProfileConflict, userID, cause)
	appErr.Context = map[string]string{"operation": "profile.upsert", "user_id": userID}
	return appErr
}

func newProfileDatabaseError(operation, userID string, cause error) error {
	return utils.NewDatabaseError(cause, map[string]string{
		"operation": operation,
		"user_id":   userID,
	})
}
