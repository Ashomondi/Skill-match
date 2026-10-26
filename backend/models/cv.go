package models

import "time"

// CVSectionType enumerates the canonical section kinds stored in
// cv_sections.section_type. Keep in sync with the CHECK constraint in
// migrations/015_cv_documents.up.sql.
type CVSectionType string

const (
	CVSectionSummary    CVSectionType = "summary"
	CVSectionExperience CVSectionType = "experience"
	CVSectionEducation  CVSectionType = "education"
	CVSectionSkills     CVSectionType = "skills"
	CVSectionProjects   CVSectionType = "projects"
	CVSectionCustom     CVSectionType = "custom"
)

// Valid reports whether t is a supported section type.
func (t CVSectionType) Valid() bool {
	switch t {
	case CVSectionSummary, CVSectionExperience, CVSectionEducation,
		CVSectionSkills, CVSectionProjects, CVSectionCustom:
		return true
	default:
		return false
	}
}

// CVDocument is a structured CV owned by a user. The document itself holds
// presentation metadata; its content lives in ordered CVSection rows.
type CVDocument struct {
	ID          string         `json:"id"`
	UserID      string         `json:"user_id"`
	Title       string         `json:"title"`
	TargetRole  string         `json:"target_role"`
	TemplateID  string         `json:"template_id"`
	ThemeConfig map[string]any `json:"theme_config"`
	IsPrimary   bool           `json:"is_primary"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Sections    []*CVSection   `json:"sections,omitempty"`
}

// CVSection is one ordered block of a structured CV. Content is intentionally
// flexible JSONB: the shape for an "experience" section differs from a
// "skills" section, and custom sections may define their own fields.
type CVSection struct {
	ID          string         `json:"id"`
	CVID        string         `json:"cv_id"`
	SectionType CVSectionType  `json:"section_type"`
	Title       string         `json:"title"`
	SortOrder   int            `json:"sort_order"`
	Content     map[string]any `json:"content"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
