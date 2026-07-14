package telegram

import (
	"context"
	"log/slog"
	"time"
)

type Handler interface {
	HandleUpdate(ctx context.Context, update Update) error
}

type Poller struct {
	client  *Client
	handler Handler
	log     *slog.Logger
	offset  int64
}

func NewPoller(client *Client, handler Handler, log *slog.Logger) *Poller {
	return &Poller{client: client, handler: handler, log: log}
}

func (p *Poller) Run(ctx context.Context) error {
	p.log.Info("telegram poller started")
	for {
		select {
		case <-ctx.Done():
			p.log.Info("telegram poller stopped")
			return ctx.Err()
		default:
		}

		updates, err := p.client.GetUpdates(ctx, p.offset, 25)
		if err != nil {
			p.log.Error("get updates failed", "error", err)
			time.Sleep(2 * time.Second)
			continue
		}
		for _, u := range updates {
			p.offset = u.UpdateID + 1
			if err := p.handler.HandleUpdate(ctx, u); err != nil {
				p.log.Error("handle update failed", "error", err, "update_id", u.UpdateID)
			}
		}
	}
}
