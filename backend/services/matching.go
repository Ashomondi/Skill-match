package services

import (
	"math"
	"sort"
	"strings"

	"skill-match/backend/repositories"
)

type MatchingService struct{}

func NewMatchingService() *MatchingService {
	return &MatchingService{}
}

// ComputeScore calculates a matching score (0.0 to 1.0) based on user skills vs job title and description.
func (s *MatchingService) ComputeScore(userSkills []string, job *repositories.Job) float64 {
	if len(userSkills) == 0 || job == nil {
		return 0.0
	}

	jobText := strings.ToLower(job.Title + " " + job.Description)
	matchedSkills := 0

	for _, skill := range userSkills {
		cleanSkill := strings.TrimSpace(strings.ToLower(skill))
		if cleanSkill == "" {
			continue
		}
		if strings.Contains(jobText, cleanSkill) {
			matchedSkills++
		}
	}

	if matchedSkills == 0 {
		return 0.0
	}

	// Base score: ratio of matched skills
	score := float64(matchedSkills) / float64(len(userSkills))

	// Title bonus weight
	titleText := strings.ToLower(job.Title)
	for _, skill := range userSkills {
		cleanSkill := strings.TrimSpace(strings.ToLower(skill))
		if cleanSkill != "" && strings.Contains(titleText, cleanSkill) {
			score += 0.15
			break
		}
	}

	return math.Min(1.0, math.Round(score*100)/100)
}

// RankJobs scores a list of jobs, filters out scores below minScore, and sorts highest first.
func (s *MatchingService) RankJobs(userSkills []string, jobs []*repositories.Job, minScore float64) []*repositories.MatchScore {
	if len(userSkills) == 0 {
		return []*repositories.MatchScore{}
	}

	matches := make([]*repositories.MatchScore, 0)
	for _, job := range jobs {
		score := s.ComputeScore(userSkills, job)
		if score >= minScore {
			matches = append(matches, &repositories.MatchScore{
				Job:   job,
				Score: score,
			})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	return matches
}