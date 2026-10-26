package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"skill-match/backend/models"
)

// Sentinel errors for structured CV operations.
var (
	ErrCVDocumentNotFound = errors.New("repositories: cv document not found")
	ErrCVSectionNotFound  = errors.New("repositories: cv section not found")
	ErrInvalidCVInput     = errors.New("repositories: invalid cv input")
)

// CVDocumentRepository provides persistence for structured CV documents and
// their ordered sections. Writes that touch more than one row run inside a
// transaction so a document and its sections never diverge.
type CVDocumentRepository struct {
	db *pgxpool.Pool
}

// NewCVDocumentRepository constructs a CVDocumentRepository.
func NewCVDocumentRepository(db *pgxpool.Pool) *CVDocumentRepository {
	return &CVDocumentRepository{db: db}
}

const cvDocumentColumns = `id, user_id, title, target_role, template_id,
	theme_config, is_primary, created_at, updated_at`

const cvSectionColumns = `id, cv_id, section_type, title, sort_order,
	content, created_at, updated_at`

// cvQuerier is the subset of pgxpool.Pool and pgx.Tx used by read helpers so
// the same scan logic can run inside or outside a transaction.
type cvQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Create inserts a CV document. When IsPrimary is true any previous primary
// CV for the user is demoted in the same transaction. Sections supplied on
// doc are inserted in order.
func (r *CVDocumentRepository) Create(ctx context.Context, doc *models.CVDocument) (*models.CVDocument, error) {
	if err := validateCVDocument(doc, false); err != nil {
		return nil, err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("repositories: begin cv create: %w", err)
	}
	defer tx.Rollback(ctx)

	if doc.IsPrimary {
		if err := clearPrimaryCV(ctx, tx, doc.UserID, ""); err != nil {
			return nil, err
		}
	}

	q := fmt.Sprintf(`
		INSERT INTO cv_documents (user_id, title, target_role, template_id, theme_config, is_primary)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING %s`, cvDocumentColumns)

	out, err := scanCVDocument(tx.QueryRow(ctx, q,
		doc.UserID, strings.TrimSpace(doc.Title), doc.TargetRole,
		doc.TemplateID, cvJSON(doc.ThemeConfig), doc.IsPrimary))
	if err != nil {
		return nil, err
	}

	if len(doc.Sections) > 0 {
		out.Sections, err = syncCVSections(ctx, tx, out.ID, doc.Sections)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("repositories: commit cv create: %w", err)
	}
	return out, nil
}

// GetByID fetches a CV document and its sections in sort order.
func (r *CVDocumentRepository) GetByID(ctx context.Context, id string) (*models.CVDocument, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidCVInput)
	}
	return getCVDocument(ctx, r.db, `WHERE id = $1`, id)
}

// GetByIDForUser fetches a CV document only when it belongs to userID.
func (r *CVDocumentRepository) GetByIDForUser(ctx context.Context, userID, id string) (*models.CVDocument, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("%w: user_id and id are required", ErrInvalidCVInput)
	}
	return getCVDocument(ctx, r.db, `WHERE id = $1 AND user_id = $2`, id, userID)
}

// ListByUserID returns a user's CV documents, primary first then most recently
// updated. limit <= 0 defaults to 50.
func (r *CVDocumentRepository) ListByUserID(ctx context.Context, userID string, limit int) ([]*models.CVDocument, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidCVInput)
	}
	if limit <= 0 {
		limit = 50
	}

	q := fmt.Sprintf(`
		SELECT %s FROM cv_documents
		WHERE user_id = $1
		ORDER BY is_primary DESC, updated_at DESC, id
		LIMIT $2`, cvDocumentColumns)

	rows, err := r.db.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("repositories: list cv documents: %w", err)
	}
	defer rows.Close()

	docs := make([]*models.CVDocument, 0)
	ids := make([]string, 0)
	for rows.Next() {
		doc, err := scanCVDocument(rows)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
		ids = append(ids, doc.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: list cv documents: %w", err)
	}

	sections, err := loadCVSections(ctx, r.db, ids)
	if err != nil {
		return nil, err
	}
	for _, doc := range docs {
		doc.Sections = sections[doc.ID]
		if doc.Sections == nil {
			doc.Sections = []*models.CVSection{}
		}
	}
	return docs, nil
}

// Update updates document metadata. Sections are left untouched; use
// UpdateWithSections to replace them atomically. doc.UserID scopes the write
// and doc.ID identifies the row.
func (r *CVDocumentRepository) Update(ctx context.Context, doc *models.CVDocument) (*models.CVDocument, error) {
	if err := validateCVDocument(doc, true); err != nil {
		return nil, err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("repositories: begin cv update: %w", err)
	}
	defer tx.Rollback(ctx)

	out, err := updateCVDocumentTx(ctx, tx, doc)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("repositories: commit cv update: %w", err)
	}
	return out, nil
}

// UpdateWithSections atomically updates document metadata and synchronizes its
// sections. Sections with an ID are updated in place, sections without an ID
// are inserted, and sections missing from doc.Sections are deleted. Passing a
// nil or empty slice removes every section.
func (r *CVDocumentRepository) UpdateWithSections(ctx context.Context, doc *models.CVDocument) (*models.CVDocument, error) {
	if err := validateCVDocument(doc, true); err != nil {
		return nil, err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("repositories: begin cv update with sections: %w", err)
	}
	defer tx.Rollback(ctx)

	out, err := updateCVDocumentTx(ctx, tx, doc)
	if err != nil {
		return nil, err
	}

	out.Sections, err = syncCVSections(ctx, tx, out.ID, doc.Sections)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("repositories: commit cv update with sections: %w", err)
	}
	return out, nil
}

// Delete removes a CV document by ID. Sections are removed by ON DELETE
// CASCADE.
func (r *CVDocumentRepository) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidCVInput)
	}
	return deleteCVDocument(ctx, r.db, `DELETE FROM cv_documents WHERE id = $1`, id)
}

// DeleteForUser removes a CV document only when it belongs to userID.
func (r *CVDocumentRepository) DeleteForUser(ctx context.Context, userID, id string) error {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: user_id and id are required", ErrInvalidCVInput)
	}
	return deleteCVDocument(ctx, r.db,
		`DELETE FROM cv_documents WHERE id = $1 AND user_id = $2`, id, userID)
}

// CreateSection appends a section to a CV owned by userID.
func (r *CVDocumentRepository) CreateSection(ctx context.Context, userID string, section *models.CVSection) (*models.CVSection, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidCVInput)
	}
	if err := validateCVSection(section, true); err != nil {
		return nil, err
	}

	q := fmt.Sprintf(`
		INSERT INTO cv_sections (cv_id, section_type, title, sort_order, content)
		SELECT d.id, $2, $3, $4, $5
		FROM cv_documents d
		WHERE d.id = $1 AND d.user_id = $6
		RETURNING %s`, cvSectionColumns)

	out, err := scanCVSection(r.db.QueryRow(ctx, q,
		section.CVID, section.SectionType, strings.TrimSpace(section.Title),
		section.SortOrder, cvJSON(section.Content), userID))
	if err != nil {
		// No returning row means the CV does not exist or belongs to someone
		// else; report that instead of leaking a section-scoped error.
		if errors.Is(err, ErrCVSectionNotFound) {
			return nil, ErrCVDocumentNotFound
		}
		return nil, err
	}
	return out, nil
}

// UpdateSection updates a section in place, scoped to the owning user.
func (r *CVDocumentRepository) UpdateSection(ctx context.Context, userID string, section *models.CVSection) (*models.CVSection, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidCVInput)
	}
	if err := validateCVSection(section, true); err != nil {
		return nil, err
	}
	if strings.TrimSpace(section.ID) == "" {
		return nil, fmt.Errorf("%w: section id is required", ErrInvalidCVInput)
	}

	q := fmt.Sprintf(`
		UPDATE cv_sections
		SET section_type = $1, title = $2, sort_order = $3, content = $4, updated_at = now()
		WHERE id = $5 AND cv_id = $6
		  AND cv_id IN (SELECT id FROM cv_documents WHERE id = $6 AND user_id = $7)
		RETURNING %s`, cvSectionColumns)

	return scanCVSection(r.db.QueryRow(ctx, q,
		section.SectionType, strings.TrimSpace(section.Title),
		section.SortOrder, cvJSON(section.Content), section.ID,
		section.CVID, userID))
}

// ListSections returns a CV's sections ordered for rendering.
func (r *CVDocumentRepository) ListSections(ctx context.Context, userID, cvID string) ([]*models.CVSection, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(cvID) == "" {
		return nil, fmt.Errorf("%w: user_id and cv_id are required", ErrInvalidCVInput)
	}

	const q = `
		SELECT s.id, s.cv_id, s.section_type, s.title, s.sort_order,
		       s.content, s.created_at, s.updated_at
		FROM cv_sections s
		JOIN cv_documents d ON d.id = s.cv_id
		WHERE s.cv_id = $1 AND d.user_id = $2
		ORDER BY s.sort_order ASC, s.created_at ASC, s.id`

	rows, err := r.db.Query(ctx, q, cvID, userID)
	if err != nil {
		return nil, fmt.Errorf("repositories: list cv sections: %w", err)
	}
	defer rows.Close()

	out := make([]*models.CVSection, 0)
	for rows.Next() {
		section, err := scanCVSection(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, section)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: list cv sections: %w", err)
	}
	return out, nil
}

// DeleteSection removes a section, scoped to the owning user.
func (r *CVDocumentRepository) DeleteSection(ctx context.Context, userID, cvID, sectionID string) error {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(cvID) == "" || strings.TrimSpace(sectionID) == "" {
		return fmt.Errorf("%w: user_id, cv_id, and section id are required", ErrInvalidCVInput)
	}

	tag, err := r.db.Exec(ctx, `
		DELETE FROM cv_sections
		WHERE id = $1 AND cv_id = $2
		  AND cv_id IN (SELECT id FROM cv_documents WHERE id = $2 AND user_id = $3)`,
		sectionID, cvID, userID)
	if err != nil {
		return fmt.Errorf("repositories: delete cv section: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrCVSectionNotFound
	}
	return nil
}

// ReorderSections assigns sort_order from the position of each ID in
// orderedIDs. The whole reorder is one transaction so a partial order is never
// visible.
func (r *CVDocumentRepository) ReorderSections(ctx context.Context, userID, cvID string, orderedIDs []string) error {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(cvID) == "" {
		return fmt.Errorf("%w: user_id and cv_id are required", ErrInvalidCVInput)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repositories: begin cv reorder: %w", err)
	}
	defer tx.Rollback(ctx)

	var owned bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM cv_documents WHERE id = $1 AND user_id = $2)`,
		cvID, userID).Scan(&owned); err != nil {
		return fmt.Errorf("repositories: check cv ownership: %w", err)
	}
	if !owned {
		return ErrCVDocumentNotFound
	}

	for i, id := range orderedIDs {
		tag, err := tx.Exec(ctx, `
			UPDATE cv_sections
			SET sort_order = $1, updated_at = now()
			WHERE id = $2 AND cv_id = $3`, i, id, cvID)
		if err != nil {
			return fmt.Errorf("repositories: reorder cv section: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrCVSectionNotFound
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("repositories: commit cv reorder: %w", err)
	}
	return nil
}

func getCVDocument(ctx context.Context, db cvQuerier, where string, args ...any) (*models.CVDocument, error) {
	q := fmt.Sprintf(`SELECT %s FROM cv_documents %s`, cvDocumentColumns, where)
	doc, err := scanCVDocument(db.QueryRow(ctx, q, args...))
	if err != nil {
		return nil, err
	}

	sections, err := loadCVSections(ctx, db, []string{doc.ID})
	if err != nil {
		return nil, err
	}
	doc.Sections = sections[doc.ID]
	if doc.Sections == nil {
		doc.Sections = []*models.CVSection{}
	}
	return doc, nil
}

func updateCVDocumentTx(ctx context.Context, tx pgx.Tx, doc *models.CVDocument) (*models.CVDocument, error) {
	if doc.IsPrimary {
		if err := clearPrimaryCV(ctx, tx, doc.UserID, doc.ID); err != nil {
			return nil, err
		}
	}

	q := fmt.Sprintf(`
		UPDATE cv_documents
		SET title = $1, target_role = $2, template_id = $3,
		    theme_config = $4, is_primary = $5, updated_at = now()
		WHERE id = $6 AND user_id = $7
		RETURNING %s`, cvDocumentColumns)

	return scanCVDocument(tx.QueryRow(ctx, q,
		strings.TrimSpace(doc.Title), doc.TargetRole, doc.TemplateID,
		cvJSON(doc.ThemeConfig), doc.IsPrimary, doc.ID, doc.UserID))
}

func clearPrimaryCV(ctx context.Context, tx pgx.Tx, userID, exceptID string) error {
	const q = `
		UPDATE cv_documents
		SET is_primary = false, updated_at = now()
		WHERE user_id = $1 AND is_primary`

	var err error
	if exceptID == "" {
		_, err = tx.Exec(ctx, q, userID)
	} else {
		_, err = tx.Exec(ctx, q+` AND id <> $2`, userID, exceptID)
	}
	if err != nil {
		return fmt.Errorf("repositories: clear primary cv: %w", err)
	}
	return nil
}

// syncCVSections reconciles a CV's sections inside tx. Existing section IDs
// are updated in place; unknown IDs roll the whole transaction back.
func syncCVSections(ctx context.Context, tx pgx.Tx, cvID string, sections []*models.CVSection) ([]*models.CVSection, error) {
	out := make([]*models.CVSection, 0, len(sections))
	keptIDs := make([]string, 0, len(sections))
	seen := make(map[string]bool, len(sections))

	for _, section := range sections {
		if err := validateCVSection(section, false); err != nil {
			return nil, err
		}
		if section.CVID != "" && section.CVID != cvID {
			return nil, fmt.Errorf("%w: section %q belongs to another cv", ErrInvalidCVInput, section.ID)
		}
		if section.ID != "" {
			if seen[section.ID] {
				return nil, fmt.Errorf("%w: duplicate section id %q", ErrInvalidCVInput, section.ID)
			}
			seen[section.ID] = true
		}

		if section.ID == "" {
			q := fmt.Sprintf(`
				INSERT INTO cv_sections (cv_id, section_type, title, sort_order, content)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING %s`, cvSectionColumns)
			saved, err := scanCVSection(tx.QueryRow(ctx, q, cvID,
				section.SectionType, strings.TrimSpace(section.Title),
				section.SortOrder, cvJSON(section.Content)))
			if err != nil {
				return nil, err
			}
			out = append(out, saved)
			keptIDs = append(keptIDs, saved.ID)
			continue
		}

		q := fmt.Sprintf(`
			UPDATE cv_sections
			SET section_type = $1, title = $2, sort_order = $3, content = $4, updated_at = now()
			WHERE id = $5 AND cv_id = $6
			RETURNING %s`, cvSectionColumns)
		saved, err := scanCVSection(tx.QueryRow(ctx, q,
			section.SectionType, strings.TrimSpace(section.Title),
			section.SortOrder, cvJSON(section.Content), section.ID, cvID))
		if err != nil {
			return nil, err
		}
		out = append(out, saved)
		keptIDs = append(keptIDs, saved.ID)
	}

	if len(keptIDs) == 0 {
		if _, err := tx.Exec(ctx, `DELETE FROM cv_sections WHERE cv_id = $1`, cvID); err != nil {
			return nil, fmt.Errorf("repositories: clear cv sections: %w", err)
		}
		return out, nil
	}

	if _, err := tx.Exec(ctx, `
		DELETE FROM cv_sections
		WHERE cv_id = $1 AND NOT (id = ANY($2::uuid[]))`, cvID, keptIDs); err != nil {
		return nil, fmt.Errorf("repositories: prune cv sections: %w", err)
	}
	return out, nil
}

func loadCVSections(ctx context.Context, db cvQuerier, cvIDs []string) (map[string][]*models.CVSection, error) {
	out := make(map[string][]*models.CVSection, len(cvIDs))
	if len(cvIDs) == 0 {
		return out, nil
	}

	q := fmt.Sprintf(`
		SELECT %s FROM cv_sections
		WHERE cv_id = ANY($1::uuid[])
		ORDER BY cv_id, sort_order ASC, created_at ASC`, cvSectionColumns)

	rows, err := db.Query(ctx, q, cvIDs)
	if err != nil {
		return nil, fmt.Errorf("repositories: load cv sections: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		section, err := scanCVSection(rows)
		if err != nil {
			return nil, err
		}
		out[section.CVID] = append(out[section.CVID], section)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: load cv sections: %w", err)
	}
	return out, nil
}

func deleteCVDocument(ctx context.Context, db *pgxpool.Pool, q string, args ...any) error {
	tag, err := db.Exec(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("repositories: delete cv document: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrCVDocumentNotFound
	}
	return nil
}

func validateCVDocument(doc *models.CVDocument, requireID bool) error {
	if doc == nil {
		return fmt.Errorf("%w: cv document is required", ErrInvalidCVInput)
	}
	if strings.TrimSpace(doc.UserID) == "" {
		return fmt.Errorf("%w: user_id is required", ErrInvalidCVInput)
	}
	if requireID && strings.TrimSpace(doc.ID) == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidCVInput)
	}
	if strings.TrimSpace(doc.Title) == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidCVInput)
	}
	return nil
}

func validateCVSection(section *models.CVSection, requireCVID bool) error {
	if section == nil {
		return fmt.Errorf("%w: cv section is required", ErrInvalidCVInput)
	}
	if requireCVID && strings.TrimSpace(section.CVID) == "" {
		return fmt.Errorf("%w: cv_id is required", ErrInvalidCVInput)
	}
	if !section.SectionType.Valid() {
		return fmt.Errorf("%w: unsupported section_type %q", ErrInvalidCVInput, section.SectionType)
	}
	if strings.TrimSpace(section.Title) == "" {
		return fmt.Errorf("%w: section title is required", ErrInvalidCVInput)
	}
	return nil
}

// cvJSON normalizes a nil map to an empty JSON object so JSONB columns are
// always an object rather than a JSON null.
func cvJSON(m map[string]any) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	return m
}

func scanCVDocument(row pgx.Row) (*models.CVDocument, error) {
	doc := &models.CVDocument{}
	err := row.Scan(&doc.ID, &doc.UserID, &doc.Title, &doc.TargetRole,
		&doc.TemplateID, &doc.ThemeConfig, &doc.IsPrimary, &doc.CreatedAt,
		&doc.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCVDocumentNotFound
		}
		return nil, fmt.Errorf("repositories: query cv document: %w", err)
	}
	doc.ThemeConfig = cvJSON(doc.ThemeConfig)
	return doc, nil
}

func scanCVSection(row pgx.Row) (*models.CVSection, error) {
	section := &models.CVSection{}
	var sectionType string
	err := row.Scan(&section.ID, &section.CVID, &sectionType, &section.Title,
		&section.SortOrder, &section.Content, &section.CreatedAt,
		&section.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCVSectionNotFound
		}
		return nil, fmt.Errorf("repositories: query cv section: %w", err)
	}
	section.SectionType = models.CVSectionType(sectionType)
	section.Content = cvJSON(section.Content)
	return section, nil
}
