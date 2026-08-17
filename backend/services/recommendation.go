package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"skill-match/backend/models"
	"skill-match/backend/repositories"
)

var (
	ErrInvalidRecommendationInput = errors.New("invalid recommendation input")
	ErrRecommendationUnavailable  = errors.New("recommendation unavailable")
)

const (
	defaultRecommendationLimit = 10
	maxRecommendationLimit     = 50
	candidateMultiplier        = 5
)

type recommendationUserRepository interface {
	GetByID(ctx context.Context, id string) (*repositories.User, error)
}

type recommendationResumeRepository interface {
	GetByID(ctx context.Context, id string) (*models.Resume, error)
	ListByUserID(ctx context.Context, userID string, limit int) ([]*models.Resume, error)
}

type recommendationEmbeddingRepository interface {
	GetBySource(ctx context.Context, sourceType repositories.EmbeddingSourceType, sourceID string) (*repositories.Embedding, error)
	FindSimilarJobs(ctx context.Context, queryVector []float32, k int) ([]repositories.SimilarJob, error)
}

type recommendationJobRepository interface {
	GetByIDs(ctx context.Context, ids []string) ([]*repositories.Job, error)
}

type recommendationInteractionRepository interface {
	InteractedJobIDs(ctx context.Context, userID string) (map[string]bool, error)
}

type RecommendationService struct {
	users        recommendationUserRepository
	resumes      recommendationResumeRepository
	embeddings   recommendationEmbeddingRepository
	jobs         recommendationJobRepository
	interactions recommendationInteractionRepository
}

type NewRecommendationServiceInput struct {
	Users        recommendationUserRepository
	Resumes      recommendationResumeRepository
	Embeddings   recommendationEmbeddingRepository
	Jobs         recommendationJobRepository
	Interactions recommendationInteractionRepository
}

func NewRecommendationService(input NewRecommendationServiceInput) *RecommendationService {
	return &RecommendationService{
		users:        input.Users,
		resumes:      input.Resumes,
		embeddings:   input.Embeddings,
		jobs:         input.Jobs,
		interactions: input.Interactions,
	}
}

type RecommendationPreferences struct {
	Skills            []string `json:"skills,omitempty"`
	Keywords          []string `json:"keywords,omitempty"`
	Locations         []string `json:"locations,omitempty"`
	ExcludedCompanies []string `json:"excluded_companies,omitempty"`
	RemoteOnly        bool     `json:"remote_only,omitempty"`
}

type RecommendationRequest struct {
	UserID      string                    `json:"user_id"`
	ResumeID    string                    `json:"resume_id,omitempty"`
	Limit       int                       `json:"limit,omitempty"`
	Preferences RecommendationPreferences `json:"preferences,omitempty"`
}

type JobRecommendation struct {
	Job     *repositories.Job `json:"job"`
	Score   float64           `json:"score"`
	Reasons []string          `json:"reasons,omitempty"`
}

func (s *RecommendationService) Recommend(ctx context.Context, request RecommendationRequest) ([]JobRecommendation, error) {
	request.UserID = strings.TrimSpace(request.UserID)
	if request.UserID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidRecommendationInput)
	}
	if err := s.validateDependencies(); err != nil {
		return nil, err
	}

	user, err := s.users.GetByID(ctx, request.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: retrieve user profile: %v", ErrRecommendationUnavailable, err)
	}
	if user == nil || !user.IsActive {
		return nil, fmt.Errorf("%w: user profile is inactive or missing", ErrRecommendationUnavailable)
	}

	resume, err := s.resumeForRequest(ctx, request)
	if err != nil {
		return nil, err
	}
	embedding, err := s.embeddings.GetBySource(ctx, repositories.EmbeddingSourceResume, resume.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: retrieve resume embedding: %v", ErrRecommendationUnavailable, err)
	}

	interacted, err := s.interactions.InteractedJobIDs(ctx, request.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: retrieve previous job interactions: %v", ErrRecommendationUnavailable, err)
	}

	limit := normalizeRecommendationLimit(request.Limit)
	matches, err := s.embeddings.FindSimilarJobs(ctx, embedding.Vector, limit*candidateMultiplier)
	if err != nil {
		return nil, fmt.Errorf("%w: retrieve relevant job matches: %v", ErrRecommendationUnavailable, err)
	}

	distanceByID := make(map[string]float64, len(matches))
	jobIDs := make([]string, 0, len(matches))
	for _, match := range matches {
		if match.JobID == "" || interacted[match.JobID] {
			continue
		}
		if _, exists := distanceByID[match.JobID]; exists {
			continue
		}
		distanceByID[match.JobID] = match.Distance
		jobIDs = append(jobIDs, match.JobID)
	}

	jobs, err := s.jobs.GetByIDs(ctx, jobIDs)
	if err != nil {
		return nil, fmt.Errorf("%w: retrieve matched jobs: %v", ErrRecommendationUnavailable, err)
	}

	recommendations := make([]JobRecommendation, 0, len(jobs))
	for _, job := range jobs {
		recommendation, include := rankJob(job, distanceByID[job.ID], request.Preferences)
		if include {
			recommendations = append(recommendations, recommendation)
		}
	}

	sort.SliceStable(recommendations, func(i, j int) bool {
		if recommendations[i].Score == recommendations[j].Score {
			return recommendations[i].Job.ID < recommendations[j].Job.ID
		}
		return recommendations[i].Score > recommendations[j].Score
	})
	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}
	return recommendations, nil
}

func (s *RecommendationService) validateDependencies() error {
	if s == nil || s.users == nil || s.resumes == nil || s.embeddings == nil || s.jobs == nil || s.interactions == nil {
		return fmt.Errorf("%w: recommendation service is not fully configured", ErrRecommendationUnavailable)
	}
	return nil
}

func (s *RecommendationService) resumeForRequest(ctx context.Context, request RecommendationRequest) (*models.Resume, error) {
	resumeID := strings.TrimSpace(request.ResumeID)
	if resumeID != "" {
		resume, err := s.resumes.GetByID(ctx, resumeID)
		if err != nil {
			return nil, fmt.Errorf("%w: retrieve resume: %v", ErrRecommendationUnavailable, err)
		}
		if resume.UserID != request.UserID {
			return nil, ErrResumeUnauthorized
		}
		if !usableRecommendationResume(resume) {
			return nil, fmt.Errorf("%w: resume has not been parsed", ErrRecommendationUnavailable)
		}
		return resume, nil
	}

	resumes, err := s.resumes.ListByUserID(ctx, request.UserID, 20)
	if err != nil {
		return nil, fmt.Errorf("%w: retrieve resumes: %v", ErrRecommendationUnavailable, err)
	}
	for _, resume := range resumes {
		if usableRecommendationResume(resume) {
			return resume, nil
		}
	}
	return nil, fmt.Errorf("%w: no parsed resume is available", ErrRecommendationUnavailable)
}

func usableRecommendationResume(resume *models.Resume) bool {
	return resume != nil && resume.Status == models.ResumeStatusParsed &&
		resume.ParsedText != nil && strings.TrimSpace(*resume.ParsedText) != ""
}

func normalizeRecommendationLimit(limit int) int {
	if limit <= 0 {
		return defaultRecommendationLimit
	}
	if limit > maxRecommendationLimit {
		return maxRecommendationLimit
	}
	return limit
}

func rankJob(job *repositories.Job, distance float64, preferences RecommendationPreferences) (JobRecommendation, bool) {
	if job == nil || excludedJob(job, preferences) {
		return JobRecommendation{}, false
	}

	// Cosine distance is converted to a bounded relevance score before
	// preference signals are applied.
	baseScore := 1 / (1 + math.Max(distance, 0))
	score := baseScore
	reasons := []string{"Relevant to resume experience"}
	searchable := strings.ToLower(strings.Join([]string{job.Title, job.Company, job.Location, job.Description}, " "))

	if matchesAny(searchable, preferences.Skills) {
		score += 0.12
		reasons = append(reasons, "Matches requested skills")
	}
	if matchesAny(searchable, preferences.Keywords) {
		score += 0.08
		reasons = append(reasons, "Matches job preferences")
	}
	if len(preferences.Locations) > 0 && matchesAny(strings.ToLower(job.Location), preferences.Locations) {
		score += 0.05
		reasons = append(reasons, "Matches preferred location")
	}
	if job.Remote && preferences.RemoteOnly {
		score += 0.05
		reasons = append(reasons, "Remote role")
	}

	return JobRecommendation{Job: job, Score: math.Min(score, 1), Reasons: reasons}, true
}

func excludedJob(job *repositories.Job, preferences RecommendationPreferences) bool {
	if preferences.RemoteOnly && !job.Remote {
		return true
	}
	company := strings.ToLower(strings.TrimSpace(job.Company))
	for _, excluded := range preferences.ExcludedCompanies {
		if company == strings.ToLower(strings.TrimSpace(excluded)) {
			return true
		}
	}
	return false
}

func matchesAny(text string, values []string) bool {
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value != "" && strings.Contains(text, value) {
			return true
		}
	}
	return false
}
