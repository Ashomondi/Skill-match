//go:build integration

package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"skill-match/backend/models"
)

// SM-01 — structured CV persistence. Runs against a real PostgreSQL.
//
//	TEST_DATABASE_URL=... go test -tags integration ./repositories -run TestCVDocumentRepository -v

func TestCVDocumentRepositoryCreateAndRetrieveSections(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	repo := NewCVDocumentRepository(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, &models.CVDocument{
		UserID:     user.ID,
		Title:      "Backend CV",
		TargetRole: "Go Engineer",
		TemplateID: "modern",
		ThemeConfig: map[string]any{
			"accent": "#123456",
			"font":   "Inter",
		},
		IsPrimary: true,
	})
	if err != nil {
		t.Fatalf("create cv: %v", err)
	}
	if created.ID == "" || !created.IsPrimary || created.TargetRole != "Go Engineer" {
		t.Fatalf("unexpected created cv: %+v", created)
	}

	created.Sections = []*models.CVSection{
		{
			SectionType: models.CVSectionSummary,
			Title:       "Summary",
			SortOrder:   0,
			Content:     map[string]any{"text": "Backend engineer focused on Go."},
		},
		{
			SectionType: models.CVSectionExperience,
			Title:       "Experience",
			SortOrder:   1,
			Content: map[string]any{
				"items": []any{
					map[string]any{"company": "Acme", "role": "Senior Engineer"},
				},
			},
		},
		{
			SectionType: models.CVSectionSkills,
			Title:       "Skills",
			SortOrder:   2,
			Content:     map[string]any{"skills": []any{"Go", "PostgreSQL"}},
		},
	}

	saved, err := repo.UpdateWithSections(ctx, created)
	if err != nil {
		t.Fatalf("save sections: %v", err)
	}
	if len(saved.Sections) != 3 {
		t.Fatalf("expected 3 saved sections, got %d", len(saved.Sections))
	}

	got, err := repo.GetByIDForUser(ctx, user.ID, created.ID)
	if err != nil {
		t.Fatalf("get cv: %v", err)
	}
	if got.Title != "Backend CV" || got.TemplateID != "modern" {
		t.Fatalf("document metadata did not round-trip: %+v", got)
	}
	if got.ThemeConfig["accent"] != "#123456" || got.ThemeConfig["font"] != "Inter" {
		t.Fatalf("theme_config did not round-trip: %+v", got.ThemeConfig)
	}
	if len(got.Sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(got.Sections))
	}

	wantTypes := []models.CVSectionType{
		models.CVSectionSummary,
		models.CVSectionExperience,
		models.CVSectionSkills,
	}
	for i, want := range wantTypes {
		if got.Sections[i].SectionType != want {
			t.Errorf("section %d: expected type %q, got %q", i, want, got.Sections[i].SectionType)
		}
		if got.Sections[i].SortOrder != i {
			t.Errorf("section %d: expected sort_order %d, got %d", i, i, got.Sections[i].SortOrder)
		}
		if got.Sections[i].CVID != created.ID {
			t.Errorf("section %d: expected cv_id %q, got %q", i, created.ID, got.Sections[i].CVID)
		}
	}
	if got.Sections[0].Content["text"] != "Backend engineer focused on Go." {
		t.Errorf("summary content did not round-trip: %+v", got.Sections[0].Content)
	}
	skills, ok := got.Sections[2].Content["skills"].([]any)
	if !ok || len(skills) != 2 || skills[0] != "Go" {
		t.Errorf("nested skills JSON did not round-trip: %+v", got.Sections[2].Content)
	}

	sections, err := repo.ListSections(ctx, user.ID, created.ID)
	if err != nil {
		t.Fatalf("list sections: %v", err)
	}
	if len(sections) != 3 {
		t.Fatalf("expected 3 listed sections, got %d", len(sections))
	}
}

func TestCVDocumentRepositoryUpdateTransaction(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	repo := NewCVDocumentRepository(pool)
	ctx := context.Background()

	doc, err := repo.Create(ctx, &models.CVDocument{
		UserID: user.ID,
		Title:  "Original",
	})
	if err != nil {
		t.Fatalf("create cv: %v", err)
	}

	doc.Sections = []*models.CVSection{{
		SectionType: models.CVSectionSummary,
		Title:       "Summary",
		SortOrder:   0,
		Content:     map[string]any{"text": "original"},
	}}
	doc, err = repo.UpdateWithSections(ctx, doc)
	if err != nil {
		t.Fatalf("create section: %v", err)
	}
	sectionID := doc.Sections[0].ID

	// Metadata changes first, then a nonexistent section ID fails. The whole
	// update must roll back, including the title change.
	doc.Title = "Should Roll Back"
	doc.Sections = []*models.CVSection{
		{
			ID:          sectionID,
			SectionType: models.CVSectionSummary,
			Title:       "Summary",
			SortOrder:   0,
			Content:     map[string]any{"text": "changed"},
		},
		{
			ID:          "00000000-0000-0000-0000-000000000000",
			SectionType: models.CVSectionSkills,
			Title:       "Missing",
		},
	}
	if _, err := repo.UpdateWithSections(ctx, doc); !errors.Is(err, ErrCVSectionNotFound) {
		t.Fatalf("expected ErrCVSectionNotFound, got %v", err)
	}

	got, err := repo.getByID(ctx, doc.ID)
	if err != nil {
		t.Fatalf("get after rollback: %v", err)
	}
	if got.Title != "Original" {
		t.Fatalf("expected title rollback to Original, got %q", got.Title)
	}
	if len(got.Sections) != 1 || got.Sections[0].Content["text"] != "original" {
		t.Fatalf("expected sections to roll back untouched, got %+v", got.Sections)
	}

	// A valid update reconciles sections atomically: existing updated in place,
	// new inserted, omitted deleted.
	doc.Title = "Updated"
	doc.Sections = []*models.CVSection{
		{
			ID:          sectionID,
			SectionType: models.CVSectionSummary,
			Title:       "Profile",
			SortOrder:   0,
			Content:     map[string]any{"text": "new summary"},
		},
		{
			SectionType: models.CVSectionEducation,
			Title:       "Education",
			SortOrder:   1,
			Content:     map[string]any{"school": "State University"},
		},
	}
	updated, err := repo.UpdateWithSections(ctx, doc)
	if err != nil {
		t.Fatalf("valid update: %v", err)
	}
	if updated.Title != "Updated" || len(updated.Sections) != 2 {
		t.Fatalf("unexpected updated cv: %+v", updated)
	}
	if updated.Sections[0].ID != sectionID {
		t.Fatalf("expected existing section to be updated in place, got %q", updated.Sections[0].ID)
	}

	got, err = repo.getByID(ctx, doc.ID)
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if got.Title != "Updated" || len(got.Sections) != 2 {
		t.Fatalf("unexpected persisted cv: %+v", got)
	}
	if got.Sections[0].Title != "Profile" {
		t.Fatalf("expected section title update, got %q", got.Sections[0].Title)
	}
}

func TestCVDocumentRepositoryPrimaryExclusivity(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	repo := NewCVDocumentRepository(pool)
	ctx := context.Background()

	first, err := repo.Create(ctx, &models.CVDocument{
		UserID:    user.ID,
		Title:     "First",
		IsPrimary: true,
	})
	if err != nil {
		t.Fatalf("create first cv: %v", err)
	}
	second, err := repo.Create(ctx, &models.CVDocument{
		UserID:    user.ID,
		Title:     "Second",
		IsPrimary: true,
	})
	if err != nil {
		t.Fatalf("create second cv: %v", err)
	}

	gotFirst, err := repo.getByID(ctx, first.ID)
	if err != nil {
		t.Fatalf("get first: %v", err)
	}
	if gotFirst.IsPrimary {
		t.Fatal("expected creating a new primary to demote the previous primary")
	}

	list, err := repo.ListByUserID(ctx, user.ID, 0)
	if err != nil {
		t.Fatalf("list cvs: %v", err)
	}
	if len(list) != 2 || !list[0].IsPrimary || list[0].ID != second.ID {
		t.Fatalf("expected primary cv first, got %+v", list)
	}

	// Promoting the first through Update must demote the second.
	gotFirst.IsPrimary = true
	if _, err := repo.Update(ctx, gotFirst); err != nil {
		t.Fatalf("promote first: %v", err)
	}
	gotSecond, err := repo.getByID(ctx, second.ID)
	if err != nil {
		t.Fatalf("get second: %v", err)
	}
	if gotSecond.IsPrimary {
		t.Fatal("expected update to demote the other primary cv")
	}
}

func TestCVDocumentRepositorySectionCRUD(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	repo := NewCVDocumentRepository(pool)
	ctx := context.Background()

	doc, err := repo.Create(ctx, &models.CVDocument{UserID: user.ID, Title: "CRUD CV"})
	if err != nil {
		t.Fatalf("create cv: %v", err)
	}

	section, err := repo.CreateSection(ctx, user.ID, &models.CVSection{
		CVID:        doc.ID,
		SectionType: models.CVSectionProjects,
		Title:       "Projects",
		SortOrder:   2,
		Content:     map[string]any{"name": "Skill Match"},
	})
	if err != nil {
		t.Fatalf("create section: %v", err)
	}

	section.Title = "Selected Projects"
	section.SortOrder = 0
	updated, err := repo.UpdateSection(ctx, user.ID, section)
	if err != nil {
		t.Fatalf("update section: %v", err)
	}
	if updated.Title != "Selected Projects" || updated.SortOrder != 0 {
		t.Fatalf("unexpected updated section: %+v", updated)
	}

	second, err := repo.CreateSection(ctx, user.ID, &models.CVSection{
		CVID:        doc.ID,
		SectionType: models.CVSectionCustom,
		Title:       "Awards",
		SortOrder:   1,
	})
	if err != nil {
		t.Fatalf("create second section: %v", err)
	}
	if err := repo.ReorderSections(ctx, user.ID, doc.ID, []string{second.ID, section.ID}); err != nil {
		t.Fatalf("reorder sections: %v", err)
	}

	listed, err := repo.ListSections(ctx, user.ID, doc.ID)
	if err != nil {
		t.Fatalf("list sections: %v", err)
	}
	if len(listed) != 2 || listed[0].ID != second.ID || listed[1].ID != section.ID {
		t.Fatalf("expected reordered sections, got %+v", listed)
	}

	if err := repo.DeleteSection(ctx, user.ID, doc.ID, section.ID); err != nil {
		t.Fatalf("delete section: %v", err)
	}
	if _, err := repo.ListSections(ctx, user.ID, doc.ID); err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if err := repo.DeleteSection(ctx, user.ID, doc.ID, section.ID); !errors.Is(err, ErrCVSectionNotFound) {
		t.Fatalf("expected ErrCVSectionNotFound, got %v", err)
	}
}

func TestCVDocumentRepositoryCascadesOnDelete(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	repo := NewCVDocumentRepository(pool)
	ctx := context.Background()

	doc, err := repo.Create(ctx, &models.CVDocument{
		UserID: user.ID,
		Title:  "Cascade CV",
		Sections: []*models.CVSection{{
			SectionType: models.CVSectionSummary,
			Title:       "Summary",
			Content:     map[string]any{"text": "cascade me"},
		}},
	})
	if err != nil {
		t.Fatalf("create cv: %v", err)
	}
	if len(doc.Sections) != 1 {
		t.Fatalf("expected one section, got %d", len(doc.Sections))
	}

	// Deleting the CV must cascade to its sections.
	if err := repo.DeleteForUser(ctx, user.ID, doc.ID); err != nil {
		t.Fatalf("delete cv: %v", err)
	}
	if n := countRows(t, pool, `SELECT count(*) FROM cv_sections WHERE cv_id = $1`, doc.ID); n != 0 {
		t.Fatalf("expected sections to cascade on cv delete, got %d", n)
	}

	// Deleting the user must cascade to documents and sections.
	doc, err = repo.Create(ctx, &models.CVDocument{
		UserID: user.ID,
		Title:  "User Cascade CV",
		Sections: []*models.CVSection{{
			SectionType: models.CVSectionSkills,
			Title:       "Skills",
		}},
	})
	if err != nil {
		t.Fatalf("create second cv: %v", err)
	}
	if err := NewUserRepository(pool).Delete(ctx, user.ID); err != nil {
		t.Fatalf("delete user: %v", err)
	}
	if n := countRows(t, pool, `SELECT count(*) FROM cv_documents WHERE user_id = $1`, user.ID); n != 0 {
		t.Fatalf("expected cv documents to cascade on user delete, got %d", n)
	}
	if n := countRows(t, pool, `SELECT count(*) FROM cv_sections WHERE cv_id = $1`, doc.ID); n != 0 {
		t.Fatalf("expected sections to cascade on user delete, got %d", n)
	}
	if _, err := repo.getByID(ctx, doc.ID); !errors.Is(err, ErrCVDocumentNotFound) {
		t.Fatalf("expected ErrCVDocumentNotFound after user delete, got %v", err)
	}
}

func TestCVDocumentRepositoryUserIsolation(t *testing.T) {
	pool := connectTestPool(t)
	userA := createTestUser(t, pool)
	userB := createTestUser(t, pool)
	repo := NewCVDocumentRepository(pool)
	ctx := context.Background()

	doc, err := repo.Create(ctx, &models.CVDocument{
		UserID: userA.ID,
		Title:  "Private CV",
		Sections: []*models.CVSection{{
			SectionType: models.CVSectionSummary,
			Title:       "Summary",
		}},
	})
	if err != nil {
		t.Fatalf("create cv: %v", err)
	}

	if _, err := repo.GetByIDForUser(ctx, userB.ID, doc.ID); !errors.Is(err, ErrCVDocumentNotFound) {
		t.Fatalf("expected cross-user get to be hidden, got %v", err)
	}
	if err := repo.DeleteForUser(ctx, userB.ID, doc.ID); !errors.Is(err, ErrCVDocumentNotFound) {
		t.Fatalf("expected cross-user delete to be hidden, got %v", err)
	}
	if _, err := repo.CreateSection(ctx, userB.ID, &models.CVSection{
		CVID:        doc.ID,
		SectionType: models.CVSectionSkills,
		Title:       "Injected",
	}); !errors.Is(err, ErrCVDocumentNotFound) {
		t.Fatalf("expected cross-user section create to be rejected, got %v", err)
	}

	listB, err := repo.ListByUserID(ctx, userB.ID, 0)
	if err != nil {
		t.Fatalf("list B: %v", err)
	}
	if len(listB) != 0 {
		t.Fatalf("expected no cvs for user B, got %d", len(listB))
	}

	if _, err := repo.GetByIDForUser(ctx, userA.ID, doc.ID); err != nil {
		t.Fatalf("owner should still read the cv: %v", err)
	}
}

func TestCVDocumentRepositorySectionMutationsTouchDocument(t *testing.T) {
	pool := connectTestPool(t)
	user := createTestUser(t, pool)
	repo := NewCVDocumentRepository(pool)
	ctx := context.Background()

	doc, err := repo.Create(ctx, &models.CVDocument{UserID: user.ID, Title: "Touch CV"})
	if err != nil {
		t.Fatalf("create cv: %v", err)
	}
	last := doc.UpdatedAt

	assertAdvanced := func(label string) {
		t.Helper()
		got, err := repo.getByID(ctx, doc.ID)
		if err != nil {
			t.Fatalf("%s: get cv: %v", label, err)
		}
		if !got.UpdatedAt.After(last) {
			t.Fatalf("%s: expected document updated_at to advance; previous=%s current=%s", label, last, got.UpdatedAt)
		}
		last = got.UpdatedAt
	}

	time.Sleep(10 * time.Millisecond)
	section, err := repo.CreateSection(ctx, user.ID, &models.CVSection{
		CVID:        doc.ID,
		SectionType: models.CVSectionSummary,
		Title:       "Summary",
	})
	if err != nil {
		t.Fatalf("create section: %v", err)
	}
	assertAdvanced("create section")

	time.Sleep(10 * time.Millisecond)
	section.Title = "Profile"
	if _, err := repo.UpdateSection(ctx, user.ID, section); err != nil {
		t.Fatalf("update section: %v", err)
	}
	assertAdvanced("update section")

	time.Sleep(10 * time.Millisecond)
	second, err := repo.CreateSection(ctx, user.ID, &models.CVSection{
		CVID:        doc.ID,
		SectionType: models.CVSectionSkills,
		Title:       "Skills",
	})
	if err != nil {
		t.Fatalf("create second section: %v", err)
	}
	if err := repo.ReorderSections(ctx, user.ID, doc.ID, []string{second.ID, section.ID}); err != nil {
		t.Fatalf("reorder sections: %v", err)
	}
	assertAdvanced("reorder sections")

	time.Sleep(10 * time.Millisecond)
	if err := repo.DeleteSection(ctx, user.ID, doc.ID, section.ID); err != nil {
		t.Fatalf("delete section: %v", err)
	}
	assertAdvanced("delete section")
}

func countRows(t *testing.T, pool *pgxpool.Pool, query string, args ...any) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(), query, args...).Scan(&n); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	return n
}
