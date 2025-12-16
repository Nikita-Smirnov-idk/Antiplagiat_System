package text_extractor

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type TextExtractor struct {
	httpClient *http.Client
}

func NewTextExtractor() *TextExtractor {
	return &TextExtractor{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (e *TextExtractor) ExtractFromURL(fileURL string) (string, error) {
	resp, err := e.httpClient.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("failed to download file from URL %s: %w", fileURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file: HTTP status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	detectedType := http.DetectContentType(data)

	isTextFile := strings.Contains(contentType, "text/plain") ||
		strings.Contains(contentType, "text/") ||
		strings.Contains(detectedType, "text/plain") ||
		strings.Contains(detectedType, "text/")

	if !isTextFile {
		return "", fmt.Errorf("file is not a text file")
	}

	text := string(data)

	text = strings.Map(func(r rune) rune {
		if r >= 32 && r <= 126 || r == '\n' || r == '\t' || r == '\r' {
			return r
		}
		if r >= 0x400 && r <= 0x4FF {
			return r
		}
		return ' '
	}, text)

	return strings.Join(strings.Fields(text), " "), nil
}
