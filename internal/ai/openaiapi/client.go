package openaiapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultHTTPTimeout caps chat so unknown-intent UX degrades quickly (same as ollama).
const DefaultHTTPTimeout = 8 * time.Second

const (
	defaultBaseURL = "https://api.groq.com/openai/v1"
	defaultModel   = "llama-3.3-70b-versatile"
)

type Client struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey, model string) *Client {
	return NewClientWithTimeout(baseURL, apiKey, model, DefaultHTTPTimeout)
}

func NewClientWithTimeout(baseURL, apiKey, model string, timeout time.Duration) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if model == "" {
		model = defaultModel
	}
	if timeout <= 0 {
		timeout = DefaultHTTPTimeout
	}
	return &Client{
		baseURL: baseURL,
		apiKey:  strings.TrimSpace(apiKey),
		model:   model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	Temperature    float64         `json:"temperature"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func (c *Client) Chat(ctx context.Context, systemPrompt, userText string) (string, error) {
	return c.chat(ctx, systemPrompt, userText, false)
}

// ChatJSON asks for JSON via the system prompt and, when supported, response_format.
// Providers that reject response_format (HTTP 400) are retried without it; the prompt alone still requests JSON.
func (c *Client) ChatJSON(ctx context.Context, systemPrompt, userText string) (string, error) {
	content, err := c.chat(ctx, systemPrompt, userText, true)
	if err != nil && isBadRequest(err) {
		return c.chat(ctx, systemPrompt, userText, false)
	}
	return content, err
}

type statusError struct {
	code int
	body string
}

func (e *statusError) Error() string {
	return fmt.Sprintf("openaiapi status %d: %s", e.code, truncate(e.body, 200))
}

func isBadRequest(err error) bool {
	se, ok := err.(*statusError)
	return ok && se.code == http.StatusBadRequest
}

func (c *Client) chat(ctx context.Context, systemPrompt, userText string, wantJSONObject bool) (string, error) {
	reqBody := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userText},
		},
		Temperature: 0,
	}
	if wantJSONObject {
		reqBody.ResponseFormat = &responseFormat{Type: "json_object"}
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openaiapi chat: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", &statusError{code: resp.StatusCode, body: string(raw)}
	}

	var out chatResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("decode openaiapi response: %w", err)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("empty openaiapi choices")
	}
	content := strings.TrimSpace(out.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("empty openaiapi response")
	}
	return content, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
