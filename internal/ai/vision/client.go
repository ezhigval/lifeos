package vision

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://api.groq.com/openai/v1"
	defaultModel   = "meta-llama/llama-4-scout-17b-16e-instruct"
	defaultTimeout = 45 * time.Second
	maxAttempts    = 3
)

// lifeOSPrompt converts a photo into a short Russian phrase the LifeOS agent can treat as user text.
const lifeOSPrompt = `Ты превращаешь фото в короткую команду пользователя для LifeOS — личного ассистента (задачи, финансы, здоровье, заметки, проекты).
Ответь ОДНОЙ короткой фразой на русском — как будто пользователь сам написал текстом, что сделать с этим фото.
Правила:
- чек / счёт / ценник → сформулируй расход («запиши расход 450 еда» или «запиши расход по чеку …»)
- список / доска / стикеры → задачи («добавь задачи: …»)
- рукопись / текст → заметка или задачи по смыслу
- если непонятно, что делать → «что на фото: …» (коротко опиши суть)
Без кавычек, без markdown, без пояснений и преамбул.`

// Client calls OpenAI-compatible multimodal chat completions (Groq Llama 4 Scout by default).
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

type contentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *imageURL `json:"image_url,omitempty"`
}

type imageURL struct {
	URL string `json:"url"`
}

type chatMessage struct {
	Role    string        `json:"role"`
	Content []contentPart `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *Client) ImageToUserText(ctx context.Context, image []byte, mimeType string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("vision client is nil")
	}
	if len(image) == 0 {
		return "", fmt.Errorf("empty image")
	}
	if strings.TrimSpace(c.apiKey) == "" {
		return "", fmt.Errorf("vision api key is required")
	}
	mimeType = normalizeMIME(mimeType)
	dataURL := "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(image)

	body, err := json.Marshal(chatRequest{
		Model: c.model,
		Messages: []chatMessage{{
			Role: "user",
			Content: []contentPart{
				{Type: "text", Text: lifeOSPrompt},
				{Type: "image_url", ImageURL: &imageURL{URL: dataURL}},
			},
		}},
		Temperature: 0.2,
		MaxTokens:   256,
	})
	if err != nil {
		return "", err
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		text, err := c.doChat(ctx, body)
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

func (c *Client) doChat(ctx context.Context, body []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
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
	var out chatResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("vision decode: %w", err)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("vision returned no choices")
	}
	text := strings.TrimSpace(out.Choices[0].Message.Content)
	if text == "" {
		return "", fmt.Errorf("vision returned empty text")
	}
	return text, nil
}

type statusError struct {
	code int
	body string
}

func (e *statusError) Error() string {
	return fmt.Sprintf("vision HTTP %d: %s", e.code, truncate(e.body, 200))
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
	// network / timeout
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

func normalizeMIME(mimeType string) string {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	switch {
	case mimeType == "" || mimeType == "application/octet-stream":
		return "image/jpeg"
	case strings.HasPrefix(mimeType, "image/"):
		return mimeType
	default:
		return "image/jpeg"
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
