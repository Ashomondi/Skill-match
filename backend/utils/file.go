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

var magicBytes = map[string][]byte{
	".pdf":  []byte("%PDF-"),
	".docx": {0x50, 0x4B, 0x03, 0x04},
	".doc":  {0xD0, 0xCF, 0x11, 0xE0},
}

func ValidateResumeFile(filename, contentType string, size int64, content []byte) error {
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

	if err := ValidateFileContent(filename, content); err != nil {
		return err
	}

	return nil
}

func ValidateFileContent(filename string, content []byte) error {
	ext := strings.ToLower(filepath.Ext(filename))
	signature, ok := magicBytes[ext]
	if !ok {
		return fmt.Errorf("no known signature for extension %q", ext)
	}

	if len(content) < len(signature) {
		return fmt.Errorf("file is too small to be a valid %s (possibly corrupted or empty)", ext)
	}

	if !bytesEqual(content[:len(signature)], signature) {
		return fmt.Errorf("file content does not match expected %s format — file may be corrupted or mislabeled", ext)
	}

	return nil
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func GenerateFileID(originalFilename string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generating random file id: %w", err)
	}
	ext := strings.ToLower(filepath.Ext(originalFilename))
	return hex.EncodeToString(buf) + ext, nil
}