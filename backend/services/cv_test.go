package services

import (
	"context"
	"fmt"
	"testing"

	"skill-match/backend/models"
	"skill-match/backend/repositories"
	"skill-match/backend/utils"
)

const (
	testCVUser      = "11111111-1111-4111-8111-111111111111"
	testCVOtherUser = "22222222-2222-4222-8222-222222222222"
	testCVID        = "33333333-3333-4333-8333-333333333333"
	testCVResumeID  = "99999999-9999-4999-8999-999999999999"
)

type fakeCVStore struct {
	docs        map[string]*models.CVDocument
	order       []string
	createErr   error
	getErr      error
	listErr     error
	updateErr   error
	sectionsErr error
	deleteErr   error
	lastCreated *models.CVDocument
	lastUpdated *models.CVDocument
}

func newFakeCVStore(docs ...*models.CVDocument) *fakeCVStore {
	store := &fakeCVStore{docs: map[string]*models.CVDocument{}}
	for _, doc := range docs {
		store.docs[doc.ID] = doc
		store.order = append(store.order, doc.ID)
	}
	return store
}

func (f *fakeCVStore) Create(_ context.Context, doc *models.CVDocument) (*models.CVDocument, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	out := *doc
	if out.ID == "" {
		out.ID = testCVID
	}
	assignFakeSectionIDs(out.Sections, out.ID)
	f.docs[out.ID] = &out
	f.order = append(f.order, out.ID)
	f.lastCreated = &out
	return &out, nil
}

func (f *fakeCVStore) GetByIDForUser(_ context.Context, userID, id string) (*models.CVDocument, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	doc, ok := f.docs[id]
	if !ok || doc.UserID != userID {
		return nil, repositories.ErrCVDocumentNotFound
	}
	return doc, nil
}

func (f *fakeCVStore) ListByUserID(_ context.Context, userID string, _ int) ([]*models.CVDocument, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	out := make([]*models.CVDocument, 0)
	for _, id := range f.order {
		if doc := f.docs[id]; doc.UserID == userID {
			out = append(out, doc)
		}
	}
	return out, nil
}

func (f *fakeCVStore) Update(_ context.Context, doc *models.CVDocument) (*models.CVDocument, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	existing, ok := f.docs[doc.ID]
	if !ok || existing.UserID != doc.UserID {
		return nil, repositories.ErrCVDocumentNotFound
	}
	out := *doc
	f.docs[doc.ID] = &out
	f.lastUpdated = &out
	return &out, nil
}

func (f *fakeCVStore) UpdateWithSections(_ context.Context, doc *models.CVDocument) (*models.CVDocument, error) {
	if f.sectionsErr != nil {
		return nil, f.sectionsErr
	}
	existing, ok := f.docs[doc.ID]
	if !ok || existing.UserID != doc.UserID {
		return nil, repositories.ErrCVDocumentNotFound
	}
	out := *doc
	assignFakeSectionIDs(out.Sections, out.ID)
	f.docs[out.ID] = &out
	f.lastUpdated = &out
	return &out, nil
}

func (f *fakeCVStore) DeleteForUser(_ context.Context, userID, id string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	doc, ok := f.docs[id]
	if !ok || doc.UserID != userID {
		return repositories.ErrCVDocumentNotFound
	}
	delete(f.docs, id)
	return nil
}

func assignFakeSectionIDs(sections []*models.CVSection, cvID string) {
	for i, section := range sections {
		section.CVID = cvID
		if section.ID == "" {
			section.ID = fmt.Sprintf("55555555-5555-4555-8555-%012d", i+1)
		}
	}
}

type fakeCVResumeLookup struct {
	resume *models.Resume
	err    error
}

func (f *fakeCVResumeLookup) GetByID(_ context.Context, _ string) (*models.Resume, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.resume, nil
}

func requireCVStatus(t *testing.T, err error, status int) {
	t.Helper()
	appErr, ok := utils.AsAppError(err)
	if !ok {
		t.Fatalf("expected an AppError, got %T: %v", err, err)
	}
	if appErr.StatusCode != status {
		t.Fatalf("expected status %d, got %d (%v)", status, appErr.StatusCode, err)
	}
}

func TestCVServiceCreateRequiresTitle(t *testing.T) {
	service := NewCVService(newFakeCVStore(), &fakeCVResumeLookup{})
	_, err := service.Create(context.Background(), testCVUser, CreateCVInput{})
	requireCVStatus(t, err, 400)
}

func TestCVServiceCreateImportsResume(t *testing.T) {
	parsed := "Jane Doe\nSenior Backend Engineer"
	store := newFakeCVStore()
	service := NewCVService(store, &fakeCVResumeLookup{resume: &models.Resume{
		ID:         testCVResumeID,
		UserID:     testCVUser,
		ParsedText: &parsed,
	}})

	doc, err := service.Create(context.Background(), testCVUser, CreateCVInput{ResumeID: testCVResumeID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Title != "Imported CV" {
		t.Fatalf("expected default imported title, got %q", doc.Title)
	}
	if len(doc.Sections) != 1 {
		t.Fatalf("expected one imported section, got %d", len(doc.Sections))
	}
	section := doc.Sections[0]
	if section.SectionType != models.CVSectionSummary || section.Title != "Imported Resume" {
		t.Fatalf("unexpected imported section: %+v", section)
	}
	if section.SortOrder != 0 {
		t.Fatalf("expected sort order 0, got %d", section.SortOrder)
	}
	if section.Content["text"] != parsed {
		t.Fatalf("expected parsed text in content, got %+v", section.Content)
	}
}

func TestCVServiceCreateRejectsForeignResume(t *testing.T) {
	parsed := "text"
	service := NewCVService(newFakeCVStore(), &fakeCVResumeLookup{resume: &models.Resume{
		ID:         testCVResumeID,
		UserID:     testCVOtherUser,
		ParsedText: &parsed,
	}})

	_, err := service.Create(context.Background(), testCVUser, CreateCVInput{
		Title:    "CV",
		ResumeID: testCVResumeID,
	})
	requireCVStatus(t, err, 400)
}

func TestCVServiceCreateRejectsResumeWithoutParsedText(t *testing.T) {
	service := NewCVService(newFakeCVStore(), &fakeCVResumeLookup{resume: &models.Resume{
		ID:     testCVResumeID,
		UserID: testCVUser,
	}})

	_, err := service.Create(context.Background(), testCVUser, CreateCVInput{
		Title:    "CV",
		ResumeID: testCVResumeID,
	})
	requireCVStatus(t, err, 400)
}

func TestCVServiceCreateAssignsSectionOrder(t *testing.T) {
	store := newFakeCVStore()
	service := NewCVService(store, &fakeCVResumeLookup{})

	doc, err := service.Create(context.Background(), testCVUser, CreateCVInput{
		Title: "Backend CV",
		Sections: []CVSectionInput{
			{SectionType: models.CVSectionSummary, Title: "  Summary  ", Content: map[string]any{"text": "a"}},
			{SectionType: models.CVSectionExperience, Title: "Experience"},
			{SectionType: models.CVSectionSkills, Title: "Skills"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(doc.Sections))
	}
	for i, section := range doc.Sections {
		if section.SortOrder != i {
			t.Fatalf("section %d: expected sort order %d, got %d", i, i, section.SortOrder)
		}
	}
	if doc.Sections[0].Title != "Summary" {
		t.Fatalf("expected trimmed title, got %q", doc.Sections[0].Title)
	}
}

func TestCVServiceCreateRejectsSectionIDs(t *testing.T) {
	service := NewCVService(newFakeCVStore(), &fakeCVResumeLookup{})
	_, err := service.Create(context.Background(), testCVUser, CreateCVInput{
		Title: "CV",
		Sections: []CVSectionInput{{
			ID:          testCVID,
			SectionType: models.CVSectionSummary,
			Title:       "Summary",
		}},
	})
	requireCVStatus(t, err, 400)
}

func TestCVServiceGetHidesForeignCV(t *testing.T) {
	store := newFakeCVStore(&models.CVDocument{ID: testCVID, UserID: testCVUser, Title: "Private"})
	service := NewCVService(store, &fakeCVResumeLookup{})

	_, err := service.Get(context.Background(), testCVOtherUser, testCVID)
	requireCVStatus(t, err, 404)

	doc, err := service.Get(context.Background(), testCVUser, testCVID)
	if err != nil {
		t.Fatalf("owner should read CV: %v", err)
	}
	if doc.ID != testCVID {
		t.Fatalf("unexpected CV: %+v", doc)
	}
}

func TestCVServiceUpdatePartialMetadata(t *testing.T) {
	store := newFakeCVStore(&models.CVDocument{
		ID:         testCVID,
		UserID:     testCVUser,
		Title:      "Old",
		TargetRole: "Engineer",
		TemplateID: "modern",
		Sections: []*models.CVSection{{
			ID:          "55555555-5555-4555-8555-000000000001",
			SectionType: models.CVSectionSummary,
			Title:       "Summary",
		}},
	})
	service := NewCVService(store, &fakeCVResumeLookup{})

	newTitle := "New"
	doc, err := service.Update(context.Background(), testCVUser, testCVID, UpdateCVInput{Title: &newTitle})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Title != "New" || doc.TargetRole != "Engineer" || doc.TemplateID != "modern" {
		t.Fatalf("unexpected merged document: %+v", doc)
	}
	if len(doc.Sections) != 1 {
		t.Fatalf("expected sections to be preserved, got %+v", doc.Sections)
	}
}

func TestCVServiceUpdateHidesForeignCV(t *testing.T) {
	store := newFakeCVStore(&models.CVDocument{ID: testCVID, UserID: testCVUser, Title: "Private"})
	service := NewCVService(store, &fakeCVResumeLookup{})

	title := "Hijacked"
	_, err := service.Update(context.Background(), testCVOtherUser, testCVID, UpdateCVInput{Title: &title})
	requireCVStatus(t, err, 404)
}

func TestCVServiceUpdateRejectsEmptyPatch(t *testing.T) {
	service := NewCVService(newFakeCVStore(), &fakeCVResumeLookup{})
	_, err := service.Update(context.Background(), testCVUser, testCVID, UpdateCVInput{})
	requireCVStatus(t, err, 400)
}

func TestCVServicePutSectionsAssignsOrderAndOwnership(t *testing.T) {
	store := newFakeCVStore(&models.CVDocument{ID: testCVID, UserID: testCVUser, Title: "CV"})
	service := NewCVService(store, &fakeCVResumeLookup{})

	doc, err := service.PutSections(context.Background(), testCVUser, testCVID, []CVSectionInput{
		{SectionType: models.CVSectionExperience, Title: "Experience"},
		{SectionType: models.CVSectionSummary, Title: "Summary"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(doc.Sections))
	}
	for i, section := range doc.Sections {
		if section.SortOrder != i {
			t.Fatalf("section %d: expected sort order %d, got %d", i, i, section.SortOrder)
		}
		if section.CVID != testCVID {
			t.Fatalf("expected section scoped to CV, got %q", section.CVID)
		}
	}

	_, err = service.PutSections(context.Background(), testCVOtherUser, testCVID, []CVSectionInput{
		{SectionType: models.CVSectionSummary, Title: "Summary"},
	})
	requireCVStatus(t, err, 404)
}

func TestCVServicePutSectionsMapsSectionNotFound(t *testing.T) {
	store := newFakeCVStore(&models.CVDocument{ID: testCVID, UserID: testCVUser, Title: "CV"})
	store.sectionsErr = repositories.ErrCVSectionNotFound
	service := NewCVService(store, &fakeCVResumeLookup{})

	_, err := service.PutSections(context.Background(), testCVUser, testCVID, []CVSectionInput{
		{SectionType: models.CVSectionSummary, Title: "Summary"},
	})
	requireCVStatus(t, err, 400)
}

func TestCVServiceDeleteHidesForeignCV(t *testing.T) {
	store := newFakeCVStore(&models.CVDocument{ID: testCVID, UserID: testCVUser, Title: "CV"})
	service := NewCVService(store, &fakeCVResumeLookup{})

	if err := service.Delete(context.Background(), testCVOtherUser, testCVID); err == nil {
		t.Fatal("expected cross-user delete to fail")
	} else {
		requireCVStatus(t, err, 404)
	}

	if err := service.Delete(context.Background(), testCVUser, testCVID); err != nil {
		t.Fatalf("owner delete should succeed: %v", err)
	}
}

func TestCVServiceListReturnsOnlyOwnedCVs(t *testing.T) {
	store := newFakeCVStore(
		&models.CVDocument{ID: testCVID, UserID: testCVUser, Title: "Mine"},
		&models.CVDocument{ID: "66666666-6666-4666-8666-666666666666", UserID: testCVOtherUser, Title: "Theirs"},
	)
	service := NewCVService(store, &fakeCVResumeLookup{})

	docs, err := service.List(context.Background(), testCVUser, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 1 || docs[0].Title != "Mine" {
		t.Fatalf("unexpected list: %+v", docs)
	}
}
