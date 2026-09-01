package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"skill-match/backend/repositories"
)

type RecommendationService struct {
	jobRepo     *repositories.JobRepository
	profileRepo *repositories.ProfileRepository
	matchingSvc *MatchingService
}

func NewRecommendationService(
	jobRepo *repositories.JobRepository,
	profileRepo *repositories.ProfileRepository,
	matchingSvc *MatchingService,
) *RecommendationService {
	return &RecommendationService{
		jobRepo:     jobRepo,
		profileRepo: profileRepo,
		matchingSvc: matchingSvc,
	}
}

type MatchingContext struct {
	UserID     string   `json:"user_id"`
	UserSkills []string `json:"user_skills"`
	MinScore   float64  `json:"min_score"`
	Limit      int      `json:"limit"`
}

func (s *RecommendationService) RecommendForUser(
	ctx context.Context,
	requestingUserID string,
	targetUserID string,
	limit int,
) ([]*repositories.MatchScore, error) {
	if requestingUserID != targetUserID {
		return nil, fmt.Errorf("services: unauthorized access to user profile data")
	}

	profile, err := s.profileRepo.GetProfileByUserID(ctx, targetUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*repositories.MatchScore{}, nil
		}
		return nil, fmt.Errorf("services: fetch user resume/profile: %w", err)
	}

	matchingSkills := append([]string{}, profile.Skills...)
	matchingSkills = append(matchingSkills, profile.Experience...)

	filter := repositories.SemanticMatchFilter{
		UserSkills: matchingSkills,
		MinScore:   0.1,
		Limit:      limit,
	}

	// 1. Fetch candidate jobs from repository layer
	candidateJobs, err := s.jobRepo.MatchJobs(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("services: recommend jobs match: %w", err)
	}

	// 2. Score and rank jobs using MatchingService within the service layer
	recommendations := s.matchingSvc.RankJobs(matchingSkills, candidateJobs, filter.MinScore)

	return recommendations, nil
}
