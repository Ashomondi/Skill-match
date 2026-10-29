package services

import (
	"errors"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"skill-match/backend/models"
	"skill-match/backend/repositories"
)

type MatchingService struct{}

func NewMatchingService() *MatchingService {
	return &MatchingService{}
}

var skillDictionary = []string{
	"JavaScript", "TypeScript", "Node.js", "React", "Next.js",
	"Vite", "HTML", "CSS", "Go", "Golang", "Python", "Java",
	"C++", "C#", "SQL", "PostgreSQL", "MySQL", "MongoDB",
	"Redis", "Docker", "Kubernetes", "AWS", "Azure", "GCP",
	"Git", "Linux", "REST API", "GraphQL", "PHP", "Ruby",
	"Rust", "C", "Django", "Flask", "FastAPI", "Express",
	"Spring", "Gin", "Machine Learning", "Artificial Intelligence",
	"Agile", "Scrum", "CI/CD", "Terraform", "GitHub Actions",
}

var preferredPattern = regexp.MustCompile(
	`(?i)\b(preferred|preferably|nice to have|desirable|bonus|an advantage|plus)\b`,
)

var mandatoryPattern = regexp.MustCompile(
	`(?i)\b(required|requirement|mandatory|must have|essential|must|required skills)\b`,
)

var yearsPattern = regexp.MustCompile(
	`(?i)(\d+)\s*\+?\s*(?:years?|yrs?)\s+(?:of\s+)?(?:relevant\s+)?(?:work\s+)?experience`,
)

var anyYearsPattern = regexp.MustCompile(
	`(?i)(\d+)\s*\+?\s*(?:years?|yrs?)`,
)

var keywordPattern = regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9+#./-]*`)

var stopWords = map[string]bool{
	"a": true, "an": true, "and": true, "are": true, "as": true,
	"at": true, "be": true, "by": true, "for": true, "from": true,
	"in": true, "is": true, "it": true, "of": true, "on": true,
	"or": true, "our": true, "that": true, "the": true, "their": true,
	"this": true, "to": true, "we": true, "with": true, "you": true,
	"will": true, "work": true, "working": true, "role": true,
	"job": true, "candidate": true, "skills": true, "experience": true,
	"years": true, "year": true, "ability": true, "knowledge": true,
	"including": true, "using": true, "strong": true, "good": true,
	"have": true, "required": true, "preferred": true,
}

// EvaluateMatch returns a reproducible, explainable CV-to-job evaluation.
// CategoryScores and OverallScore use a 0-100 scale.
func (s *MatchingService) EvaluateMatch(
	cvText string,
	job *repositories.Job,
) (*models.MatchingReport, error) {
	cvText = strings.TrimSpace(cvText)

	if cvText == "" {
		return nil, errors.New("CV text must not be empty")
	}
	if job == nil {
		return nil, errors.New("job must not be nil")
	}
	if strings.TrimSpace(job.Description) == "" {
		return nil, errors.New("job description must not be empty")
	}

	jobText := strings.TrimSpace(job.Title + " " + job.Description)
	cvLower := strings.ToLower(cvText)
	jobLower := strings.ToLower(jobText)

	requiredSkills, preferredSkills := extractJobSkills(jobText)
	cvSkills := extractSkills(cvText)

	matchedSkills := make([]string, 0)
	missingMandatory := make([]string, 0)
	missingPreferred := make([]string, 0)

	mandatoryMatched := 0
	for _, skill := range requiredSkills {
		if containsSkill(cvSkills, skill) {
			matchedSkills = append(matchedSkills, skill)
			mandatoryMatched++
		} else {
			missingMandatory = append(missingMandatory, skill)
		}
	}

	preferredMatched := 0
	for _, skill := range preferredSkills {
		if containsSkill(cvSkills, skill) {
			matchedSkills = appendUnique(matchedSkills, skill)
			preferredMatched++
		} else {
			missingPreferred = append(missingPreferred, skill)
		}
	}

	allJobSkills := append([]string{}, requiredSkills...)
	allJobSkills = append(allJobSkills, preferredSkills...)

	skillsScore := 100.0
	if len(allJobSkills) > 0 {
		skillsScore = float64(mandatoryMatched+preferredMatched) /
			float64(len(allJobSkills)) * 100
	}

	experienceScore := calculateExperienceScore(cvText, jobText)
	educationScore := calculateEducationScore(cvLower, jobLower)
	keywordScore, keywordSuggestions := calculateKeywordScore(cvText, jobText)

	overall := skillsScore*0.40 +
		experienceScore*0.30 +
		educationScore*0.15 +
		keywordScore*0.15

	sort.Strings(matchedSkills)
	sort.Strings(missingMandatory)
	sort.Strings(missingPreferred)
	sort.Strings(keywordSuggestions)

	return &models.MatchingReport{
		OverallScore:           roundScore(overall),
		MatchedSkills:          matchedSkills,
		MissingMandatorySkills: missingMandatory,
		MissingPreferredSkills: missingPreferred,
		KeywordSuggestions:     keywordSuggestions,
		CategoryScores: models.CategoryScores{
			Skills:     roundScore(skillsScore),
			Experience: roundScore(experienceScore),
			Education:  roundScore(educationScore),
			Keywords:   roundScore(keywordScore),
		},
	}, nil
}

// extractJobSkills classifies skills mentioned in the job description.
// Explicitly preferred skills are kept separate unless also required.
func extractJobSkills(jobText string) ([]string, []string) {
	classification := make(map[string]string)

	for _, sentence := range splitSentences(jobText) {
		lowerSentence := strings.ToLower(sentence)
		preferred := preferredPattern.MatchString(lowerSentence)
		mandatory := mandatoryPattern.MatchString(lowerSentence)

		for _, skill := range extractSkills(sentence) {
			key := strings.ToLower(skill)

			switch {
			case mandatory:
				classification[key] = "mandatory"
			case preferred:
				if classification[key] != "mandatory" {
					classification[key] = "preferred"
				}
			case classification[key] == "":
				// Skills without an explicit preference label are
				// treated as required by default.
				classification[key] = "mandatory"
			}
		}
	}

	required := make([]string, 0)
	preferred := make([]string, 0)

	for _, skill := range skillDictionary {
		switch classification[strings.ToLower(skill)] {
		case "mandatory":
			required = append(required, skill)
		case "preferred":
			preferred = append(preferred, skill)
		}
	}

	return required, preferred
}

func splitSentences(text string) []string {
	// Split on sentence-ending punctuation followed by whitespace,
	// or on newlines. This avoids splitting terms such as Node.js.
	re := regexp.MustCompile(`[.!?;]\s+|\n+`)
	parts := re.Split(text, -1)

	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			result = append(result, part)
		}
	}
	return result
}

func extractSkills(text string) []string {
	found := make([]string, 0)

	for _, skill := range skillDictionary {
		if containsWholePhrase(text, skill) {
			found = append(found, skill)
		}
	}

	return found
}

func containsSkill(skills []string, target string) bool {
	for _, skill := range skills {
		if strings.EqualFold(skill, target) {
			return true
		}
	}
	return false
}

func appendUnique(items []string, value string) []string {
	for _, item := range items {
		if strings.EqualFold(item, value) {
			return items
		}
	}
	return append(items, value)
}

func containsWholePhrase(text, phrase string) bool {
	text = strings.ToLower(text)
	phrase = strings.ToLower(strings.TrimSpace(phrase))

	if phrase == "" {
		return false
	}

	pattern := `(^|[^a-z0-9_])` +
		regexp.QuoteMeta(phrase) +
		`($|[^a-z0-9_])`

	matched, err := regexp.MatchString(pattern, text)
	return err == nil && matched
}

func calculateExperienceScore(cvText, jobText string) float64 {
	requiredYears := extractYears(jobText, yearsPattern)
	if requiredYears == 0 {
		requiredYears = extractYears(jobText, anyYearsPattern)
	}

	// No explicit experience requirement means this category
	// does not penalize the candidate.
	if requiredYears == 0 {
		return 100
	}

	candidateYears := extractYears(cvText, yearsPattern)
	if candidateYears == 0 {
		candidateYears = extractYears(cvText, anyYearsPattern)
	}
	if candidateYears == 0 {
		return 0
	}

	score := float64(candidateYears) / float64(requiredYears) * 100
	return math.Min(score, 100)
}

func extractYears(text string, pattern *regexp.Regexp) int {
	matches := pattern.FindAllStringSubmatch(text, -1)

	maxYears := 0
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		years, err := strconv.Atoi(match[1])
		if err == nil && years > maxYears {
			maxYears = years
		}
	}
	return maxYears
}

func calculateEducationScore(cvText, jobText string) float64 {
	requiredLevel := educationLevel(jobText)

	// No explicit education requirement means neutral/full credit.
	if requiredLevel == 0 {
		return 100
	}

	candidateLevel := educationLevel(cvText)
	if candidateLevel >= requiredLevel {
		return 100
	}
	return 0
}

func educationLevel(text string) int {
	text = strings.ToLower(text)

	levels := []struct {
		level int
		terms []string
	}{
		{5, []string{"phd", "ph.d", "doctorate", "doctoral degree"}},
		{4, []string{"master's", "masters", "master of", "msc", "mba"}},
		{3, []string{"bachelor's", "bachelors", "bachelor of", "bsc", "b.eng"}},
		{2, []string{"associate degree", "higher diploma", "diploma"}},
		{1, []string{"certificate", "certification"}},
	}

	for _, item := range levels {
		for _, term := range item.terms {
			if strings.Contains(text, term) {
				return item.level
			}
		}
	}
	return 0
}

func calculateKeywordScore(cvText, jobText string) (float64, []string) {
	cvKeywords := extractKeywords(cvText)
	jobKeywords := extractKeywords(jobText)

	if len(jobKeywords) == 0 {
		return 100, []string{}
	}

	matched := 0
	suggestions := make([]string, 0)

	for keyword := range jobKeywords {
		if cvKeywords[keyword] {
			matched++
		} else {
			suggestions = append(suggestions, keyword)
		}
	}

	score := float64(matched) / float64(len(jobKeywords)) * 100
	return score, suggestions
}

func extractKeywords(text string) map[string]bool {
	keywords := make(map[string]bool)

	for _, token := range keywordPattern.FindAllString(strings.ToLower(text), -1) {
		token = strings.Trim(token, "./-")
		if len(token) < 2 || stopWords[token] {
			continue
		}
		keywords[token] = true
	}

	return keywords
}

func roundScore(score float64) float64 {
	return math.Round(score*100) / 100
}

// ComputeScore preserves the existing 0-1 scoring interface.
func (s *MatchingService) ComputeScore(
	userSkills []string,
	job *repositories.Job,
) float64 {
	if job == nil {
		return 0
	}

	validSkills := make([]string, 0, len(userSkills))
	for _, skill := range userSkills {
		if strings.TrimSpace(skill) != "" {
			validSkills = append(validSkills, strings.TrimSpace(skill))
		}
	}
	if len(validSkills) == 0 {
		return 0
	}

	jobText := job.Title + " " + job.Description
	matched := 0
	titleMatch := false

	for _, skill := range validSkills {
		if containsWholePhrase(jobText, skill) {
			matched++
			if containsWholePhrase(job.Title, skill) {
				titleMatch = true
			}
		}
	}

	if matched == 0 {
		return 0
	}

	score := float64(matched) / float64(len(validSkills))
	if titleMatch {
		score += 0.15
	}
	return math.Min(1, math.Round(score*100)/100)
}

// RankJobs preserves the existing ranking interface.
func (s *MatchingService) RankJobs(
	userSkills []string,
	jobs []*repositories.Job,
	minScore float64,
) []*repositories.MatchScore {
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
