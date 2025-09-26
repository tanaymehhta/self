package services

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/net/html"

	"github.com/nguyenthenguyen/docx"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

type TextExtractorService struct{}

func NewTextExtractorService() *TextExtractorService {
	return &TextExtractorService{}
}

// ExtractText extracts text from various file formats
func (t *TextExtractorService) ExtractText(content []byte, filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".txt":
		return string(content), nil
	case ".pdf":
		return t.extractPDFText(content)
	case ".epub":
		return t.extractEPUBText(content)
	case ".docx":
		return t.extractDOCXText(content)
	case ".html", ".htm":
		return t.extractHTMLText(content), nil
	default:
		// Try to treat as plain text
		return string(content), nil
	}
}

// extractPDFText extracts text from PDF files using unidoc
func (t *TextExtractorService) extractPDFText(content []byte) (string, error) {
	reader := bytes.NewReader(content)

	// Create PDF reader
	pdfReader, err := model.NewPdfReader(reader)
	if err != nil {
		return "", fmt.Errorf("failed to create PDF reader: %w", err)
	}

	// Check for password protection
	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return "", fmt.Errorf("failed to check PDF encryption: %w", err)
	}

	if isEncrypted {
		// Try empty password first
		auth, err := pdfReader.Decrypt([]byte(""))
		if err != nil || !auth {
			return "", fmt.Errorf("PDF is password protected and cannot be decrypted")
		}
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return "", fmt.Errorf("failed to get number of pages: %w", err)
	}

	var textBuilder strings.Builder

	// Extract text from each page
	for pageNum := 1; pageNum <= numPages; pageNum++ {
		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			continue // Skip problematic pages
		}

		textExtractor, err := extractor.New(page)
		if err != nil {
			continue // Skip problematic pages
		}

		pageText, err := textExtractor.ExtractText()
		if err != nil {
			continue // Skip problematic pages
		}

		textBuilder.WriteString(pageText)
		textBuilder.WriteString("\n\n")
	}

	return strings.TrimSpace(textBuilder.String()), nil
}

// extractEPUBText extracts text from EPUB files
func (t *TextExtractorService) extractEPUBText(content []byte) (string, error) {
	reader := bytes.NewReader(content)

	// EPUB files are ZIP archives
	zipReader, err := zip.NewReader(reader, int64(len(content)))
	if err != nil {
		return "", fmt.Errorf("failed to open EPUB as ZIP: %w", err)
	}

	var textBuilder strings.Builder

	// Look for content files (usually .xhtml or .html files)
	for _, file := range zipReader.File {
		if strings.HasSuffix(strings.ToLower(file.Name), ".xhtml") ||
			strings.HasSuffix(strings.ToLower(file.Name), ".html") {

			rc, err := file.Open()
			if err != nil {
				continue // Skip problematic files
			}

			fileContent, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue // Skip problematic files
			}

			// Extract text from HTML/XHTML
			plainText := t.extractHTMLText(fileContent)
			textBuilder.WriteString(plainText)
			textBuilder.WriteString("\n\n")
		}
	}

	result := strings.TrimSpace(textBuilder.String())
	if result == "" {
		return "", fmt.Errorf("no readable text content found in EPUB")
	}

	return result, nil
}

// extractDOCXText extracts text from DOCX files
func (t *TextExtractorService) extractDOCXText(content []byte) (string, error) {
	reader := bytes.NewReader(content)

	// Read DOCX file
	r, err := docx.ReadDocxFromMemory(reader, int64(len(content)))
	if err != nil {
		return "", fmt.Errorf("failed to read DOCX file: %w", err)
	}

	// Use the built-in text extraction method
	text := r.Editable().GetContent()

	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("no text content found in DOCX document")
	}

	return strings.TrimSpace(text), nil
}

// extractHTMLText strips HTML tags and extracts plain text
func (t *TextExtractorService) extractHTMLText(content []byte) string {
	reader := strings.NewReader(string(content))
	tokenizer := html.NewTokenizer(reader)

	var textBuilder strings.Builder
	var skipContent bool

	for {
		tokenType := tokenizer.Next()

		switch tokenType {
		case html.ErrorToken:
			// End of document
			goto cleanup

		case html.StartTagToken:
			token := tokenizer.Token()
			tagName := strings.ToLower(token.Data)

			// Skip script and style content
			if tagName == "script" || tagName == "style" {
				skipContent = true
			}

			// Add spacing for block elements
			if t.isBlockElement(tagName) {
				textBuilder.WriteString("\n")
			}

		case html.EndTagToken:
			token := tokenizer.Token()
			tagName := strings.ToLower(token.Data)

			// Resume content after script/style
			if tagName == "script" || tagName == "style" {
				skipContent = false
			}

			// Add spacing for block elements
			if t.isBlockElement(tagName) {
				textBuilder.WriteString("\n")
			}

		case html.TextToken:
			if !skipContent {
				text := strings.TrimSpace(tokenizer.Token().Data)
				if text != "" {
					textBuilder.WriteString(text)
					textBuilder.WriteString(" ")
				}
			}
		}
	}

cleanup:
	// Clean up the extracted text
	text := textBuilder.String()

	// Remove extra whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`\n\s*\n`).ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}

// isBlockElement checks if an HTML tag is a block-level element
func (t *TextExtractorService) isBlockElement(tagName string) bool {
	blockElements := map[string]bool{
		"div": true, "p": true, "h1": true, "h2": true, "h3": true,
		"h4": true, "h5": true, "h6": true, "br": true, "hr": true,
		"blockquote": true, "pre": true, "ul": true, "ol": true,
		"li": true, "table": true, "tr": true, "td": true, "th": true,
		"section": true, "article": true, "header": true, "footer": true,
		"main": true, "aside": true, "nav": true,
	}

	return blockElements[tagName]
}