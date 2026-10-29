
package models

type MatchingReport struct {
	OverallScore           float64        `json:"overall_score"`
	MatchedSkills          []string       `json:"matched_skills"`
	MissingMandatorySkills []string       `json:"missing_mandatory_skills"`
	MissingPreferredSkills []string       `json:"missing_preferred_skills"`
	KeywordSuggestions     []string       `json:"keyword_suggestions"`
	CategoryScores         CategoryScores `json:"category_scores"`
}

type CategoryScores struct {
	Skills     float64 `json:"skills"`
	Experience float64 `json:"experience"`
	Education  float64 `json:"education"`
	Keywords   float64 `json:"keywords"`
}
