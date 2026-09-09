package services

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"

	"skill-match/backend/models"
	"skill-match/backend/utils"
)

// ParseResume extracts plain text from a resume file and returns a Resume
// with status transitioned to parsed or failed.
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

	if strings.TrimSpace(parsedText) == "" {
		// The file parsed without error but yielded no text (e.g. a
		// scanned/image-only PDF). Surface this instead of silently
		// marking the resume "parsed" with an empty body.
		failureReason = "no extractable text found; the file may be a scanned image"
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

// extractPDFText extracts plain text from a PDF byte buffer using the
// pure-Go ledongthuc/pdf parser (no CGO, no external binaries).
func extractPDFText(data []byte) (string, string) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Sprintf("failed to open PDF: %v", err)
	}

	text, err := reader.GetPlainText()
	if err != nil {
		return "", fmt.Sprintf("failed to extract PDF text: %v", err)
	}

	raw, err := io.ReadAll(text)
	if err != nil {
		return "", fmt.Sprintf("failed to read PDF text: %v", err)
	}

	return cleanExtractedText(string(raw)), ""
}

// extractDocText attempts to extract text from a legacy .doc (OLE2) file.
// The OLE2 compound-file format is not supported by a lightweight pure-Go
// parser; antiword/libreoffice would be needed. Fail with a clear message
// rather than returning misleading empty text.
func extractDocText(data []byte) (string, string) {
	return "", "text extraction for legacy .doc files is not supported; please convert to .docx, .pdf, or .txt"
}

// extractDocxText extracts text from a .docx (ZIP/XML) file by reading
// word/document.xml and joining the <w:t> run contents. Pure Go stdlib only.
func extractDocxText(data []byte) (string, string) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Sprintf("failed to open DOCX archive: %v", err)
	}

	doc := findZipEntry(zr, "word/document.xml")
	if doc == nil {
		return "", "DOCX file is missing word/document.xml"
	}

	rc, err := doc.Open()
	if err != nil {
		return "", fmt.Sprintf("failed to read DOCX document: %v", err)
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return "", fmt.Sprintf("failed to read DOCX document: %v", err)
	}

	var document struct {
		Body struct {
			Paragraphs []struct {
				Runs []struct {
					Text string `xml:"t"`
				} `xml:"r"`
			} `xml:"p"`
		} `xml:"body"`
	}
	if err := xml.Unmarshal(content, &document); err != nil {
		return "", fmt.Sprintf("failed to parse DOCX XML: %v", err)
	}

	var builder strings.Builder
	for _, para := range document.Body.Paragraphs {
		for _, run := range para.Runs {
			builder.WriteString(run.Text)
		}
		builder.WriteString("\n")
	}

	return cleanExtractedText(builder.String()), ""
}

// extractTxtText directly returns the file content as text.
func extractTxtText(data []byte) (string, string) {
	return cleanExtractedText(string(data)), ""
}

// findZipEntry returns the entry matching name, or nil.
func findZipEntry(zr *zip.Reader, name string) *zip.File {
	for _, f := range zr.File {
		if f.Name == name {
			return f
		}
	}
	return nil
}

// cleanExtractedText strips null bytes and other control characters (except
// newline/carriage-return) from extracted text for cleanliness.
func cleanExtractedText(text string) string {
	return strings.Map(func(r rune) rune {
		if r >= 32 || r == 10 || r == 13 {
			return r
		}
		return -1
	}, text)
}
