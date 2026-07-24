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
	maxAttempts    = 3
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

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		text, err := c.doTranscribe(ctx, audio, filename, language)
		if err == nil {
			return text, nil
		}
		lastErr = err
		if !isRetryable(err) || attempt == maxAttempts {
			break
		}
		if err := sleepBackoff(ctx, attempt); err != nil {
			return "", err
		}
	}
	return "", lastErr
}

func (c *Client) doTranscribe(ctx context.Context, audio []byte, filename, language string) (string, error) {
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
		return "", &statusError{code: resp.StatusCode, body: strings.TrimSpace(string(raw))}
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

type statusError struct {
	code int
	body string
}

func (e *statusError) Error() string {
	return fmt.Sprintf("stt HTTP %d: %s", e.code, e.body)
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	if se, ok := err.(*statusError); ok {
		switch se.code {
		case http.StatusTooManyRequests, http.StatusInternalServerError,
			http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return true
		default:
			return false
		}
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "temporary") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "eof")
}

func sleepBackoff(ctx context.Context, attempt int) error {
	d := time.Duration(attempt) * 400 * time.Millisecond
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
