package ollama

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

// DefaultHTTPTimeout caps Ollama chat so unknown-intent UX degrades quickly.
const DefaultHTTPTimeout = 8 * time.Second

type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

func NewClient(baseURL, model string) *Client {
	return NewClientWithTimeout(baseURL, model, DefaultHTTPTimeout)
}

func NewClientWithTimeout(baseURL, model string, timeout time.Duration) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3.2"
	}
	if timeout <= 0 {
		timeout = DefaultHTTPTimeout
	}
	return &Client{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Format   string        `json:"format,omitempty"`
	Stream   bool          `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Message chatMessage `json:"message"`
}

func (c *Client) Chat(ctx context.Context, systemPrompt, userText string) (string, error) {
	return c.chat(ctx, systemPrompt, userText, "")
}

func (c *Client) ChatJSON(ctx context.Context, systemPrompt, userText string) (string, error) {
	return c.chat(ctx, systemPrompt, userText, "json")
}

func (c *Client) chat(ctx context.Context, systemPrompt, userText, format string) (string, error) {
	reqBody := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userText},
		},
		Stream: false,
	}
	if format != "" {
		reqBody.Format = format
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama chat: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("ollama status %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}

	var out chatResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("decode ollama response: %w", err)
	}
	content := strings.TrimSpace(out.Message.Content)
	if content == "" {
		return "", fmt.Errorf("empty ollama response")
	}
	return content, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
