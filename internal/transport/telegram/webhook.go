package telegram

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

// Webhook serves Telegram Bot API updates over HTTP.
type Webhook struct {
	handler Handler
	secret  string
	log     *slog.Logger
}

func NewWebhook(handler Handler, secret string, log *slog.Logger) *Webhook {
	return &Webhook{handler: handler, secret: secret, log: log}
}

func (w *Webhook) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(rw, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if w.secret != "" && r.Header.Get("X-Telegram-Bot-Api-Secret-Token") != w.secret {
		http.Error(rw, "unauthorized", http.StatusUnauthorized)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, "bad request", http.StatusBadRequest)
		return
	}
	var update Update
	if err := json.Unmarshal(body, &update); err != nil {
		http.Error(rw, "bad request", http.StatusBadRequest)
		return
	}
	if err := w.handler.HandleUpdate(r.Context(), update); err != nil {
		w.log.Error("handle webhook update failed", "error", err, "update_id", update.UpdateID)
	}
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte("ok"))
}

// RegisterWebhook configures Telegram to push updates to url.
func RegisterWebhook(ctx context.Context, client *Client, url, secret string) error {
	return client.SetWebhook(ctx, url, secret)
}

// ClearWebhook removes webhook and re-enables getUpdates polling at Telegram side.
func ClearWebhook(ctx context.Context, client *Client) error {
	return client.DeleteWebhook(ctx)
}
