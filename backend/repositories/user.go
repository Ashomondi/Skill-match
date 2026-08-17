// Package repositories contains the persistence layer. Repositories are the
// only layer permitted to issue SQL against CockroachDB; callers (services)
// depend on the exported methods here and never touch *pgxpool.Pool
// directly.
package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Sentinel errors. Services should compare against these with errors.Is
// rather than inspecting driver-specific error types.
var (
	ErrUserNotFound     = errors.New("repositories: user not found")
	ErrUserEmailTaken   = errors.New("repositories: email already registered")
	ErrInvalidUserInput = errors.New("repositories: invalid user input")
)

// User is the persistence-layer representation of a user row.
//
// NOTE: models.User (Issue 3, owned by Ashley) is expected to become the
// canonical type once it lands. This struct mirrors the schema in
// migrations/001_initial_schema.sql and should be replaced by a type alias
// to models.User (or the repository methods updated to accept/return it) as
// soon as that file exists, to avoid two divergent definitions.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"fullName"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// UserRepository provides persistence operations for users backed by
// CockroachDB.
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository constructs a UserRepository. db must be a live,
// pinged connection pool; construction does not verify connectivity.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user and returns the row as persisted (including
// generated ID and timestamps). Email is normalized to lowercase to match
// the schema's case-insensitivity constraint.
//
// Returns ErrUserEmailTaken if the email is already registered.
func (r *UserRepository) Create(ctx context.Context, u *User) (*User, error) {
	if u == nil || strings.TrimSpace(u.Email) == "" || u.PasswordHash == "" {
		return nil, fmt.Errorf("%w: email and password_hash are required", ErrInvalidUserInput)
	}

	const q = `
		INSERT INTO users (email, password_hash, full_name)
		VALUES ($1, $2, $3)
		RETURNING id, email, password_hash, full_name, is_active, created_at, updated_at`

	email := strings.ToLower(strings.TrimSpace(u.Email))

	row := r.db.QueryRow(ctx, q, email, u.PasswordHash, u.FullName)

	out := &User{}
	if err := row.Scan(&out.ID, &out.Email, &out.PasswordHash, &out.FullName,
		&out.IsActive, &out.CreatedAt, &out.UpdatedAt); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrUserEmailTaken
		}
		return nil, fmt.Errorf("repositories: create user: %w", err)
	}

	return out, nil
}

// GetByID fetches a user by primary key. Returns ErrUserNotFound if no row
// matches.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*User, error) {
	const q = `
		SELECT id, email, password_hash, full_name, is_active, created_at, updated_at
		FROM users
		WHERE id = $1`

	return r.scanOne(ctx, q, id)
}

// GetByEmail fetches a user by email (case-insensitive). Returns
// ErrUserNotFound if no row matches. This is the primary lookup used by
// the auth login flow.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	const q = `
		SELECT id, email, password_hash, full_name, is_active, created_at, updated_at
		FROM users
		WHERE email = $1`

	return r.scanOne(ctx, q, strings.ToLower(strings.TrimSpace(email)))
}

// UpdatePassword updates a user's password hash and bumps updated_at.
// Returns ErrUserNotFound if no row matches id.
func (r *UserRepository) UpdatePassword(ctx context.Context, id, newPasswordHash string) error {
	if newPasswordHash == "" {
		return fmt.Errorf("%w: password_hash is required", ErrInvalidUserInput)
	}

	const q = `
		UPDATE users
		SET password_hash = $1, updated_at = now()
		WHERE id = $2`

	tag, err := r.db.Exec(ctx, q, newPasswordHash, id)
	if err != nil {
		return fmt.Errorf("repositories: update password: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// SetActive enables or disables a user account (soft-disable, not delete).
// Returns ErrUserNotFound if no row matches id.
func (r *UserRepository) SetActive(ctx context.Context, id string, active bool) error {
	const q = `
		UPDATE users
		SET is_active = $1, updated_at = now()
		WHERE id = $2`

	tag, err := r.db.Exec(ctx, q, active, id)
	if err != nil {
		return fmt.Errorf("repositories: set active: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// Delete permanently removes a user. Prefer SetActive(ctx, id, false) for
// account deactivation; this is a hard delete and should only be used for
// GDPR-style erasure requests.
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM users WHERE id = $1`

	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("repositories: delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// scanOne runs a single-row query and maps it to a *User, translating
// pgx.ErrNoRows into the package's sentinel ErrUserNotFound so callers never
// need to import pgx directly.
func (r *UserRepository) scanOne(ctx context.Context, query string, args ...any) (*User, error) {
	row := r.db.QueryRow(ctx, query, args...)

	u := &User{}
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	switch {
	case err == nil:
		return u, nil
	case errors.Is(err, pgx.ErrNoRows):
		return nil, ErrUserNotFound
	default:
		return nil, fmt.Errorf("repositories: query user: %w", err)
	}
}

// isUniqueViolation reports whether err is a CockroachDB/Postgres unique
// constraint violation (SQLSTATE 23505), e.g. duplicate email on insert.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
