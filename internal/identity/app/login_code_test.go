package app

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valentinezhov/lifeos/internal/identity/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type memUsers struct {
	byTG   map[int64]domain.User
	byNick map[string]int64
}

func (m *memUsers) GetByTelegramID(_ context.Context, telegramID int64) (domain.User, error) {
	u, ok := m.byTG[telegramID]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return u, nil
}

func (m *memUsers) Upsert(_ context.Context, user domain.User) error {
	if m.byTG == nil {
		m.byTG = map[int64]domain.User{}
	}
	m.byTG[user.TelegramID] = user
	return nil
}

func (m *memUsers) GetByUsername(_ context.Context, username string) (domain.User, error) {
	id, ok := m.byNick[username]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	u, ok := m.byTG[id]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return u, nil
}

func (m *memUsers) RememberUsername(_ context.Context, userID ids.UserID, username string) error {
	if m.byNick == nil {
		m.byNick = map[string]int64{}
	}
	for tg, u := range m.byTG {
		if u.ID == userID {
			u.TelegramUsername = username
			m.byTG[tg] = u
			m.byNick[username] = tg
			return nil
		}
	}
	return domain.ErrNotFound
}

type memCodes struct {
	rows []LoginCode
}

func (m *memCodes) CountSince(_ context.Context, username string, since time.Time) (int, error) {
	n := 0
	for _, row := range m.rows {
		if row.Username == username && !row.CreatedAt.Before(since) {
			n++
		}
	}
	return n, nil
}

func (m *memCodes) LatestActive(_ context.Context, username string, now time.Time) (LoginCode, error) {
	var found LoginCode
	ok := false
	for _, row := range m.rows {
		if row.Username != username || !row.ExpiresAt.After(now) || row.Attempts >= 5 {
			continue
		}
		if !ok || row.CreatedAt.After(found.CreatedAt) {
			found = row
			ok = true
		}
	}
	if !ok {
		return LoginCode{}, domain.ErrNotFound
	}
	return found, nil
}

func (m *memCodes) Insert(_ context.Context, row LoginCode) error {
	m.rows = append(m.rows, row)
	return nil
}

func (m *memCodes) AddAttempt(_ context.Context, id uuid.UUID) error {
	for i := range m.rows {
		if m.rows[i].ID == id {
			m.rows[i].Attempts++
		}
	}
	return nil
}

func (m *memCodes) Delete(_ context.Context, id uuid.UUID) error {
	out := m.rows[:0]
	for _, row := range m.rows {
		if row.ID != id {
			out = append(out, row)
		}
	}
	m.rows = out
	return nil
}

func (m *memCodes) DeleteOlderThan(_ context.Context, before time.Time) error {
	out := m.rows[:0]
	for _, row := range m.rows {
		if !row.CreatedAt.Before(before) {
			out = append(out, row)
		}
	}
	m.rows = out
	return nil
}

type captureSender struct {
	text string
	id   int64
	fail error
}

func (c *captureSender) Send(_ context.Context, telegramID int64, text string) error {
	if c.fail != nil {
		return c.fail
	}
	c.id = telegramID
	c.text = text
	return nil
}

func seedNick(users *memUsers, nick string, tg int64) {
	id := ids.NewUserID()
	users.byTG[tg] = domain.User{ID: id, TelegramID: tg, DisplayName: nick, Timezone: "Europe/Moscow", TelegramUsername: nick}
	users.byNick[nick] = tg
}

func codeFrom(text string) string {
	m := regexp.MustCompile(`\d{6}`).FindString(text)
	return m
}

func TestTelegramLoginCodeRoundTrip(t *testing.T) {
	users := &memUsers{byTG: map[int64]domain.User{}, byNick: map[string]int64{}}
	seedNick(users, "adalovelace", 42)
	codes := &memCodes{}
	sender := &captureSender{}
	svc := NewTelegramLogin(users, codes, nil, sender, nil, "pepper")

	if err := svc.Request(context.Background(), "@AdaLovelace"); err != nil {
		t.Fatal(err)
	}
	if sender.id != 42 {
		t.Fatalf("sent to %d", sender.id)
	}
	code := codeFrom(sender.text)
	if code == "" {
		t.Fatal("message has no code")
	}
	if codes.rows[0].CodeHash == code {
		t.Fatal("code stored in plaintext")
	}

	if _, err := svc.Verify(context.Background(), "adalovelace", "000000"); err != ErrLoginBadCode {
		t.Fatalf("wrong code: %v", err)
	}
	user, err := svc.Verify(context.Background(), "AdaLovelace", code)
	if err != nil {
		t.Fatal(err)
	}
	if user.TelegramID != 42 {
		t.Fatalf("user %d", user.TelegramID)
	}
	if _, err := svc.Verify(context.Background(), "adalovelace", code); err != ErrLoginBadCode {
		t.Fatalf("replay: %v", err)
	}
}

func TestTelegramLoginRateLimitAndUnknown(t *testing.T) {
	users := &memUsers{byTG: map[int64]domain.User{}, byNick: map[string]int64{}}
	seedNick(users, "adalovelace", 42)
	svc := NewTelegramLogin(users, &memCodes{}, nil, &captureSender{}, nil, "pepper")
	if err := svc.Request(context.Background(), "adalovelace"); err != nil {
		t.Fatal(err)
	}
	if err := svc.Request(context.Background(), "adalovelace"); err != ErrLoginRateLimited {
		t.Fatalf("second request: %v", err)
	}
	if err := svc.Request(context.Background(), "nope"); err != ErrLoginBadUsername {
		t.Fatalf("short nick: %v", err)
	}
	if err := svc.Request(context.Background(), "unknownname"); err != ErrLoginUnknownUser {
		t.Fatalf("unknown: %v", err)
	}
}

func TestTelegramLoginResolvesViaBot(t *testing.T) {
	users := &memUsers{byTG: map[int64]domain.User{}, byNick: map[string]int64{}}
	codes := &memCodes{}
	sender := &captureSender{}
	ensure := NewEnsureUserByTelegram(users, nil, "Europe/Moscow", nil)
	resolver := usernameResolverFunc(func(_ context.Context, username string) (int64, error) {
		if username != "ezhigval" {
			return 0, domain.ErrNotFound
		}
		return 77, nil
	})
	svc := NewTelegramLogin(users, codes, ensure, sender, resolver, "pepper")
	if err := svc.Request(context.Background(), "ezhigval"); err != nil {
		t.Fatal(err)
	}
	user, err := svc.Verify(context.Background(), "ezhigval", codeFrom(sender.text))
	if err != nil {
		t.Fatal(err)
	}
	if user.TelegramID != 77 || user.TelegramUsername != "ezhigval" {
		t.Fatalf("user %+v", user)
	}
}

func TestTelegramLoginDeliveryFailureDropsCode(t *testing.T) {
	users := &memUsers{byTG: map[int64]domain.User{}, byNick: map[string]int64{}}
	seedNick(users, "adalovelace", 42)
	codes := &memCodes{}
	sender := &captureSender{fail: domain.ErrNotFound}
	svc := NewTelegramLogin(users, codes, nil, sender, nil, "pepper")
	err := svc.Request(context.Background(), "adalovelace")
	if err == nil {
		t.Fatal("expected delivery error")
	}
	if len(codes.rows) != 0 {
		t.Fatalf("code left behind: %d", len(codes.rows))
	}
}

type usernameResolverFunc func(ctx context.Context, username string) (int64, error)

func (f usernameResolverFunc) Resolve(ctx context.Context, username string) (int64, error) {
	return f(ctx, username)
}
