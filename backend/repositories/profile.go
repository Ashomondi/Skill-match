package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserProfile struct {
	UserID     string    `json:"user_id"`
	Bio        string    `json:"bio"`
	Skills     []string  `json:"skills"`
	Experience []string  `json:"experience"`
	ResumeURL  string    `json:"resume_url"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ProfileRepository struct {
	db *pgxpool.Pool
}

func NewProfileRepository(db *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) GetProfileByUserID(ctx context.Context, userID string) (*UserProfile, error) {
	query := `
		SELECT user_id, COALESCE(bio, ''), skills, experience, COALESCE(resume_url, ''), updated_at
		FROM user_profiles
		WHERE user_id = $1
	`

	p := &UserProfile{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&p.UserID, &p.Bio, &p.Skills, &p.Experience, &p.ResumeURL, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("repositories: get profile by user id: %w", err)
	}

	return p, nil
}
