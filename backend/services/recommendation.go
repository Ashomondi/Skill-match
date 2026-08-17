package services

import (
	"context"
	"fmt"

	"skill-match/backend/repositories"
)

type RecommendationService struct {
	jobRepo     *repositories.JobRepository
	profileRepo *repositories.ProfileRepository
}

func NewRecommendationService(
	jobRepo *repositories.JobRepository,
	profileRepo *repositories.ProfileRepository,
) *RecommendationService {
	return &RecommendationService{
		jobRepo:     jobRepo,
		profileRepo: profileRepo,
	}
}

type MatchingContext struct {
	UserID     string   `json:"user_id"`
	UserSkills []string `json:"user_skills"`
	MinScore   float64  `json:"min_score"`
	Limit      int      `json:"limit"`
}

func (s *RecommendationService) RecommendForUser(ctx context.Context, requestingUserID string, targetUserID string, limit int) ([]*repositories.MatchScore, error) {
	if requestingUserID != targetUserID {
		return nil, fmt.Errorf("services: unauthorized access to user profile data")
	}

	profile, err := s.profileRepo.GetProfileByUserID(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("services: fetch user resume/profile: %w", err)
	}

	matchingSkills := append([]string{}, profile.Skills...)
	matchingSkills = append(matchingSkills, profile.Experience...)

	filter := repositories.SemanticMatchFilter{
		UserSkills: matchingSkills,
		MinScore:   0.1,
		Limit:      limit,
	}

	recommendations, err := s.jobRepo.MatchJobs(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("services: recommend jobs match: %w", err)
	}

	return recommendations, nil
}