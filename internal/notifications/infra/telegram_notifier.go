package infra

import (
	"context"
	"errors"
	"log/slog"

	tg "github.com/valentinezhov/lifeos/internal/transport/telegram"
)

// ErrEmptyChat means the user has no Telegram chat to deliver to.
var ErrEmptyChat = errors.New("telegram chat id is empty")

type TelegramNotifier struct {
	client *tg.Client
	log    *slog.Logger
}

func NewTelegramNotifier(client *tg.Client, log *slog.Logger) *TelegramNotifier {
	return &TelegramNotifier{client: client, log: log}
}

func (n *TelegramNotifier) Send(ctx context.Context, chatID int64, text string) error {
	if chatID == 0 {
		if n.log != nil {
			n.log.Warn("skip telegram notify: empty chat id")
		}
		return ErrEmptyChat
	}
	return n.client.SendMessage(ctx, chatID, text)
}

func (n *TelegramNotifier) SendWithKeyboard(ctx context.Context, chatID int64, text string, kb [][]tg.InlineButton) error {
	if chatID == 0 {
		if n.log != nil {
			n.log.Warn("skip telegram notify: empty chat id")
		}
		return ErrEmptyChat
	}
	return n.client.SendMessageWithKeyboard(ctx, chatID, text, kb)
}
