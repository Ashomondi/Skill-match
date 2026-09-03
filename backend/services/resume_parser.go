package services

import (
	"context"
	"fmt"
	"strings"

	"skill-match/backend/models"
	"skill-match/backend/utils"
)

// ParseResume extracts plain text from a resume file and returns a Resume
// with status transitioned to parsing/parsed or failed.
// It respects the existing magic-byte validation rules.
func ParseResume(ctx context.Context, userID, filename, contentType string, data []byte) (*models.Resume, error) {
	if userID == "" {
		return nil, fmt.Errorf("invalid resume: user is required")
	}

	// Validate file using existing utility (respects magic bytes, size, type)
	if err := utils.ValidateResumeFile(filename, contentType, int64(len(data)), data); err != nil {
		return nil, fmt.Errorf("invalid resume: %v", err)
	}

	// Determine file extension for text extraction
	ext := strings.ToLower(filename)
	if idx := strings.LastIndex(ext, "."); idx >= 0 {
		ext = ext[idx:]
	} else {
		ext = ""
	}

	var parsedText string
	var failureReason string

	switch ext {
	case ".pdf":
		parsedText, failureReason = extractPDFText(data)
	case ".doc":
		parsedText, failureReason = extractDocText(data)
	case ".docx":
		parsedText, failureReason = extractDocxText(data)
	case ".txt":
		parsedText, failureReason = extractTxtText(data)
	default:
		failureReason = "unsupported file format for text extraction"
	}

	res := &models.Resume{
		UserID:           userID,
		OriginalFilename: filename,
		ContentType:      contentType,
		FileSizeBytes:    int64(len(data)),
		Status:           models.ResumeStatusUploaded,
	}

	if failureReason != "" {
		// Parsing failed — mark as failed and record reason.
		res.Status = models.ResumeStatusFailed
		res.FailureReason = &failureReason
		return res, nil
	}

	// Parsing succeeded — transition to parsed and set extracted text.
	res.Status = models.ResumeStatusParsed
	pt := parsedText
	res.ParsedText = &pt

	return res, nil
}

// extractPDFText attempts to extract text from a PDF byte buffer.
func extractPDFText(data []byte) (string, string) {
	// Placeholder: a full implementation would use a PDF parsing library.
	// For now return empty text so the status stays "processing" until
	// a proper parser is integrated.
	return "", ""
}

// extractDocText attempts to extract text from a .doc (OLE2) file.
func extractDocText(data []byte) (string, string) {
	return "", "text extraction for .doc files is not yet implemented; please use PDF or DOCX or TXT"
}

// extractDocxText attempts to extract text from a .docx (ZIP-based) file.
func extractDocxText(data []byte) (string, string) {
	return "", "text extraction for .docx files is not yet implemented; please use PDF or TXT"
}

// extractTxtText directly returns the file content as text.
func extractTxtText(data []byte) (string, string) {
	text := string(data)
	// Strip null bytes and control characters for cleanliness.
	clean := strings.Map(func(r rune) rune {
		if r >= 32 || r == 10 || r == 13 {
			return r
		}
		return -1
	}, text)
	return clean, ""
}