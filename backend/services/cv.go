package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"skill-match/backend/models"
	"skill-match/backend/repositories"
	"skill-match/backend/utils"
)

// Structured CV payload limits.
const (
	MaxCVTitleLength        = 200
	MaxCVTargetRoleLength   = 200
	MaxCVTemplateIDLength   = 100
	MaxCVSections           = 100
	MaxCVSectionTitleLength = 200
	MaxCVContentBytes       = 128 << 10
	MaxCVThemeConfigBytes   = 64 << 10
	MaxCVListLimit          = 100
)

// CVStore is the persistence subset the CV service needs.
// repositories.CVDocumentRepository satisfies this interface.
type CVStore interface {
	Create(ctx context.Context, doc *models.CVDocument) (*models.CVDocument, error)
	GetByIDForUser(ctx context.Context, userID, id string) (*models.CVDocument, error)
	ListByUserID(ctx context.Context, userID string, limit int) ([]*models.CVDocument, error)
	Update(ctx context.Context, doc *models.CVDocument) (*models.CVDocument, error)
	UpdateWithSections(ctx context.Context, doc *models.CVDocument) (*models.CVDocument, error)
	DeleteForUser(ctx context.Context, userID, id string) error
}

// CVResumeLookup is the resume read the CV service needs for optional import.
type CVResumeLookup interface {
	GetByID(ctx context.Context, id string) (*models.Resume, error)
}

// CVSectionInput is a client-supplied section. Sort order is derived from the
// position in the request, never trusted from the client.
type CVSectionInput struct {
	ID          string
	SectionType models.CVSectionType
	Title       string
	Content     map[string]any
}

// CreateCVInput describes a new structured CV.
type CreateCVInput struct {
	Title       string
	TargetRole  string
	TemplateID  string
	ThemeConfig map[string]any
	IsPrimary   bool
	ResumeID    string
	Sections    []CVSectionInput
}

// UpdateCVInput is a partial metadata update. Nil fields are left unchanged.
type UpdateCVInput struct {
	Title       *string
	TargetRole  *string
	TemplateID  *string
	ThemeConfig *map[string]any
	IsPrimary   *bool
}

// Empty reports whether the update requests no changes.
func (u UpdateCVInput) Empty() bool {
	return u.Title == nil && u.TargetRole == nil && u.TemplateID == nil &&
		u.ThemeConfig == nil && u.IsPrimary == nil
}

// CVService owns structured CV validation, ownership scoping, and optional
// resume import.
type CVService struct {
	store   CVStore
	resumes CVResumeLookup
}

func NewCVService(store CVStore, resumes CVResumeLookup) *CVService {
	return &CVService{store: store, resumes: resumes}
}

// List returns a user's CVs, most recently updated first.
func (s *CVService) List(ctx context.Context, userID string, limit int) ([]*models.CVDocument, error) {
	if err := validateCVUser(userID); err != nil {
		return nil, err
	}
	if limit < 0 {
		limit = 0
	}
	if limit > MaxCVListLimit {
		limit = MaxCVListLimit
	}

	docs, err := s.store.ListByUserID(ctx, userID, limit)
	if err != nil {
		return nil, cvDatabaseError("cv.list", err)
	}
	if docs == nil {
		docs = []*models.CVDocument{}
	}
	return docs, nil
}

// Create creates a CV, optionally seeding a summary section from a parsed
// resume owned by the same user.
func (s *CVService) Create(ctx context.Context, userID string, input CreateCVInput) (*models.CVDocument, error) {
	if err := validateCVUser(userID); err != nil {
		return nil, err
	}

	doc, err := s.buildCVDocument(ctx, userID, input)
	if err != nil {
		return nil, err
	}

	created, err := s.store.Create(ctx, doc)
	if err != nil {
		if errors.Is(err, repositories.ErrInvalidCVInput) {
			return nil, cvValidationError("invalid CV data")
		}
		return nil, cvDatabaseError("cv.create", err)
	}
	return created, nil
}

// Get returns a CV and its ordered sections, scoped to the owner.
func (s *CVService) Get(ctx context.Context, userID, id string) (*models.CVDocument, error) {
	if err := validateCVUser(userID); err != nil {
		return nil, err
	}
	if err := validateCVID(id); err != nil {
		return nil, err
	}

	doc, err := s.store.GetByIDForUser(ctx, userID, id)
	if err != nil {
		if errors.Is(err, repositories.ErrCVDocumentNotFound) {
			return nil, cvNotFoundError()
		}
		return nil, cvDatabaseError("cv.get", err)
	}
	return doc, nil
}

// Update applies a partial metadata update. Sections are left untouched.
func (s *CVService) Update(ctx context.Context, userID, id string, patch UpdateCVInput) (*models.CVDocument, error) {
	if err := validateCVUser(userID); err != nil {
		return nil, err
	}
	if err := validateCVID(id); err != nil {
		return nil, err
	}
	if patch.Empty() {
		return nil, cvValidationError("at least one field is required")
	}

	doc, err := s.store.GetByIDForUser(ctx, userID, id)
	if err != nil {
		if errors.Is(err, repositories.ErrCVDocumentNotFound) {
			return nil, cvNotFoundError()
		}
		return nil, cvDatabaseError("cv.update", err)
	}
	existingSections := doc.Sections

	if patch.Title != nil {
		title, err := validateCVTitle(*patch.Title, "")
		if err != nil {
			return nil, err
		}
		doc.Title = title
	}
	if patch.TargetRole != nil {
		targetRole, err := validateCVText("target_role", *patch.TargetRole, MaxCVTargetRoleLength)
		if err != nil {
			return nil, err
		}
		doc.TargetRole = targetRole
	}
	if patch.TemplateID != nil {
		templateID, err := validateCVText("template_id", *patch.TemplateID, MaxCVTemplateIDLength)
		if err != nil {
			return nil, err
		}
		doc.TemplateID = templateID
	}
	if patch.ThemeConfig != nil {
		if err := validateCVJSONObject("theme_config", *patch.ThemeConfig, MaxCVThemeConfigBytes); err != nil {
			return nil, err
		}
		doc.ThemeConfig = *patch.ThemeConfig
	}
	if patch.IsPrimary != nil {
		doc.IsPrimary = *patch.IsPrimary
	}

	updated, err := s.store.Update(ctx, doc)
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrCVDocumentNotFound):
			return nil, cvNotFoundError()
		case errors.Is(err, repositories.ErrInvalidCVInput):
			return nil, cvValidationError("invalid CV data")
		default:
			return nil, cvDatabaseError("cv.update", err)
		}
	}
	// Update deliberately leaves sections untouched; return the full document
	// so the response matches GET /api/cvs/{id}.
	updated.Sections = existingSections
	return updated, nil
}

// PutSections replaces the CV's sections in the supplied order. Array position
// becomes sort_order, guaranteeing the requested ordering is persisted.
func (s *CVService) PutSections(ctx context.Context, userID, id string, sections []CVSectionInput) (*models.CVDocument, error) {
	if err := validateCVUser(userID); err != nil {
		return nil, err
	}
	if err := validateCVID(id); err != nil {
		return nil, err
	}

	normalized, err := normalizeCVSections(sections, true, id)
	if err != nil {
		return nil, err
	}

	doc, err := s.store.GetByIDForUser(ctx, userID, id)
	if err != nil {
		if errors.Is(err, repositories.ErrCVDocumentNotFound) {
			return nil, cvNotFoundError()
		}
		return nil, cvDatabaseError("cv.sections", err)
	}

	doc.Sections = normalized
	updated, err := s.store.UpdateWithSections(ctx, doc)
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrCVDocumentNotFound):
			return nil, cvNotFoundError()
		case errors.Is(err, repositories.ErrCVSectionNotFound):
			return nil, cvValidationError("one or more sections do not belong to this CV")
		case errors.Is(err, repositories.ErrInvalidCVInput):
			return nil, cvValidationError("invalid CV section data")
		default:
			return nil, cvDatabaseError("cv.sections", err)
		}
	}
	return updated, nil
}

// Delete removes a CV (and, by cascade, its sections), scoped to the owner.
func (s *CVService) Delete(ctx context.Context, userID, id string) error {
	if err := validateCVUser(userID); err != nil {
		return err
	}
	if err := validateCVID(id); err != nil {
		return err
	}

	if err := s.store.DeleteForUser(ctx, userID, id); err != nil {
		if errors.Is(err, repositories.ErrCVDocumentNotFound) {
			return cvNotFoundError()
		}
		return cvDatabaseError("cv.delete", err)
	}
	return nil
}

func (s *CVService) buildCVDocument(ctx context.Context, userID string, input CreateCVInput) (*models.CVDocument, error) {
	resumeID := strings.TrimSpace(input.ResumeID)
	title, err := validateCVTitle(input.Title, resumeID)
	if err != nil {
		return nil, err
	}

	targetRole, err := validateCVText("target_role", input.TargetRole, MaxCVTargetRoleLength)
	if err != nil {
		return nil, err
	}
	templateID, err := validateCVText("template_id", input.TemplateID, MaxCVTemplateIDLength)
	if err != nil {
		return nil, err
	}
	if err := validateCVJSONObject("theme_config", input.ThemeConfig, MaxCVThemeConfigBytes); err != nil {
		return nil, err
	}

	sections := input.Sections
	if len(sections) == 0 && resumeID != "" {
		imported, err := s.importResumeSection(ctx, userID, resumeID)
		if err != nil {
			return nil, err
		}
		sections = []CVSectionInput{*imported}
	}

	normalized, err := normalizeCVSections(sections, false, "")
	if err != nil {
		return nil, err
	}

	return &models.CVDocument{
		UserID:      userID,
		Title:       title,
		TargetRole:  targetRole,
		TemplateID:  templateID,
		ThemeConfig: normalizeCVMap(input.ThemeConfig),
		IsPrimary:   input.IsPrimary,
		Sections:    normalized,
	}, nil
}

// importResumeSection seeds a single summary section from the resume's parsed
// text. Full structural parsing (splitting into experience/education/...) is a
// separate feature.
func (s *CVService) importResumeSection(ctx context.Context, userID, resumeID string) (*CVSectionInput, error) {
	if err := uuid.Validate(resumeID); err != nil {
		return nil, cvValidationError("resume_id must be a valid UUID")
	}
	if s.resumes == nil {
		return nil, cvValidationError("resume import is not available")
	}

	resume, err := s.resumes.GetByID(ctx, resumeID)
	if err != nil {
		if errors.Is(err, repositories.ErrResumeNotFound) {
			return nil, cvValidationError("resume not found")
		}
		return nil, cvDatabaseError("cv.import_resume", err)
	}
	if resume == nil || resume.UserID != userID {
		return nil, cvValidationError("resume not found")
	}
	if resume.ParsedText == nil || strings.TrimSpace(*resume.ParsedText) == "" {
		return nil, cvValidationError("resume has no parsed text to import")
	}

	text := strings.TrimSpace(*resume.ParsedText)
	if len(text) > MaxCVContentBytes {
		return nil, cvValidationError("resume text is too large to import")
	}

	return &CVSectionInput{
		SectionType: models.CVSectionSummary,
		Title:       "Imported Resume",
		Content:     map[string]any{"text": text},
	}, nil
}

// normalizeCVSections trims titles, validates types/content, and assigns
// sort_order from the array position.
func normalizeCVSections(sections []CVSectionInput, allowIDs bool, cvID string) ([]*models.CVSection, error) {
	if len(sections) > MaxCVSections {
		return nil, cvValidationError(fmt.Sprintf("at most %d sections are allowed", MaxCVSections))
	}

	out := make([]*models.CVSection, 0, len(sections))
	for i, section := range sections {
		if !section.SectionType.Valid() {
			return nil, cvValidationError(fmt.Sprintf("unsupported section_type %q", section.SectionType))
		}

		title := strings.TrimSpace(section.Title)
		if title == "" {
			return nil, cvValidationError("section title is required")
		}
		if utf8.RuneCountInString(title) > MaxCVSectionTitleLength {
			return nil, cvValidationError(
				fmt.Sprintf("section titles must be at most %d characters", MaxCVSectionTitleLength))
		}

		sectionID := strings.TrimSpace(section.ID)
		if sectionID != "" && !allowIDs {
			return nil, cvValidationError("section ids are not allowed when creating a CV")
		}
		if sectionID != "" {
			if err := uuid.Validate(sectionID); err != nil {
				return nil, cvValidationError("section id must be a valid UUID")
			}
		}

		if err := validateCVJSONObject("section content", section.Content, MaxCVContentBytes); err != nil {
			return nil, err
		}

		out = append(out, &models.CVSection{
			ID:          sectionID,
			CVID:        cvID,
			SectionType: section.SectionType,
			Title:       title,
			SortOrder:   i,
			Content:     normalizeCVMap(section.Content),
		})
	}
	return out, nil
}

func validateCVUser(userID string) error {
	if strings.TrimSpace(userID) == "" {
		return cvValidationError("user_id is required")
	}
	return nil
}

func validateCVID(id string) error {
	if strings.TrimSpace(id) == "" {
		return cvValidationError("cv id is required")
	}
	if err := uuid.Validate(strings.TrimSpace(id)); err != nil {
		return cvValidationError("cv id must be a valid UUID")
	}
	return nil
}

// validateCVTitle trims and length-checks a title. A blank title is allowed
// only when a resume import will supply the default.
func validateCVTitle(raw, resumeID string) (string, error) {
	title := strings.TrimSpace(raw)
	if title == "" {
		if resumeID != "" {
			return "Imported CV", nil
		}
		return "", cvValidationError("title is required")
	}
	if utf8.RuneCountInString(title) > MaxCVTitleLength {
		return "", cvValidationError(
			fmt.Sprintf("title must be at most %d characters", MaxCVTitleLength))
	}
	return title, nil
}

func validateCVText(field, raw string, maxLength int) (string, error) {
	value := strings.TrimSpace(raw)
	if utf8.RuneCountInString(value) > maxLength {
		return "", cvValidationError(
			fmt.Sprintf("%s must be at most %d characters", field, maxLength))
	}
	return value, nil
}

func validateCVJSONObject(field string, value map[string]any, maxBytes int) error {
	encoded, err := json.Marshal(normalizeCVMap(value))
	if err != nil {
		return cvValidationError(fmt.Sprintf("%s must be a valid JSON object", field))
	}
	if len(encoded) > maxBytes {
		return cvValidationError(fmt.Sprintf("%s is too large", field))
	}
	return nil
}

func normalizeCVMap(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return value
}

func cvValidationError(message string) error {
	return utils.NewValidationError(message, map[string]string{"operation": "cv.validate"})
}

func cvNotFoundError() error {
	return utils.NewNotFoundError("CV not found.")
}

func cvDatabaseError(operation string, err error) error {
	return utils.NewDatabaseError(err, map[string]string{"operation": operation})
}
