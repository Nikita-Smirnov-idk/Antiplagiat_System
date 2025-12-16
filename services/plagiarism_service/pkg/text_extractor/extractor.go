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

// ExtractFromURL скачивает файл по URL и извлекает текст
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
		return "", fmt.Errorf("can not extract")
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

// CleanText очищает текст для сравнения
func (e *TextExtractor) CleanText(text string) string {
	if text == "" {
		return ""
	}

	text = strings.ToLower(text)

	text = strings.Map(func(r rune) rune {
		switch {
		case (r >= 'а' && r <= 'я') || r == 'ё':
			return r
		case (r >= 'a' && r <= 'z'):
			return r
		case (r >= '0' && r <= '9'):
			return r
		case r == ' ' || r == '\n' || r == '\t':
			return ' '
		default:
			return ' '
		}
	}, text)

	words := strings.Fields(text)

	stopWords := map[string]bool{
		"и": true, "в": true, "не": true, "на": true, "с": true,
		"по": true, "к": true, "у": true, "о": true, "за": true,
		"из": true, "от": true, "до": true, "для": true, "это": true,
		"как": true, "так": true, "но": true, "а": true, "же": true,
		"что": true, "он": true, "она": true, "они": true, "мы": true,
		"вы": true, "его": true, "ее": true, "их": true, "все": true,
		"то": true, "бы": true, "во": true,
	}

	filtered := make([]string, 0, len(words))
	for _, word := range words {
		if !stopWords[word] && len(word) > 2 {
			filtered = append(filtered, word)
		}
	}

	return strings.Join(filtered, " ")
}
