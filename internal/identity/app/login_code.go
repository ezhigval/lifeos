package app

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/valentinezhov/lifeos/internal/identity/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

const (
	loginCodeTTL         = 5 * time.Minute
	loginCodeMinGap      = time.Minute
	loginCodeHourlyMax   = 5
	loginCodeMaxAttempts = 5
)

var (
	ErrLoginBadUsername = errors.New("bad telegram username")
	ErrLoginUnknownUser = errors.New("telegram user is not known to the bot")
	ErrLoginRateLimited = errors.New("login code rate limited")
	ErrLoginDelivery    = errors.New("login code delivery failed")
	ErrLoginBadCode     = errors.New("login code rejected")
)

// LoginCode is one issued challenge. CodeHash is HMAC-SHA256, never the digits.
type LoginCode struct {
	ID         uuid.UUID
	TelegramID int64
	Username   string
	CodeHash   string
	ExpiresAt  time.Time
	Attempts   int
	CreatedAt  time.Time
}

// LoginCodeStore persists challenges. Implementations must not store the raw code.
type LoginCodeStore interface {
	CountSince(ctx context.Context, username string, since time.Time) (int, error)
	LatestActive(ctx context.Context, username string, now time.Time) (LoginCode, error)
	Insert(ctx context.Context, row LoginCode) error
	AddAttempt(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteOlderThan(ctx context.Context, before time.Time) error
}

// LoginUserDirectory finds a user by the nick saved when they talked to the bot.
type LoginUserDirectory interface {
	GetByUsername(ctx context.Context, username string) (domain.User, error)
	RememberUsername(ctx context.Context, userID ids.UserID, username string) error
}

// LoginCodeSender delivers the digits. The text contains the code; do not log it.
type LoginCodeSender interface {
	Send(ctx context.Context, telegramID int64, text string) error
}

// LoginUsernameResolver asks Telegram for a numeric id. Optional.
// getChat only works for people who already opened the bot.
type LoginUsernameResolver interface {
	Resolve(ctx context.Context, username string) (int64, error)
}

// TelegramLogin issues a one-time code in the bot chat and exchanges it for the
// same user row Mini App auth uses. Later providers (Yandex, account linking)
// should resolve to this user id and then call TokenService.Issue — they do
// not get a second session type.
type TelegramLogin struct {
	users    LoginUserDirectory
	codes    LoginCodeStore
	ensure   *EnsureUserByTelegram
	sender   LoginCodeSender
	resolver LoginUsernameResolver
	secret   []byte
	now      func() time.Time
}

func NewTelegramLogin(
	users LoginUserDirectory,
	codes LoginCodeStore,
	ensure *EnsureUserByTelegram,
	sender LoginCodeSender,
	resolver LoginUsernameResolver,
	secret string,
) *TelegramLogin {
	return &TelegramLogin{
		users:    users,
		codes:    codes,
		ensure:   ensure,
		sender:   sender,
		resolver: resolver,
		secret:   []byte(secret),
		now:      time.Now,
	}
}

// Request sends a fresh code. The previous code for this nick stops being the
// one LatestActive returns only after it expires or runs out of attempts;
// a new row is the latest.
func (s *TelegramLogin) Request(ctx context.Context, rawUsername string) error {
	username, ok := NormalizeTelegramUsername(rawUsername)
	if !ok {
		return ErrLoginBadUsername
	}
	if s.sender == nil || s.codes == nil || s.users == nil {
		return ErrLoginDelivery
	}

	now := s.now().UTC()
	_ = s.codes.DeleteOlderThan(ctx, now.Add(-24*time.Hour))

	perMinute, err := s.codes.CountSince(ctx, username, now.Add(-loginCodeMinGap))
	if err != nil {
		return err
	}
	if perMinute >= 1 {
		return ErrLoginRateLimited
	}
	perHour, err := s.codes.CountSince(ctx, username, now.Add(-time.Hour))
	if err != nil {
		return err
	}
	if perHour >= loginCodeHourlyMax {
		return ErrLoginRateLimited
	}

	user, err := s.lookup(ctx, username)
	if err != nil {
		return err
	}

	code, err := randomCode()
	if err != nil {
		return err
	}
	row := LoginCode{
		ID:         uuid.Must(uuid.NewV7()),
		TelegramID: user.TelegramID,
		Username:   username,
		CodeHash:   hashLoginCode(s.secret, username, code),
		ExpiresAt:  now.Add(loginCodeTTL),
		CreatedAt:  now,
	}
	if err := s.codes.Insert(ctx, row); err != nil {
		return err
	}
	text := fmt.Sprintf("Код для входа в LifeOS: %s\n\nДействует 5 минут. Если это не ты — проигнорируй сообщение.", code)
	if err := s.sender.Send(ctx, user.TelegramID, text); err != nil {
		_ = s.codes.Delete(ctx, row.ID)
		return fmt.Errorf("%w: %s", ErrLoginDelivery, err.Error())
	}
	return nil
}

// Verify checks the latest code and returns the user the JWT should be issued for.
func (s *TelegramLogin) Verify(ctx context.Context, rawUsername, rawCode string) (domain.User, error) {
	username, ok := NormalizeTelegramUsername(rawUsername)
	if !ok {
		return domain.User{}, ErrLoginBadUsername
	}
	code := strings.TrimSpace(rawCode)
	if len(code) != 6 {
		return domain.User{}, ErrLoginBadCode
	}
	for _, r := range code {
		if r < '0' || r > '9' {
			return domain.User{}, ErrLoginBadCode
		}
	}

	now := s.now().UTC()
	row, err := s.codes.LatestActive(ctx, username, now)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.User{}, ErrLoginBadCode
		}
		return domain.User{}, err
	}
	if row.Attempts >= loginCodeMaxAttempts {
		return domain.User{}, ErrLoginBadCode
	}

	got := hashLoginCode(s.secret, username, code)
	if !hmac.Equal([]byte(row.CodeHash), []byte(got)) {
		_ = s.codes.AddAttempt(ctx, row.ID)
		return domain.User{}, ErrLoginBadCode
	}
	_ = s.codes.Delete(ctx, row.ID)

	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.User{}, ErrLoginUnknownUser
		}
		return domain.User{}, err
	}
	if user.TelegramID != row.TelegramID {
		return domain.User{}, ErrLoginBadCode
	}
	return user, nil
}

func (s *TelegramLogin) lookup(ctx context.Context, username string) (domain.User, error) {
	user, err := s.users.GetByUsername(ctx, username)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return domain.User{}, err
	}
	if s.resolver == nil || s.ensure == nil {
		return domain.User{}, ErrLoginUnknownUser
	}
	tgID, rerr := s.resolver.Resolve(ctx, username)
	if rerr != nil || tgID <= 0 {
		return domain.User{}, ErrLoginUnknownUser
	}
	user, err = s.ensure.Execute(ctx, EnsureUserInput{
		TelegramID:  tgID,
		DisplayName: username,
		Username:    username,
	})
	if err != nil {
		return domain.User{}, err
	}
	if recErr := s.users.RememberUsername(ctx, user.ID, username); recErr != nil {
		return domain.User{}, recErr
	}
	user.TelegramUsername = username
	return user, nil
}

func randomCode() (string, error) {
	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	n := binary.BigEndian.Uint32(buf[:]) % 1000000
	return fmt.Sprintf("%06d", n), nil
}

func hashLoginCode(secret []byte, username, code string) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(username))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(code))
	return hex.EncodeToString(mac.Sum(nil))
}
