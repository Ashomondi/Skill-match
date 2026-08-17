package utils

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

const MaxResumeSize int64 = 5 * 1024 * 1024 // 5 MB

var (
	ErrInvalidFileType = errors.New("invalid resume file type")
	ErrFileTooLarge    = errors.New("resume file is too large")
	ErrEmptyFile       = errors.New("resume file is empty")
)

func ValidateResume(filename string, size int64, content io.Reader) error {
	if size <= 0 {
		return ErrEmptyFile
	}

	if size > MaxResumeSize {
		return ErrFileTooLarge
	}

	extension := strings.ToLower(filepath.Ext(filename))

	if extension != ".pdf" && extension != ".docx" {
		return ErrInvalidFileType
	}

	// Read enough bytes to detect the content type.
	buffer := make([]byte, 512)

	n, err := content.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to inspect file: %w", err)
	}

	contentType := http.DetectContentType(buffer[:n])

	if extension == ".pdf" && contentType != "application/pdf" {
		return ErrInvalidFileType
	}

	// DOCX files are ZIP-based and are commonly detected as
	// application/zip or application/x-zip-compressed.
	if extension == ".docx" &&
		contentType != "application/zip" &&
		contentType != "application/x-zip-compressed" {
		return ErrInvalidFileType
	}

	return nil
}
