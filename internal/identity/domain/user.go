package domain

import (
	"errors"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var (
	ErrEmptyDisplayName = errors.New("display name is required")
	ErrInvalidTimezone  = errors.New("timezone is required")
	ErrInvalidTelegram  = errors.New("telegram id must be positive")
)

type User struct {
	ID          ids.UserID
	TelegramID  int64
	DisplayName string
	Timezone    string
	CreatedAt   time.Time
}

func NewUser(telegramID int64, displayName, timezone string, now time.Time) (User, error) {
	if telegramID <= 0 {
		return User{}, ErrInvalidTelegram
	}
	if displayName == "" {
		return User{}, ErrEmptyDisplayName
	}
	if timezone == "" {
		return User{}, ErrInvalidTimezone
	}

	return User{
		ID:          ids.NewUserID(),
		TelegramID:  telegramID,
		DisplayName: displayName,
		Timezone:    timezone,
		CreatedAt:   now,
	}, nil
}
