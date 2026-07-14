package telegram

import (
	"context"
	"strings"
	"time"
)

const maxAPIAttempts = 3

func isBenignEditError(desc string) bool {
	return strings.Contains(desc, "message is not modified")
}

func withRetry(ctx context.Context, fn func() error) error {
	var last error
	backoff := 200 * time.Millisecond
	for attempt := 0; attempt < maxAPIAttempts; attempt++ {
		if err := fn(); err == nil {
			return nil
		} else {
			last = err
			if attempt+1 >= maxAPIAttempts {
				break
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				backoff *= 2
			}
		}
	}
	return last
}
