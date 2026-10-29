package services

import (
	"testing"

	"skill-match/backend/repositories"
)

func TestEvaluateMatchReturnsStructuredReport(t *testing.T) {
	service := NewMatchingService()

	job := &repositories.Job{
		Title: "Go Backend Developer",
		Description: `Required: Go and PostgreSQL.
Preferred: Docker.
Requires 3 years experience and a Bachelor's degree.`,
	}

	cv := `Backend developer with Go and PostgreSQL.
3 years experience. Bachelor's degree in Computer Science.`

	report, err := service.EvaluateMatch(cv, job)
	if err != nil {
		t.Fatalf("EvaluateMatch returned error: %v", err)
	}

	if report.OverallScore < 0 || report.OverallScore > 100 {
		t.Fatalf(
			"overall score out of range: %v",
			report.OverallScore,
		)
	}

	// Two of the three job skills are present in the CV:
	// Go and PostgreSQL are matched; Docker is missing.
	if report.CategoryScores.Skills != 66.67 {
		t.Errorf(
			"skills score = %v, want 66.67",
			report.CategoryScores.Skills,
		)
	}

	if !containsSkill(report.MatchedSkills, "Go") {
		t.Errorf(
			"expected Go in matched skills: %v",
			report.MatchedSkills,
		)
	}

	if !containsSkill(report.MatchedSkills, "PostgreSQL") {
		t.Errorf(
			"expected PostgreSQL in matched skills: %v",
			report.MatchedSkills,
		)
	}

	if !containsSkill(report.MissingPreferredSkills, "Docker") {
		t.Errorf(
			"expected Docker in missing preferred skills: %v",
			report.MissingPreferredSkills,
		)
	}

	if len(report.MissingMandatorySkills) != 0 {
		t.Errorf(
			"expected no missing mandatory skills, got: %v",
			report.MissingMandatorySkills,
		)
	}
}

func TestEvaluateMatchRejectsEmptyCV(t *testing.T) {
	service := NewMatchingService()

	job := &repositories.Job{
		Title:       "Go Developer",
		Description: "Required: Go",
	}

	_, err := service.EvaluateMatch("  ", job)
	if err == nil {
		t.Fatal("expected an error for empty CV")
	}
}

func TestEvaluateMatchRejectsEmptyJobDescription(t *testing.T) {
	service := NewMatchingService()

	job := &repositories.Job{
		Title: "Go Developer",
	}

	_, err := service.EvaluateMatch(
		"Experienced Go developer",
		job,
	)

	if err == nil {
		t.Fatal("expected an error for empty job description")
	}
}

func TestEvaluateMatchRejectsNilJob(t *testing.T) {
	service := NewMatchingService()

	_, err := service.EvaluateMatch(
		"Experienced Go developer",
		nil,
	)

	if err == nil {
		t.Fatal("expected an error for nil job")
	}
}

func TestEvaluateMatchIsDeterministic(t *testing.T) {
	service := NewMatchingService()

	job := &repositories.Job{
		Title:       "Backend Developer",
		Description: "Required: Go and SQL. Preferred: Docker. 2 years experience.",
	}

	cv := "Go developer with SQL and 2 years experience."

	first, err := service.EvaluateMatch(cv, job)
	if err != nil {
		t.Fatalf("first evaluation returned error: %v", err)
	}

	second, err := service.EvaluateMatch(cv, job)
	if err != nil {
		t.Fatalf("second evaluation returned error: %v", err)
	}

	if first.OverallScore != second.OverallScore {
		t.Fatalf(
			"scores differ: %v and %v",
			first.OverallScore,
			second.OverallScore,
		)
	}

	if first.CategoryScores != second.CategoryScores {
		t.Fatalf(
			"category scores differ: %+v and %+v",
			first.CategoryScores,
			second.CategoryScores,
		)
	}
}

func TestContainsWholePhraseAvoidsSubstringFalsePositive(t *testing.T) {
	if containsWholePhrase("Django developer", "Go") {
		t.Fatal("Go must not match inside Django")
	}

	if !containsWholePhrase("Go developer", "Go") {
		t.Fatal("expected Go to match as a complete word")
	}
}
