package wordcloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	QuickChartAPIURL = "https://quickchart.io/wordcloud"
	DefaultFormat    = "png"
	DefaultWidth     = 1000
	DefaultHeight    = 1000
	DefaultFontScale = 15
)

type WordCloudRequest struct {
	Text            string   `json:"text"`
	Format          string   `json:"format,omitempty"`          // svg или png
	Width           int      `json:"width,omitempty"`           // ширина изображения
	Height          int      `json:"height,omitempty"`          // высота изображения
	BackgroundColor string   `json:"backgroundColor,omitempty"` // цвет фона
	FontFamily      string   `json:"fontFamily,omitempty"`      // семейство шрифтов
	FontScale       int      `json:"fontScale,omitempty"`       // размер шрифта
	Scale           string   `json:"scale,omitempty"`           // linear, sqrt, или log
	MaxNumWords     int      `json:"maxNumWords,omitempty"`     // максимальное количество слов
	MinWordLength   int      `json:"minWordLength,omitempty"`   // минимальная длина слова
	RemoveStopwords bool     `json:"removeStopwords,omitempty"` // удалять стоп-слова
	Language        string   `json:"language,omitempty"`        // код языка для стоп-слов
	Colors          []string `json:"colors,omitempty"`          // цвета для слов
}

type QuickChartClient struct {
	httpClient *http.Client
	apiURL     string
}

func NewQuickChartClient() *QuickChartClient {
	return &QuickChartClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiURL: QuickChartAPIURL,
	}
}

func (c *QuickChartClient) GenerateWordCloud(req WordCloudRequest) ([]byte, string, error) {
	if req.Format == "" {
		req.Format = DefaultFormat
	}
	if req.Width == 0 {
		req.Width = DefaultWidth
	}
	if req.Height == 0 {
		req.Height = DefaultHeight
	}
	if req.FontScale == 0 {
		req.FontScale = DefaultFontScale
	}
	if req.Scale == "" {
		req.Scale = "linear"
	}
	if req.MaxNumWords == 0 {
		req.MaxNumWords = 200
	}
	if req.MinWordLength == 0 {
		req.MinWordLength = 3
	}
	if req.Language == "" {
		req.Language = "ru"
	}
	if req.RemoveStopwords {
		req.RemoveStopwords = true
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("quickchart API returned status %d: %s", resp.StatusCode, string(body))
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		if req.Format == "svg" {
			contentType = "image/svg+xml"
		} else {
			contentType = "image/png"
		}
	}

	return imageData, contentType, nil
}

func (c *QuickChartClient) GenerateWordCloudURL(text string, options map[string]string) string {
	cloudURL := fmt.Sprintf("%s?text=%s", c.apiURL, url.QueryEscape(text))

	for key, value := range options {
		cloudURL += fmt.Sprintf("&%s=%s", key, url.QueryEscape(value))
	}

	return cloudURL
}
