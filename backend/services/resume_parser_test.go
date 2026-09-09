package services

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestExtractTxtText(t *testing.T) {
	data := []byte("John Doe\nSenior Go Developer\nSkills: Go, PostgreSQL\n")
	text, failure := extractTxtText(data)
	if failure != "" {
		t.Fatalf("unexpected failure: %s", failure)
	}
	if !strings.Contains(text, "John Doe") || !strings.Contains(text, "PostgreSQL") {
		t.Fatalf("extracted text missing expected content: %q", text)
	}
}

func TestExtractDocTextUnsupported(t *testing.T) {
	_, failure := extractDocText([]byte("not really a doc"))
	if failure == "" {
		t.Fatal("expected legacy .doc extraction to report unsupported")
	}
}

func buildTestDocx(t *testing.T, paragraphs []string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatalf("create zip entry: %v", err)
	}
	var xmlBody strings.Builder
	for _, p := range paragraphs {
		xmlBody.WriteString("<w:p><w:r><w:t>")
		xmlBody.WriteString(p)
		xmlBody.WriteString("</w:t></w:r></w:p>")
	}
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:body>` + xmlBody.String() + `</w:body></w:document>`
	if _, err := w.Write([]byte(content)); err != nil {
		t.Fatalf("write xml: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func TestExtractDocxText(t *testing.T) {
	docx := buildTestDocx(t, []string{"Software Engineer", "AWS and Go"})
	text, failure := extractDocxText(docx)
	if failure != "" {
		t.Fatalf("unexpected failure: %s", failure)
	}
	if !strings.Contains(text, "Software Engineer") || !strings.Contains(text, "AWS and Go") {
		t.Fatalf("extracted docx text missing content: %q", text)
	}
}

func TestExtractDocxTextMissingDocumentXML(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("random.txt")
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	_, _ = w.Write([]byte("hi"))
	_ = zw.Close()

	if _, failure := extractDocxText(buf.Bytes()); failure == "" {
		t.Fatal("expected failure when word/document.xml is absent")
	}
}

// buildMinimalPDF assembles a tiny but structurally valid single-page PDF
// whose text content we control, computing xref offsets so a strict parser
// can read it.
func buildMinimalPDF(text string) []byte {
	object := func(n int, body []byte) []byte {
		out := []byte(fmt.Sprintf("%d 0 obj\n", n))
		out = append(out, body...)
		return append(out, []byte("\nendobj\n")...)
	}

	var objects [][]byte
	objects = append(objects, object(1, []byte("<< /Type /Catalog /Pages 2 0 R >>")))
	objects = append(objects, object(2, []byte("<< /Type /Pages /Kids [3 0 R] /Count 1 >>")))
	objects = append(objects, object(3, []byte("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>")))

	stream := []byte("BT /F1 12 Tf 72 720 Td (")
	stream = append(stream, text...)
	stream = append(stream, []byte(") Tj ET")...)
	obj4 := []byte("<< /Length " + strconv.Itoa(len(stream)) + " >>\nstream\n")
	obj4 = append(obj4, stream...)
	obj4 = append(obj4, []byte("\nendstream")...)
	objects = append(objects, object(4, obj4))

	objects = append(objects, object(5, []byte("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")))

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects))
	for i, obj := range objects {
		offsets[i] = buf.Len()
		buf.Write(obj)
	}
	xrefPos := buf.Len()
	buf.WriteString("xref\n0 " + strconv.Itoa(len(objects)+1) + "\n")
	buf.WriteString("0000000000 65535 f \n")
	for _, off := range offsets {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", off))
	}
	buf.WriteString("trailer\n<< /Size " + strconv.Itoa(len(objects)+1) + " /Root 1 0 R >>\n")
	buf.WriteString("startxref\n" + strconv.Itoa(xrefPos) + "\n%%EOF")

	return buf.Bytes()
}

func TestExtractPDFText(t *testing.T) {
	pdfData := buildMinimalPDF("Hello PDF Parser")
	text, failure := extractPDFText(pdfData)
	if failure != "" {
		t.Fatalf("unexpected failure: %s", failure)
	}
	if !strings.Contains(text, "Hello PDF Parser") {
		t.Fatalf("extracted pdf text missing content: %q", text)
	}
}
