package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"skill-match/backend/middleware"
	"skill-match/backend/services"
)

type TailorHandler struct{ ai *services.AIService }

func NewTailorHandler(ai *services.AIService) *TailorHandler { return &TailorHandler{ai: ai} }

type tailorRequest struct {
	ResumeID       string `json:"resume_id"`
	JobTitle       string `json:"job_title"`
	Company        string `json:"company"`
	JobDescription string `json:"job_description"`
	CurrentContent string `json:"current_content"`
}

func (h *TailorHandler) Generate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	var input tailorRequest
	if json.NewDecoder(r.Body).Decode(&input) != nil || strings.TrimSpace(input.ResumeID) == "" || strings.TrimSpace(input.JobTitle) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "resume_id and job_title are required"})
		return
	}
	prompt := "Tailor this user's CV for the job below. Return only a complete, professional CV in plain text, preserving truthful experience and improving relevance.\n\nJob: " + input.JobTitle + "\nCompany: " + input.Company + "\nDescription:\n" + input.JobDescription
	if strings.TrimSpace(input.CurrentContent) != "" {
		prompt += "\n\nCurrent edited CV:\n" + input.CurrentContent
	}
	response, err := h.ai.GenerateResponse(r.Context(), services.AIRequest{UserID: userID, Message: prompt, ResumeID: input.ResumeID})
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to tailor CV"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"content": response.Message})
}
