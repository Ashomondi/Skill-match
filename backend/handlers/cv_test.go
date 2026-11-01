package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"skill-match/backend/middleware"
	"skill-match/backend/models"
	"skill-match/backend/repositories"
	"skill-match/backend/services"
	"skill-match/backend/utils"
)

const (
	cvTestSecret = "cv-handler-test-secret"
	cvTestUser   = "11111111-1111-4111-8111-111111111111"
	cvOtherUser  = "22222222-2222-4222-8222-222222222222"
	cvTestID     = "33333333-3333-4333-8333-333333333333"
	cvResumeID   = "99999999-9999-4999-8999-999999999999"
)

type handlerFakeCVStore struct {
	docs        map[string]*models.CVDocument
	order       []string
	createErr   error
	getErr      error
	listErr     error
	updateErr   error
	sectionsErr error
	deleteErr   error
}

func newHandlerFakeCVStore(docs ...*models.CVDocument) *handlerFakeCVStore {
	store := &handlerFakeCVStore{docs: map[string]*models.CVDocument{}}
	for _, doc := range docs {
		store.docs[doc.ID] = doc
		store.order = append(store.order, doc.ID)
	}
	return store
}

func (f *handlerFakeCVStore) Create(_ context.Context, doc *models.CVDocument) (*models.CVDocument, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	out := *doc
	if out.ID == "" {
		out.ID = cvTestID
	}
	assignHandlerCVSectionIDs(out.Sections, out.ID)
	f.docs[out.ID] = &out
	f.order = append(f.order, out.ID)
	return &out, nil
}

func (f *handlerFakeCVStore) GetByIDForUser(_ context.Context, userID, id string) (*models.CVDocument, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	doc, ok := f.docs[id]
	if !ok || doc.UserID != userID {
		return nil, repositories.ErrCVDocumentNotFound
	}
	return doc, nil
}

func (f *handlerFakeCVStore) ListByUserID(_ context.Context, userID string, _ int) ([]*models.CVDocument, error) {
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

func (f *handlerFakeCVStore) Update(_ context.Context, doc *models.CVDocument) (*models.CVDocument, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	existing, ok := f.docs[doc.ID]
	if !ok || existing.UserID != doc.UserID {
		return nil, repositories.ErrCVDocumentNotFound
	}
	out := *doc
	f.docs[doc.ID] = &out
	return &out, nil
}

func (f *handlerFakeCVStore) UpdateWithSections(_ context.Context, doc *models.CVDocument) (*models.CVDocument, error) {
	if f.sectionsErr != nil {
		return nil, f.sectionsErr
	}
	existing, ok := f.docs[doc.ID]
	if !ok || existing.UserID != doc.UserID {
		return nil, repositories.ErrCVDocumentNotFound
	}
	out := *doc
	assignHandlerCVSectionIDs(out.Sections, out.ID)
	f.docs[out.ID] = &out
	return &out, nil
}

func (f *handlerFakeCVStore) DeleteForUser(_ context.Context, userID, id string) error {
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

func assignHandlerCVSectionIDs(sections []*models.CVSection, cvID string) {
	for i, section := range sections {
		section.CVID = cvID
		if section.ID == "" {
			section.ID = fmt.Sprintf("55555555-5555-4555-8555-%012d", i+1)
		}
	}
}

type handlerFakeCVResumeLookup struct {
	resume *models.Resume
	err    error
}

func (f *handlerFakeCVResumeLookup) GetByID(_ context.Context, _ string) (*models.Resume, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.resume, nil
}

func cvBearerToken(t *testing.T, userID string) string {
	t.Helper()
	token, err := utils.NewJWTManager(cvTestSecret, time.Hour).GenerateToken(userID, "user@example.com")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return token
}

func cvHandler(handler *CVHandler, name string) http.Handler {
	auth := middleware.Auth(utils.NewJWTManager(cvTestSecret, time.Hour))
	switch name {
	case "list":
		return auth(http.HandlerFunc(handler.List))
	case "create":
		return auth(http.HandlerFunc(handler.Create))
	case "get":
		return auth(http.HandlerFunc(handler.Get))
	case "update":
		return auth(http.HandlerFunc(handler.Update))
	case "sections":
		return auth(http.HandlerFunc(handler.PutSections))
	case "delete":
		return auth(http.HandlerFunc(handler.Delete))
	default:
		panic("unknown CV handler: " + name)
	}
}

func doCVRequest(t *testing.T, handler http.Handler, method, path, id, userID, body string, authorized bool) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}
	if id != "" {
		req.SetPathValue("id", id)
	}
	if authorized {
		req.Header.Set("Authorization", "Bearer "+cvBearerToken(t, userID))
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

type cvEnvelope struct {
	Data models.CVDocument `json:"data"`
}

type cvListEnvelope struct {
	Data []models.CVDocument `json:"data"`
}

func decodeCVEnvelope(t *testing.T, rec *httptest.ResponseRecorder) cvEnvelope {
	t.Helper()
	var envelope cvEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v (body=%s)", err, rec.Body.String())
	}
	return envelope
}

func TestCVHandlerListRequiresAuth(t *testing.T) {
	h := NewCVHandler(services.NewCVService(newHandlerFakeCVStore(), &handlerFakeCVResumeLookup{}))
	rec := doCVRequest(t, cvHandler(h, "list"), http.MethodGet, "/api/cvs", "", cvTestUser, "", false)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCVHandlerListReturnsOnlyOwnedCVs(t *testing.T) {
	store := newHandlerFakeCVStore(
		&models.CVDocument{ID: cvTestID, UserID: cvTestUser, Title: "Mine"},
		&models.CVDocument{ID: "66666666-6666-4666-8666-666666666666", UserID: cvOtherUser, Title: "Theirs"},
	)
	h := NewCVHandler(services.NewCVService(store, &handlerFakeCVResumeLookup{}))

	rec := doCVRequest(t, cvHandler(h, "list"), http.MethodGet, "/api/cvs", "", cvTestUser, "", true)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var envelope cvListEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(envelope.Data) != 1 || envelope.Data[0].Title != "Mine" {
		t.Fatalf("unexpected list: %+v", envelope.Data)
	}
}

func TestCVHandlerCreateRequiresAuth(t *testing.T) {
	h := NewCVHandler(services.NewCVService(newHandlerFakeCVStore(), &handlerFakeCVResumeLookup{}))
	rec := doCVRequest(t, cvHandler(h, "create"), http.MethodPost, "/api/cvs", "", cvTestUser, `{"title":"CV"}`, false)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCVHandlerCreateReturnsCreated(t *testing.T) {
	h := NewCVHandler(services.NewCVService(newHandlerFakeCVStore(), &handlerFakeCVResumeLookup{}))
	rec := doCVRequest(t, cvHandler(h, "create"), http.MethodPost, "/api/cvs", "", cvTestUser,
		`{"title":"Backend CV","target_role":"Go Engineer","sections":[{"section_type":"summary","title":"Summary","content":{"text":"hi"}},{"section_type":"skills","title":"Skills"}]}`, true)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	envelope := decodeCVEnvelope(t, rec)
	if envelope.Data.Title != "Backend CV" {
		t.Fatalf("unexpected title: %q", envelope.Data.Title)
	}
	if len(envelope.Data.Sections) != 2 {
		t.Fatalf("expected 2 sections, got %+v", envelope.Data.Sections)
	}
	if envelope.Data.Sections[0].SortOrder != 0 || envelope.Data.Sections[1].SortOrder != 1 {
		t.Fatalf("expected ascending sort order, got %+v", envelope.Data.Sections)
	}
}

func TestCVHandlerCreateImportsResume(t *testing.T) {
	parsed := "Jane Doe\nSenior Backend Engineer"
	h := NewCVHandler(services.NewCVService(newHandlerFakeCVStore(), &handlerFakeCVResumeLookup{
		resume: &models.Resume{ID: cvResumeID, UserID: cvTestUser, ParsedText: &parsed},
	}))

	rec := doCVRequest(t, cvHandler(h, "create"), http.MethodPost, "/api/cvs", "", cvTestUser,
		`{"title":"Imported","resume_id":"`+cvResumeID+`"}`, true)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	envelope := decodeCVEnvelope(t, rec)
	if len(envelope.Data.Sections) != 1 {
		t.Fatalf("expected imported section, got %+v", envelope.Data.Sections)
	}
	if envelope.Data.Sections[0].Content["text"] != parsed {
		t.Fatalf("unexpected imported content: %+v", envelope.Data.Sections[0].Content)
	}
}

func TestCVHandlerCreateRejectsInvalidJSON(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "malformed json", body: `not-json`},
		{name: "unknown field", body: `{"title":"CV","unexpected":true}`},
		{name: "trailing json", body: `{"title":"a"}{"title":"b"}`},
		{name: "missing title", body: `{}`},
		{name: "invalid section type", body: `{"title":"CV","sections":[{"section_type":"nope","title":"X"}]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewCVHandler(services.NewCVService(newHandlerFakeCVStore(), &handlerFakeCVResumeLookup{}))
			rec := doCVRequest(t, cvHandler(h, "create"), http.MethodPost, "/api/cvs", "", cvTestUser, tt.body, true)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCVHandlerGetHidesForeignCV(t *testing.T) {
	store := newHandlerFakeCVStore(&models.CVDocument{ID: cvTestID, UserID: cvTestUser, Title: "Private"})
	h := NewCVHandler(services.NewCVService(store, &handlerFakeCVResumeLookup{}))

	rec := doCVRequest(t, cvHandler(h, "get"), http.MethodGet, "/api/cvs/"+cvTestID, cvTestID, cvOtherUser, "", true)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for another user's CV, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = doCVRequest(t, cvHandler(h, "get"), http.MethodGet, "/api/cvs/"+cvTestID, cvTestID, cvTestUser, "", true)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for owner, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCVHandlerUpdateMetadata(t *testing.T) {
	store := newHandlerFakeCVStore(&models.CVDocument{
		ID: cvTestID, UserID: cvTestUser, Title: "Old", TargetRole: "Engineer",
	})
	h := NewCVHandler(services.NewCVService(store, &handlerFakeCVResumeLookup{}))

	rec := doCVRequest(t, cvHandler(h, "update"), http.MethodPut, "/api/cvs/"+cvTestID, cvTestID, cvTestUser,
		`{"title":"New","template_id":"modern"}`, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	envelope := decodeCVEnvelope(t, rec)
	if envelope.Data.Title != "New" || envelope.Data.TemplateID != "modern" {
		t.Fatalf("unexpected updated CV: %+v", envelope.Data)
	}
	if envelope.Data.TargetRole != "Engineer" {
		t.Fatalf("expected untouched target role, got %q", envelope.Data.TargetRole)
	}
}

func TestCVHandlerUpdateRejectsEmptyPatch(t *testing.T) {
	h := NewCVHandler(services.NewCVService(newHandlerFakeCVStore(), &handlerFakeCVResumeLookup{}))
	rec := doCVRequest(t, cvHandler(h, "update"), http.MethodPut, "/api/cvs/"+cvTestID, cvTestID, cvTestUser, `{}`, true)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCVHandlerPutSectionsPreservesOrder(t *testing.T) {
	store := newHandlerFakeCVStore(&models.CVDocument{ID: cvTestID, UserID: cvTestUser, Title: "CV"})
	h := NewCVHandler(services.NewCVService(store, &handlerFakeCVResumeLookup{}))

	rec := doCVRequest(t, cvHandler(h, "sections"), http.MethodPut, "/api/cvs/"+cvTestID+"/sections", cvTestID, cvTestUser,
		`{"sections":[{"section_type":"experience","title":"Experience"},{"section_type":"education","title":"Education"},{"section_type":"summary","title":"Summary"}]}`, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	envelope := decodeCVEnvelope(t, rec)
	if len(envelope.Data.Sections) != 3 {
		t.Fatalf("expected 3 sections, got %+v", envelope.Data.Sections)
	}
	for i, section := range envelope.Data.Sections {
		if section.SortOrder != i {
			t.Fatalf("section %d: expected sort order %d, got %d", i, i, section.SortOrder)
		}
	}
	if envelope.Data.Sections[0].Title != "Experience" || envelope.Data.Sections[2].Title != "Summary" {
		t.Fatalf("unexpected section ordering: %+v", envelope.Data.Sections)
	}
}

func TestCVHandlerPutSectionsRequiresSectionsField(t *testing.T) {
	h := NewCVHandler(services.NewCVService(newHandlerFakeCVStore(), &handlerFakeCVResumeLookup{}))
	rec := doCVRequest(t, cvHandler(h, "sections"), http.MethodPut, "/api/cvs/"+cvTestID+"/sections", cvTestID, cvTestUser, `{}`, true)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCVHandlerDeleteReturnsNoContent(t *testing.T) {
	store := newHandlerFakeCVStore(&models.CVDocument{ID: cvTestID, UserID: cvTestUser, Title: "CV"})
	h := NewCVHandler(services.NewCVService(store, &handlerFakeCVResumeLookup{}))

	rec := doCVRequest(t, cvHandler(h, "delete"), http.MethodDelete, "/api/cvs/"+cvTestID, cvTestID, cvTestUser, "", true)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = doCVRequest(t, cvHandler(h, "delete"), http.MethodDelete, "/api/cvs/"+cvTestID, cvTestID, cvOtherUser, "", true)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for another user's CV, got %d", rec.Code)
	}
}

func TestCVHandlerRejectsMalformedID(t *testing.T) {
	h := NewCVHandler(services.NewCVService(newHandlerFakeCVStore(), &handlerFakeCVResumeLookup{}))
	rec := doCVRequest(t, cvHandler(h, "get"), http.MethodGet, "/api/cvs/not-a-uuid", "not-a-uuid", cvTestUser, "", true)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed id, got %d: %s", rec.Code, rec.Body.String())
	}
}
