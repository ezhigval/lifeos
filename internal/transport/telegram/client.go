package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	token string
	http  *http.Client
	base  string
}

func NewClient(token string) *Client {
	return &Client{
		token: token,
		base:  "https://api.telegram.org/bot" + token,
		http:  &http.Client{Timeout: 15 * time.Second},
	}
}

type Update struct {
	UpdateID      int64          `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type Message struct {
	MessageID int64  `json:"message_id"`
	Text      string `json:"text"`
	Chat      Chat   `json:"chat"`
	From      User   `json:"from"`
}

type CallbackQuery struct {
	ID      string  `json:"id"`
	Data    string  `json:"data"`
	Message Message `json:"message"`
	From    User    `json:"from"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

func FormatDisplayName(u User) string {
	switch {
	case u.FirstName != "" && u.LastName != "":
		return u.FirstName + " " + u.LastName
	case u.FirstName != "":
		return u.FirstName
	case u.Username != "":
		return u.Username
	default:
		return fmt.Sprintf("User %d", u.ID)
	}
}

type UpdatesResponse struct {
	OK     bool     `json:"ok"`
	Result []Update `json:"result"`
}

func (c *Client) GetUpdates(ctx context.Context, offset int64, timeout int) ([]Update, error) {
	url := fmt.Sprintf("%s/getUpdates?offset=%d&timeout=%d", c.base, offset, timeout)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	// Long polling: Telegram holds the connection up to `timeout` seconds.
	pollClient := &http.Client{Timeout: time.Duration(timeout+10) * time.Second}
	resp, err := pollClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out UpdatesResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	if !out.OK {
		return nil, fmt.Errorf("telegram getUpdates failed")
	}
	return out.Result, nil
}

type sendMessageRequest struct {
	ChatID      int64  `json:"chat_id"`
	Text        string `json:"text"`
	ParseMode   string `json:"parse_mode,omitempty"`
	ReplyMarkup any    `json:"reply_markup,omitempty"`
}

type messageResponse struct {
	OK     bool `json:"ok"`
	Result struct {
		MessageID int64 `json:"message_id"`
	} `json:"result"`
	Description string `json:"description"`
}

func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	_, err := c.SendScreen(ctx, chatID, text, nil)
	return err
}

// SendScreen posts the main dashboard message with inline actions.
func (c *Client) SendScreen(ctx context.Context, chatID int64, text string, inline [][]InlineButton) (int64, error) {
	payload := sendMessageRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "HTML",
	}
	payload.ReplyMarkup = map[string]any{"inline_keyboard": inline}
	return c.postMessage(ctx, payload)
}

// EditScreen updates the existing dashboard message instead of sending a new one.
func (c *Client) EditScreen(ctx context.Context, chatID, messageID int64, text string, inline [][]InlineButton) error {
	payload := map[string]any{
		"chat_id":    chatID,
		"message_id": messageID,
		"text":       text,
		"parse_mode": "HTML",
	}
	payload["reply_markup"] = map[string]any{"inline_keyboard": inline}
	return withRetry(ctx, func() error {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/editMessageText", bytes.NewReader(data))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode >= 300 {
			return fmt.Errorf("telegram editMessageText: %s", string(body))
		}
		var out messageResponse
		if err := json.Unmarshal(body, &out); err != nil {
			return err
		}
		if !out.OK {
			if isBenignEditError(out.Description) {
				return nil
			}
			return fmt.Errorf("telegram editMessageText: %s", out.Description)
		}
		return nil
	})
}

func (c *Client) SetReplyKeyboard(ctx context.Context, chatID int64, keyboard [][]string) error {
	payload := sendMessageRequest{
		ChatID:    chatID,
		Text:      "Разделы — кнопками ниже",
		ParseMode: "HTML",
		ReplyMarkup: map[string]any{
			"keyboard":          keyboard,
			"resize_keyboard":   true,
			"is_persistent":     true,
			"one_time_keyboard": false,
		},
	}
	_, err := c.postMessage(ctx, payload)
	return err
}

func (c *Client) SendMessageWithKeyboard(ctx context.Context, chatID int64, text string, keyboard [][]InlineButton) error {
	_, err := c.SendScreen(ctx, chatID, text, keyboard)
	return err
}

func (c *Client) send(ctx context.Context, payload sendMessageRequest) error {
	_, err := c.postMessage(ctx, payload)
	return err
}

func (c *Client) postMessage(ctx context.Context, payload sendMessageRequest) (int64, error) {
	var id int64
	err := withRetry(ctx, func() error {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/sendMessage", bytes.NewReader(data))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode >= 300 {
			return fmt.Errorf("telegram sendMessage: %s", string(body))
		}
		var out messageResponse
		if err := json.Unmarshal(body, &out); err != nil {
			return err
		}
		if !out.OK {
			return fmt.Errorf("telegram sendMessage: %s", out.Description)
		}
		id = out.Result.MessageID
		return nil
	})
	return id, err
}

type InlineButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

func (c *Client) AnswerCallback(ctx context.Context, callbackID string) error {
	payload := map[string]string{"callback_query_id": callbackID}
	return withRetry(ctx, func() error {
		data, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/answerCallbackQuery", bytes.NewReader(data))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return nil
	})
}

type apiResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

func (c *Client) SetWebhook(ctx context.Context, webhookURL, secretToken string) error {
	payload := map[string]any{
		"url":                  webhookURL,
		"secret_token":         secretToken,
		"allowed_updates":      []string{"message", "callback_query"},
		"drop_pending_updates": true,
	}
	return c.postAPI(ctx, "setWebhook", payload)
}

func (c *Client) DeleteWebhook(ctx context.Context) error {
	return c.postAPI(ctx, "deleteWebhook", map[string]any{"drop_pending_updates": true})
}

func (c *Client) postAPI(ctx context.Context, method string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/"+method, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out apiResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return err
	}
	if !out.OK {
		return fmt.Errorf("telegram %s: %s", method, out.Description)
	}
	return nil
}
