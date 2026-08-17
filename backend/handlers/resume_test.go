package handlers

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"skill-match/backend/middleware"
	"skill-match/backend/models"
	"skill-match/backend/repositories"
	"skill-match/backend/services"
	"skill-match/backend/utils"
)

type handlerFakeStorage struct {
	objects map[string][]byte
}

func newHandlerFakeStorage() *handlerFakeStorage {
	return &handlerFakeStorage{objects: map[string][]byte{}}
}

func (f *handlerFakeStorage) Put(_ context.Context, key string, body []byte, _ string) error {
	f.objects[key] = body
	return nil
}
func (f *handlerFakeStorage) PresignDownload(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://presigned/" + key, nil
}
func (f *handlerFakeStorage) Delete(_ context.Context, key string) error {
	delete(f.objects, key)
	return nil
}
func (f *handlerFakeStorage) Key(userID, fileID string) string {
	return "resumes/" + userID + "/" + fileID
}

type handlerFakeResumeRepo struct {
	byID   map[string]*models.Resume
	byUser map[string][]*models.Resume
	seq    int
}

func newHandlerFakeResumeRepo() *handlerFakeResumeRepo {
	return &handlerFakeResumeRepo{
		byID:   map[string]*models.Resume{},
		byUser: map[string][]*models.Resume{},
	}
}

func (f *handlerFakeResumeRepo) Create(_ context.Context, r *models.Resume) (*models.Resume, error) {
	f.seq++
	r.ID = fmt.Sprintf("res-%d", f.seq)
	f.byID[r.ID] = r
	f.byUser[r.UserID] = append(f.byUser[r.UserID], r)
	return r, nil
}
func (f *handlerFakeResumeRepo) GetByID(_ context.Context, id string) (*models.Resume, error) {
	if r, ok := f.byID[id]; ok {
		return r, nil
	}
	return nil, repositories.ErrResumeNotFound
}
func (f *handlerFakeResumeRepo) ListByUserID(_ context.Context, userID string, _ int) ([]*models.Resume, error) {
	return f.byUser[userID], nil
}
func (f *handlerFakeResumeRepo) Delete(_ context.Context, id string) error {
	r, ok := f.byID[id]
	if !ok {
		return repositories.ErrResumeNotFound
	}
	delete(f.byID, id)

	userList := f.byUser[r.UserID]
	for i, item := range userList {
		if item.ID == id {
			f.byUser[r.UserID] = append(userList[:i], userList[i+1:]...)
			break
		}
	}
	return nil
}

const resumeTestSecret = "resume-handler-test-secret"

func newTestResumeHandler() (*ResumeHandler, *handlerFakeResumeRepo, *handlerFakeStorage) {
	storage := newHandlerFakeStorage()
	repo := newHandlerFakeResumeRepo()
	return NewResumeHandler(services.NewResumeService(repo, storage)), repo, storage
}

func bearerToken(t *testing.T, userID string) string {
	t.Helper()
	token, err := utils.NewJWTManager(resumeTestSecret, time.Hour).GenerateToken(userID, "user@example.com")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return token
}

// multipartRequest builds a POST /api/resumes multipart body with a file part
// carrying the given content type, plus an optional extra form field.
func multipartRequest(t *testing.T, filename, contentType, content, extraField, extraValue string) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="resume"; filename="%s"`, filename))
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("create part: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if extraField != "" {
		if err := writer.WriteField(extraField, extraValue); err != nil {
			t.Fatalf("write field: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/resumes", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func performAuthorized(t *testing.T, h http.Handler, userID string, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, userID))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestResumeCreateRequiresAuth(t *testing.T) {
	h, _, _ := newTestResumeHandler()
	handler := middleware.Auth(utils.NewJWTManager(resumeTestSecret, time.Hour))(http.HandlerFunc(h.Create))

	req := multipartRequest(t, "resume.pdf", "application/pdf", "%PDF-1.4", "", "")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", rec.Code)
	}
}

func TestResumeCreateUploads(t *testing.T) {
	h, repo, storage := newTestResumeHandler()
	handler := middleware.Auth(utils.NewJWTManager(resumeTestSecret, time.Hour))(http.HandlerFunc(h.Create))

	rec := performAuthorized(t, handler, "user-1", multipartRequest(t, "resume.pdf", "application/pdf", "%PDF-1.4", "", ""))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(storage.objects) != 1 {
		t.Fatalf("expected 1 stored object, got %d", len(storage.objects))
	}
	if len(repo.byUser["user-1"]) != 1 {
		t.Fatalf("expected 1 resume row, got %d", len(repo.byUser["user-1"]))
	}
}

func TestResumeCreateRejectsUnsupportedType(t *testing.T) {
	h, repo, storage := newTestResumeHandler()
	handler := middleware.Auth(utils.NewJWTManager(resumeTestSecret, time.Hour))(http.HandlerFunc(h.Create))

	rec := performAuthorized(t, handler, "user-1", multipartRequest(t, "resume.exe", "application/octet-stream", "MZ", "", ""))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unsupported type, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(storage.objects) != 0 || len(repo.byID) != 0 {
		t.Fatal("nothing should be stored for an invalid file")
	}
}

func TestResumeCreateReplaceSendsReplaceId(t *testing.T) {
	h, repo, _ := newTestResumeHandler()
	handler := middleware.Auth(utils.NewJWTManager(resumeTestSecret, time.Hour))(http.HandlerFunc(h.Create))

	first := performAuthorized(t, handler, "user-1", multipartRequest(t, "a.pdf", "application/pdf", "%PDF-1.4 a", "", ""))
	if first.Code != http.StatusCreated {
		t.Fatalf("setup upload: %d", first.Code)
	}
	oldID := repo.byUser["user-1"][0].ID

	rec := performAuthorized(t, handler, "user-1", multipartRequest(t, "b.pdf", "application/pdf", "%PDF-1.4 b", "replaceId", oldID))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	// replace removed the old row and added one new row
	if len(repo.byUser["user-1"]) != 1 {
		t.Fatalf("expected 1 resume after replace, got %d", len(repo.byUser["user-1"]))
	}
	if repo.byUser["user-1"][0].ID == oldID {
		t.Fatal("expected the old resume to be replaced")
	}
}

func TestResumeListRequiresAuth(t *testing.T) {
	h, _, _ := newTestResumeHandler()
	handler := middleware.Auth(utils.NewJWTManager(resumeTestSecret, time.Hour))(http.HandlerFunc(h.List))

	req := httptest.NewRequest(http.MethodGet, "/api/resumes", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestResumeListReturnsUsersResumes(t *testing.T) {
	h, repo, _ := newTestResumeHandler()
	createHandler := middleware.Auth(utils.NewJWTManager(resumeTestSecret, time.Hour))(http.HandlerFunc(h.Create))
	performAuthorized(t, createHandler, "user-1", multipartRequest(t, "a.pdf", "application/pdf", "%PDF-1.4 a", "", ""))
	performAuthorized(t, createHandler, "user-2", multipartRequest(t, "b.pdf", "application/pdf", "%PDF-1.4 b", "", ""))

	if len(repo.byUser["user-1"]) != 1 {
		t.Fatalf("setup: expected 1 resume for user-1")
	}

	listHandler := middleware.Auth(utils.NewJWTManager(resumeTestSecret, time.Hour))(http.HandlerFunc(h.List))
	rec := performAuthorized(t, listHandler, "user-1", httptest.NewRequest(http.MethodGet, "/api/resumes", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"name":"a.pdf"`) {
		t.Fatalf("expected user-1's resume in response, got %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "b.pdf") {
		t.Fatalf("response leaked another user's resume: %s", rec.Body.String())
	}
}

func TestResumeGetReturnsPresignedURL(t *testing.T) {
	h, repo, _ := newTestResumeHandler()
	createHandler := middleware.Auth(utils.NewJWTManager(resumeTestSecret, time.Hour))(http.HandlerFunc(h.Create))
	performAuthorized(t, createHandler, "user-1", multipartRequest(t, "a.pdf", "application/pdf", "%PDF-1.4 a", "", ""))
	id := repo.byUser["user-1"][0].ID

	getHandler := middleware.Auth(utils.NewJWTManager(resumeTestSecret, time.Hour))(http.HandlerFunc(h.Get))
	req := httptest.NewRequest(http.MethodGet, "/api/resumes/"+id, nil)
	req.SetPathValue("id", id)
	rec := performAuthorized(t, getHandler, "user-1", req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "https://presigned/") {
		t.Fatalf("expected a presigned url in response, got %s", rec.Body.String())
	}
}

func TestResumeDeleteRequiresAuth(t *testing.T) {
	h, _, _ := newTestResumeHandler()
	handler := middleware.Auth(utils.NewJWTManager(resumeTestSecret, time.Hour))(http.HandlerFunc(h.Delete))

	req := httptest.NewRequest(http.MethodDelete, "/api/resumes/some-id", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestResumeDeleteEnforcesOwnership(t *testing.T) {
	h, repo, storage := newTestResumeHandler()
	createHandler := middleware.Auth(utils.NewJWTManager(resumeTestSecret, time.Hour))(http.HandlerFunc(h.Create))
	performAuthorized(t, createHandler, "user-1", multipartRequest(t, "a.pdf", "application/pdf", "%PDF-1.4 a", "", ""))
	id := repo.byUser["user-1"][0].ID

	deleteHandler := middleware.Auth(utils.NewJWTManager(resumeTestSecret, time.Hour))(http.HandlerFunc(h.Delete))

	otherReq := httptest.NewRequest(http.MethodDelete, "/api/resumes/"+id, nil)
	otherReq.SetPathValue("id", id)
	// another user cannot delete it
	rec := performAuthorized(t, deleteHandler, "user-2", otherReq)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for another user, got %d", rec.Code)
	}
	if len(storage.objects) != 1 {
		t.Fatal("object should not be deleted on a denied request")
	}

	// owner can delete it
	ownReq := httptest.NewRequest(http.MethodDelete, "/api/resumes/"+id, nil)
	ownReq.SetPathValue("id", id)
	rec = performAuthorized(t, deleteHandler, "user-1", ownReq)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if len(storage.objects) != 0 {
		t.Fatal("expected object deleted after owner delete")
	}
}
