package telegram

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
	OK          bool     `json:"ok"`
	Result      []Update `json:"result"`
	Description string   `json:"description"`
	ErrorCode   int      `json:"error_code"`
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
		desc := strings.TrimSpace(out.Description)
		if desc == "" {
			desc = truncate(string(body), 200)
		}
		return nil, fmt.Errorf("telegram getUpdates failed: %s", desc)
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
	_, err := c.SendScreen(ctx, chatID, text, nil, nil)
	return err
}

// SendScreen posts the main dashboard message.
// When replyKB is set, it is attached on send (chat-level persistent keyboard).
// Inline actions are applied in a follow-up edit — Telegram allows only one
// reply_markup type per send, and deleting a dedicated "⌨️" carrier used to
// wipe the reply keyboard for the whole chat.
func (c *Client) SendScreen(ctx context.Context, chatID int64, text string, inline [][]InlineButton, replyKB [][]ReplyButton) (int64, error) {
	if inline == nil {
		inline = [][]InlineButton{}
	}
	if len(replyKB) > 0 {
		id, err := c.postMessage(ctx, sendMessageRequest{
			ChatID:      chatID,
			Text:        text,
			ParseMode:   "HTML",
			ReplyMarkup: replyKeyboardMarkup(replyKB),
		})
		if err != nil {
			return 0, err
		}
		if err := c.EditScreen(ctx, chatID, id, text, inline); err != nil {
			// Persistent reply keyboard is already installed; inline is best-effort.
			return id, nil
		}
		return id, nil
	}
	return c.postMessage(ctx, sendMessageRequest{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   "HTML",
		ReplyMarkup: inlineKeyboardMarkup(inline),
	})
}

// EditScreen updates the existing dashboard message instead of sending a new one.
func (c *Client) EditScreen(ctx context.Context, chatID, messageID int64, text string, inline [][]InlineButton) error {
	if inline == nil {
		inline = [][]InlineButton{}
	}
	payload := map[string]any{
		"chat_id":      chatID,
		"message_id":   messageID,
		"text":         text,
		"parse_mode":   "HTML",
		"reply_markup": inlineKeyboardMarkup(inline),
	}
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

func (c *Client) SetReplyKeyboard(ctx context.Context, chatID int64, keyboard [][]ReplyButton) error {
	// Prefer attaching the reply keyboard to the dashboard via SendScreen.
	// This helper remains for rare force-reinstall without a screen body.
	_, err := c.SendScreen(ctx, chatID, "Клавиатура обновлена.", nil, keyboard)
	return err
}

func replyKeyboardMarkup(rows [][]ReplyButton) map[string]any {
	kb := make([][]map[string]any, 0, len(rows))
	for _, row := range rows {
		buttons := make([]map[string]any, 0, len(row))
		for _, btn := range row {
			item := map[string]any{"text": btn.Text}
			if btn.WebApp != "" {
				item["web_app"] = map[string]string{"url": btn.WebApp}
			}
			buttons = append(buttons, item)
		}
		kb = append(kb, buttons)
	}
	return map[string]any{
		"keyboard":                kb,
		"resize_keyboard":         true,
		"is_persistent":           true,
		"one_time_keyboard":       false,
		"input_field_placeholder": "Раздел, команда или текст…",
	}
}

// SetChatMenuButton installs the global Telegram menu button that opens Mini App.
func (c *Client) SetChatMenuButton(ctx context.Context, text, webAppURL string) error {
	if webAppURL == "" {
		return nil
	}
	if text == "" {
		text = "Mini App"
	}
	return c.postAPI(ctx, "setChatMenuButton", map[string]any{
		"menu_button": map[string]any{
			"type":    "web_app",
			"text":    text,
			"web_app": map[string]string{"url": webAppURL},
		},
	})
}

// DeleteMessage removes a chat message. Best-effort: callers may ignore errors
// (e.g. message already gone / too old).
func (c *Client) DeleteMessage(ctx context.Context, chatID, messageID int64) error {
	if messageID <= 0 {
		return nil
	}
	return c.postAPI(ctx, "deleteMessage", map[string]any{
		"chat_id":    chatID,
		"message_id": messageID,
	})
}

type chatResponse struct {
	OK     bool `json:"ok"`
	Result struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Type     string `json:"type"`
	} `json:"result"`
	Description string `json:"description"`
}

// ResolveUsername resolves a public Telegram @username to a numeric id via getChat.
func (c *Client) ResolveUsername(ctx context.Context, username string) (int64, error) {
	username = strings.TrimSpace(strings.TrimPrefix(username, "@"))
	if username == "" {
		return 0, fmt.Errorf("empty username")
	}
	payload := map[string]any{"chat_id": "@" + username}
	data, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/getChat", bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out chatResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return 0, err
	}
	if !out.OK || out.Result.ID == 0 {
		if out.Description != "" {
			return 0, fmt.Errorf("telegram getChat: %s", out.Description)
		}
		return 0, fmt.Errorf("telegram getChat failed for @%s", username)
	}
	return out.Result.ID, nil
}

func (c *Client) SendMessageWithKeyboard(ctx context.Context, chatID int64, text string, keyboard [][]InlineButton) error {
	_, err := c.SendScreen(ctx, chatID, text, keyboard, nil)
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

// InlineButton is a Telegram inline keyboard button (callback or web_app).
// Prefer web_app on inline (not reply) keyboards — reply web_app often yields empty initData.
type InlineButton struct {
	Text         string
	CallbackData string
	WebApp       string // when set, Telegram opens this HTTPS Mini App URL with initData
}

func inlineKeyboardMarkup(rows [][]InlineButton) map[string]any {
	kb := make([][]map[string]any, 0, len(rows))
	for _, row := range rows {
		buttons := make([]map[string]any, 0, len(row))
		for _, btn := range row {
			item := map[string]any{"text": btn.Text}
			switch {
			case strings.TrimSpace(btn.WebApp) != "":
				item["web_app"] = map[string]string{"url": strings.TrimSpace(btn.WebApp)}
			case btn.CallbackData != "":
				item["callback_data"] = btn.CallbackData
			}
			buttons = append(buttons, item)
		}
		kb = append(kb, buttons)
	}
	return map[string]any{"inline_keyboard": kb}
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
