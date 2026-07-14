package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/identity/app"
	"github.com/valentinezhov/lifeos/internal/identity/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type stubUserRepo struct {
	users map[int64]domain.User
}

func (s *stubUserRepo) GetByTelegramID(_ context.Context, telegramID int64) (domain.User, error) {
	u, ok := s.users[telegramID]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return u, nil
}

func (s *stubUserRepo) Upsert(_ context.Context, user domain.User) error {
	if s.users == nil {
		s.users = make(map[int64]domain.User)
	}
	s.users[user.TelegramID] = user
	return nil
}

type stubSettings struct {
	called bool
	userID ids.UserID
}

func (s *stubSettings) Execute(_ context.Context, userID ids.UserID) error {
	s.called = true
	s.userID = userID
	return nil
}

func TestEnsureUserByTelegramReturnsExisting(t *testing.T) {
	t.Parallel()

	existing, _ := domain.NewUser(42, "Existing", "Europe/Moscow", time.Now().UTC())
	repo := &stubUserRepo{users: map[int64]domain.User{42: existing}}
	uc := app.NewEnsureUserByTelegram(repo, nil, "Europe/Moscow", nil)

	got, err := uc.Execute(context.Background(), app.EnsureUserInput{TelegramID: 42, DisplayName: "New"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != existing.ID {
		t.Fatalf("got new user id %s, want existing %s", got.ID, existing.ID)
	}
}

func TestEnsureUserByTelegramCreatesIsolatedUser(t *testing.T) {
	t.Parallel()

	repo := &stubUserRepo{}
	settings := &stubSettings{}
	var created domain.User
	uc := app.NewEnsureUserByTelegram(repo, settings, "Europe/Moscow", func(_ context.Context, user domain.User) error {
		created = user
		return nil
	})

	got, err := uc.Execute(context.Background(), app.EnsureUserInput{TelegramID: 99, DisplayName: "Alice"})
	if err != nil {
		t.Fatal(err)
	}
	if got.TelegramID != 99 || got.DisplayName != "Alice" {
		t.Fatalf("unexpected user %+v", got)
	}
	if !settings.called || settings.userID != got.ID {
		t.Fatal("expected default settings for new user")
	}
	if created.ID != got.ID {
		t.Fatal("onCreated should receive new user")
	}

	other, err := uc.Execute(context.Background(), app.EnsureUserInput{TelegramID: 100, DisplayName: "Bob"})
	if err != nil {
		t.Fatal(err)
	}
	if other.ID == got.ID {
		t.Fatal("users must have distinct ids")
	}
}
