package services

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
	"skill-match/backend/repositories"
)

// Errors returned by the resume parsing service.
var (
	ErrResumeParseFailed       = errors.New("resume parsing failed")
	ErrResumeUnauthorized      = errors.New("resume does not belong to user")
	ErrResumeParserNotFound    = errors.New("resume not found")
	ErrUnsupportedResumeFormat = errors.New("unsupported resume format")
	ErrEmptyResumeText         = errors.New("extracted resume text is empty")
)

// ResumeParserRepository defines the repository operations
// needed by the parsing service.
//
// This matches the existing repositories.ResumeRepository.
type ResumeParserRepository interface {
	GetByID(
		ctx context.Context,
		id string,
	) (*repositories.Resume, error)

	UpdateStatus(
		ctx context.Context,
		id string,
		status repositories.ResumeStatus,
		parsedText *string,
		failureReason *string,
	) error
}

// ResumeDownloader defines the S3 operation needed to
// download a resume.
type ResumeDownloader interface {
	Download(
		ctx context.Context,
		key string,
	) ([]byte, error)
}

// ResumeParser extracts text from PDF and DOCX resumes
// and stores the extracted text in CockroachDB.
type ResumeParser struct {
	repository ResumeParserRepository
	storage    ResumeDownloader
}

// NewResumeParser creates a new resume parser.
func NewResumeParser(
	repository ResumeParserRepository,
	storage ResumeDownloader,
) *ResumeParser {
	return &ResumeParser{
		repository: repository,
		storage:    storage,
	}
}

// ParseResumeInput contains the authenticated user and resume ID.
type ParseResumeInput struct {
	UserID   string
	ResumeID string
}

// ParseResume downloads a resume from S3, extracts its text,
// and updates its processing status.
func (p *ResumeParser) ParseResume(
	ctx context.Context,
	input ParseResumeInput,
) (string, error) {

	// Retrieve the resume metadata.
	resume, err := p.repository.GetByID(
		ctx,
		input.ResumeID,
	)
	if err != nil {
		if errors.Is(err, repositories.ErrResumeNotFound) {
			return "", ErrResumeParserNotFound
		}

		return "", fmt.Errorf(
			"%w: failed to find resume: %v",
			ErrResumeParseFailed,
			err,
		)
	}

	if resume == nil {
		return "", ErrResumeParserNotFound
	}

	// Verify ownership.
	if resume.UserID != input.UserID {
		return "", ErrResumeUnauthorized
	}

	if strings.TrimSpace(resume.S3Key) == "" {
		return "", fmt.Errorf(
			"%w: resume has no S3 key",
			ErrResumeParseFailed,
		)
	}

	// Mark the resume as currently being parsed.
	if err := p.repository.UpdateStatus(
		ctx,
		resume.ID,
		repositories.ResumeStatusParsing,
		nil,
		nil,
	); err != nil {
		return "", fmt.Errorf(
			"%w: failed to mark resume as parsing: %v",
			ErrResumeParseFailed,
			err,
		)
	}

	// Download the resume from S3.
	data, err := p.storage.Download(
		ctx,
		resume.S3Key,
	)
	if err != nil {
		return "", p.markParsingFailed(
			ctx,
			resume.ID,
			fmt.Sprintf("failed to download resume: %v", err),
		)
	}

	// Extract the file extension.
	extension := strings.ToLower(
		filepath.Ext(resume.OriginalFilename),
	)

	// Extract text based on file type.
	var text string

	switch extension {
	case ".pdf":
		text, err = extractPDFText(data)

	case ".docx":
		text, err = extractDOCXText(data)

	default:
		return "", p.markParsingFailed(
			ctx,
			resume.ID,
			fmt.Sprintf(
				"unsupported resume format: %s",
				extension,
			),
		)
	}

	if err != nil {
		return "", p.markParsingFailed(
			ctx,
			resume.ID,
			fmt.Sprintf(
				"text extraction failed: %v",
				err,
			),
		)
	}

	text = strings.TrimSpace(text)

	if text == "" {
		return "", p.markParsingFailed(
			ctx,
			resume.ID,
			ErrEmptyResumeText.Error(),
		)
	}

	// Store extracted text and mark the resume as parsed.
	if err := p.repository.UpdateStatus(
		ctx,
		resume.ID,
		repositories.ResumeStatusParsed,
		&text,
		nil,
	); err != nil {
		return "", fmt.Errorf(
			"%w: failed to store extracted text: %v",
			ErrResumeParseFailed,
			err,
		)
	}

	return text, nil
}

// markParsingFailed changes the resume status to "failed"
// and records the reason.
func (p *ResumeParser) markParsingFailed(
	ctx context.Context,
	resumeID string,
	reason string,
) error {

	failureReason := strings.TrimSpace(reason)

	if failureReason == "" {
		failureReason = ErrResumeParseFailed.Error()
	}

	err := p.repository.UpdateStatus(
		ctx,
		resumeID,
		repositories.ResumeStatusFailed,
		nil,
		&failureReason,
	)

	if err != nil {
		return fmt.Errorf(
			"%w: %s; additionally failed to update status: %v",
			ErrResumeParseFailed,
			failureReason,
			err,
		)
	}

	return fmt.Errorf(
		"%w: %s",
		ErrResumeParseFailed,
		failureReason,
	)
}

// extractPDFText extracts plain text from a PDF.
//
// ledongthuc/pdf supports plain-text extraction through
// Reader.GetPlainText().
func extractPDFText(data []byte) (string, error) {

	if len(data) == 0 {
		return "", ErrEmptyResumeText
	}

	reader := bytes.NewReader(data)

	pdfReader, err := pdf.NewReader(
		reader,
		int64(len(data)),
	)
	if err != nil {
		return "", fmt.Errorf(
			"opening PDF: %w",
			err,
		)
	}

	textReader, err := pdfReader.GetPlainText()
	if err != nil {
		return "", fmt.Errorf(
			"extracting PDF text: %w",
			err,
		)
	}

	text, err := io.ReadAll(textReader)
	if err != nil {
		return "", fmt.Errorf(
			"reading extracted PDF text: %w",
			err,
		)
	}

	return string(text), nil
}

// extractDOCXText extracts plain text from the
// word/document.xml file inside a DOCX archive.
func extractDOCXText(data []byte) (string, error) {

	if len(data) == 0 {
		return "", ErrEmptyResumeText
	}

	reader, err := zip.NewReader(
		bytes.NewReader(data),
		int64(len(data)),
	)
	if err != nil {
		return "", fmt.Errorf(
			"opening DOCX archive: %w",
			err,
		)
	}

	var documentFile *zip.File

	for _, file := range reader.File {
		if file.Name == "word/document.xml" {
			documentFile = file
			break
		}
	}

	if documentFile == nil {
		return "", errors.New(
			"DOCX document.xml not found",
		)
	}

	fileReader, err := documentFile.Open()
	if err != nil {
		return "", fmt.Errorf(
			"opening DOCX document.xml: %w",
			err,
		)
	}
	defer fileReader.Close()

	decoder := xml.NewDecoder(fileReader)

	var output strings.Builder

	for {
		token, err := decoder.Token()

		if err == io.EOF {
			break
		}

		if err != nil {
			return "", fmt.Errorf(
				"reading DOCX XML: %w",
				err,
			)
		}

		switch element := token.(type) {

		case xml.StartElement:

			switch element.Name.Local {

			case "t":
				var text string

				if err := decoder.DecodeElement(
					&text,
					&element,
				); err != nil {
					return "", fmt.Errorf(
						"reading DOCX text: %w",
						err,
					)
				}

				output.WriteString(text)

			case "tab":
				output.WriteByte('\t')

			case "br":
				output.WriteByte('\n')
			}

		case xml.EndElement:

			// Every paragraph gets a newline.
			if element.Name.Local == "p" {
				output.WriteByte('\n')
			}
		}
	}

	return output.String(), nil
}
