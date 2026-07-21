package stt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://api.groq.com/openai/v1"
	defaultModel   = "whisper-large-v3-turbo"
	defaultTimeout = 45 * time.Second
)

// Client calls OpenAI-compatible /audio/transcriptions (Groq Whisper by default).
type Client struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey, model string) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if model == "" {
		model = defaultModel
	}
	return &Client{
		baseURL: baseURL,
		apiKey:  strings.TrimSpace(apiKey),
		model:   model,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

type transcriptionResponse struct {
	Text string `json:"text"`
}

func (c *Client) Transcribe(ctx context.Context, audio []byte, filename, language string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("stt client is nil")
	}
	if len(audio) == 0 {
		return "", fmt.Errorf("empty audio")
	}
	if strings.TrimSpace(c.apiKey) == "" {
		return "", fmt.Errorf("stt api key is required")
	}
	filename = strings.TrimSpace(filename)
	if filename == "" {
		filename = "audio.ogg"
	}
	if path.Ext(filename) == "" {
		filename += ".ogg"
	}

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("model", c.model)
	if lang := strings.TrimSpace(language); lang != "" {
		_ = w.WriteField("language", lang)
	}
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	if _, err := part.Write(audio); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/audio/transcriptions", &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("stt HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var out transcriptionResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("stt decode: %w", err)
	}
	text := strings.TrimSpace(out.Text)
	if text == "" {
		return "", fmt.Errorf("stt returned empty text")
	}
	return text, nil
}
