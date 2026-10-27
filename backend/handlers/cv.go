package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"skill-match/backend/middleware"
	"skill-match/backend/models"
	"skill-match/backend/services"
	"skill-match/backend/utils"
)

// maxCVBodyBytes caps structured CV request bodies. Sections and theme content
// are bounded in the service as well.
const maxCVBodyBytes = 256 << 10

// CVHandler exposes structured CV CRUD endpoints.
type CVHandler struct {
	service *services.CVService
}

func NewCVHandler(service *services.CVService) *CVHandler {
	return &CVHandler{service: service}
}

type cvSectionRequest struct {
	ID          string               `json:"id,omitempty"`
	SectionType models.CVSectionType `json:"section_type"`
	Title       string               `json:"title"`
	Content     map[string]any       `json:"content"`
}

type createCVRequest struct {
	Title       string             `json:"title"`
	TargetRole  string             `json:"target_role"`
	TemplateID  string             `json:"template_id"`
	ThemeConfig map[string]any     `json:"theme_config"`
	IsPrimary   bool               `json:"is_primary"`
	ResumeID    string             `json:"resume_id"`
	Sections    []cvSectionRequest `json:"sections"`
}

type updateCVRequest struct {
	Title       *string         `json:"title"`
	TargetRole  *string         `json:"target_role"`
	TemplateID  *string         `json:"template_id"`
	ThemeConfig *map[string]any `json:"theme_config"`
	IsPrimary   *bool           `json:"is_primary"`
}

type putSectionsRequest struct {
	Sections *[]cvSectionRequest `json:"sections"`
}

func toCVSectionInputs(sections []cvSectionRequest) []services.CVSectionInput {
	out := make([]services.CVSectionInput, 0, len(sections))
	for _, section := range sections {
		out = append(out, services.CVSectionInput{
			ID:          section.ID,
			SectionType: section.SectionType,
			Title:       section.Title,
			Content:     section.Content,
		})
	}
	return out
}

// List handles GET /api/cvs.
func (h *CVHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := cvUserID(w, r)
	if !ok {
		return
	}

	limit, ok := parseCVLimit(r.URL.Query().Get("limit"))
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "limit must be a non-negative integer"})
		return
	}

	docs, err := h.service.List(r.Context(), userID, limit)
	if err != nil {
		utils.WriteRequestError(w, r, err)
		return
	}
	utils.WriteSuccess(w, http.StatusOK, docs)
}

// Create handles POST /api/cvs.
func (h *CVHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := cvUserID(w, r)
	if !ok {
		return
	}

	var req createCVRequest
	if !decodeCVBody(w, r, &req) {
		return
	}

	doc, err := h.service.Create(r.Context(), userID, services.CreateCVInput{
		Title:       req.Title,
		TargetRole:  req.TargetRole,
		TemplateID:  req.TemplateID,
		ThemeConfig: req.ThemeConfig,
		IsPrimary:   req.IsPrimary,
		ResumeID:    req.ResumeID,
		Sections:    toCVSectionInputs(req.Sections),
	})
	if err != nil {
		utils.WriteRequestError(w, r, err)
		return
	}
	utils.WriteSuccess(w, http.StatusCreated, doc)
}

// Get handles GET /api/cvs/{id}.
func (h *CVHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := cvUserID(w, r)
	if !ok {
		return
	}

	doc, err := h.service.Get(r.Context(), userID, r.PathValue("id"))
	if err != nil {
		utils.WriteRequestError(w, r, err)
		return
	}
	utils.WriteSuccess(w, http.StatusOK, doc)
}

// Update handles PUT /api/cvs/{id} (partial metadata update).
func (h *CVHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := cvUserID(w, r)
	if !ok {
		return
	}

	var req updateCVRequest
	if !decodeCVBody(w, r, &req) {
		return
	}

	doc, err := h.service.Update(r.Context(), userID, r.PathValue("id"), services.UpdateCVInput{
		Title:       req.Title,
		TargetRole:  req.TargetRole,
		TemplateID:  req.TemplateID,
		ThemeConfig: req.ThemeConfig,
		IsPrimary:   req.IsPrimary,
	})
	if err != nil {
		utils.WriteRequestError(w, r, err)
		return
	}
	utils.WriteSuccess(w, http.StatusOK, doc)
}

// PutSections handles PUT /api/cvs/{id}/sections. Array position defines order.
func (h *CVHandler) PutSections(w http.ResponseWriter, r *http.Request) {
	userID, ok := cvUserID(w, r)
	if !ok {
		return
	}

	var req putSectionsRequest
	if !decodeCVBody(w, r, &req) {
		return
	}
	if req.Sections == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "sections is required"})
		return
	}

	doc, err := h.service.PutSections(r.Context(), userID, r.PathValue("id"), toCVSectionInputs(*req.Sections))
	if err != nil {
		utils.WriteRequestError(w, r, err)
		return
	}
	utils.WriteSuccess(w, http.StatusOK, doc)
}

// Delete handles DELETE /api/cvs/{id}.
func (h *CVHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := cvUserID(w, r)
	if !ok {
		return
	}

	if err := h.service.Delete(r.Context(), userID, r.PathValue("id")); err != nil {
		utils.WriteRequestError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func cvUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return "", false
	}
	return userID, true
}

func parseCVLimit(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, true
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 0 {
		return 0, false
	}
	return limit, true
}

func decodeCVBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxCVBodyBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid CV payload"})
		return false
	}
	// Reject a second JSON value in the same body.
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid CV payload"})
		return false
	}
	return true
}
