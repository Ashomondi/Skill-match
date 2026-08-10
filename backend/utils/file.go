package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
)


const MaxResumeFileSize = 5 * 1024 * 1024 

var allowedResumeExtensions = map[string]string{
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
}


func ValidateResumeFile(filename, contentType string, size int64) error {
	if size <= 0 {
		return fmt.Errorf("file is empty")
	}
	if size > MaxResumeFileSize {
		return fmt.Errorf("file exceeds maximum allowed size of %d bytes", MaxResumeFileSize)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	expectedType, ok := allowedResumeExtensions[ext]
	if !ok {
		return fmt.Errorf("unsupported file type %q: only .pdf, .doc, .docx are allowed", ext)
	}

	if contentType != expectedType {
		return fmt.Errorf("content type %q does not match expected type %q for extension %q", contentType, expectedType, ext)
	}

	return nil
}


func GenerateFileID(originalFilename string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generating random file id: %w", err)
	}
	ext := strings.ToLower(filepath.Ext(originalFilename))
	return hex.EncodeToString(buf) + ext, nil
}