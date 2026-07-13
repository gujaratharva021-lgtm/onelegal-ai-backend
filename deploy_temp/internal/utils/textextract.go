package utils

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
)

// ExtractText pulls plain text out of a saved upload so it can be handed to
// the AI as context. Supports PDF, DOCX, and plain text; legacy .doc is
// intentionally unsupported since it isn't a simple zip/text format.
func ExtractText(filePath, ext string) (string, error) {
	switch strings.ToLower(ext) {
	case ".pdf":
		return extractPDFText(filePath)
	case ".docx":
		return extractDocxText(filePath)
	case ".txt":
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		return string(data), nil
	default:
		return "", errors.New("unsupported file type: please upload a PDF, DOCX, or TXT file")
	}
}

func extractPDFText(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	totalPage := r.NumPage()
	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}

	if buf.Len() == 0 {
		return "", errors.New("could not extract any text from the PDF")
	}
	return buf.String(), nil
}

type docxDocument struct {
	XMLName xml.Name `xml:"document"`
	Body    docxBody `xml:"body"`
}

type docxBody struct {
	Paragraphs []docxParagraph `xml:"p"`
}

type docxParagraph struct {
	Runs []docxRun `xml:"r"`
}

type docxRun struct {
	Text string `xml:"t"`
}

func extractDocxText(filePath string) (string, error) {
	zr, err := zip.OpenReader(filePath)
	if err != nil {
		return "", err
	}
	defer zr.Close()

	var docXML io.ReadCloser
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			docXML, err = f.Open()
			if err != nil {
				return "", err
			}
			break
		}
	}
	if docXML == nil {
		return "", errors.New("not a valid DOCX file")
	}
	defer docXML.Close()

	data, err := io.ReadAll(docXML)
	if err != nil {
		return "", err
	}

	var doc docxDocument
	if err := xml.Unmarshal(data, &doc); err != nil {
		return "", err
	}

	var buf bytes.Buffer
	for _, p := range doc.Body.Paragraphs {
		for _, r := range p.Runs {
			buf.WriteString(r.Text)
		}
		buf.WriteString("\n")
	}

	if buf.Len() == 0 {
		return "", errors.New("could not extract any text from the DOCX")
	}
	return buf.String(), nil
}
