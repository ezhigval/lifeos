package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/identity/app"
	"github.com/valentinezhov/lifeos/internal/identity/domain"
)

type seedRepo struct {
	users map[int64]domain.User
}

func (s *seedRepo) GetByTelegramID(_ context.Context, telegramID int64) (domain.User, error) {
	u, ok := s.users[telegramID]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return u, nil
}

func (s *seedRepo) Upsert(_ context.Context, user domain.User) error {
	if s.users == nil {
		s.users = make(map[int64]domain.User)
	}
	if existing, ok := s.users[user.TelegramID]; ok {
		user.ID = existing.ID
		user.CreatedAt = existing.CreatedAt
	}
	s.users[user.TelegramID] = user
	return nil
}

func TestSeedUserReturnsExistingID(t *testing.T) {
	t.Parallel()

	existing, _ := domain.NewUser(42, "Old", "Europe/Moscow", time.Now().UTC())
	repo := &seedRepo{users: map[int64]domain.User{42: existing}}
	uc := app.NewSeedUser(repo)

	got, err := uc.Execute(context.Background(), app.SeedInput{
		TelegramID: 42, DisplayName: "New", Timezone: "Europe/Moscow",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != existing.ID {
		t.Fatalf("seed returned new id %s, want existing %s", got.ID, existing.ID)
	}
	if got.DisplayName != "New" {
		t.Fatalf("display name = %q", got.DisplayName)
	}
}
