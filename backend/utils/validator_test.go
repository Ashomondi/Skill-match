package utils

import (
	"bytes"
	"errors"
	"testing"
)

func TestValidateResume(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		size     int64
		content  []byte
		wantErr  error
	}{
		{
			name:     "valid PDF",
			filename: "resume.pdf",
			size:     int64(len([]byte("%PDF-1.7\n"))),
			content:  []byte("%PDF-1.7\n"),
			wantErr:  nil,
		},
		{
			name:     "valid DOCX",
			filename: "resume.docx",
			size:     int64(len([]byte("PK\x03\x04"))),
			content:  []byte("PK\x03\x04"),
			wantErr:  nil,
		},
		{
			name:     "empty file",
			filename: "resume.pdf",
			size:     0,
			content:  []byte{},
			wantErr:  ErrEmptyFile,
		},
		{
			name:     "file too large",
			filename: "resume.pdf",
			size:     MaxResumeSize + 1,
			content:  []byte("%PDF-1.7\n"),
			wantErr:  ErrFileTooLarge,
		},
		{
			name:     "invalid file extension",
			filename: "resume.txt",
			size:     10,
			content:  []byte("some text"),
			wantErr:  ErrInvalidFileType,
		},
		{
			name:     "PDF with invalid content",
			filename: "resume.pdf",
			size:     10,
			content:  []byte("not a pdf"),
			wantErr:  ErrInvalidFileType,
		},
		{
			name:     "DOCX with invalid content",
			filename: "resume.docx",
			size:     10,
			content:  []byte("not a docx"),
			wantErr:  ErrInvalidFileType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResume(
				tt.filename,
				tt.size,
				bytes.NewReader(tt.content),
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"expected error %v, got %v",
					tt.wantErr,
					err,
				)
			}
		})
	}
}
