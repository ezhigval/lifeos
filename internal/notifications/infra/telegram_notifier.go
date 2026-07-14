package infra

import (
	"context"
	"log/slog"

	tg "github.com/valentinezhov/lifeos/internal/transport/telegram"
)

type TelegramNotifier struct {
	client *tg.Client
	log    *slog.Logger
}

func NewTelegramNotifier(client *tg.Client, log *slog.Logger) *TelegramNotifier {
	return &TelegramNotifier{client: client, log: log}
}

func (n *TelegramNotifier) Send(ctx context.Context, chatID int64, text string) error {
	return n.client.SendMessage(ctx, chatID, text)
}

func (n *TelegramNotifier) SendWithKeyboard(ctx context.Context, chatID int64, text string, kb [][]tg.InlineButton) error {
	return n.client.SendMessageWithKeyboard(ctx, chatID, text, kb)
}
