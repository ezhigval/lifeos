package telegram

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	notifapp "github.com/valentinezhov/lifeos/internal/notifications/app"
)

func TestMatchReminderToCancel(t *testing.T) {
	t.Parallel()
	first := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	second := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	items := []notifapp.ReminderDTO{
		{ID: first.String(), Message: "позвонить маме"},
		{ID: second.String(), Message: "купить молоко"},
	}

	id, hint, err := matchReminderToCancel(items, "")
	if err != nil || id != first || hint != "" {
		t.Fatalf("empty hint → first: id=%s hint=%q err=%v", id, hint, err)
	}

	id, hint, err = matchReminderToCancel(items, "  МОЛОКО ")
	if err != nil || id != second || hint != "МОЛОКО" {
		t.Fatalf("substring match: id=%s hint=%q err=%v", id, hint, err)
	}

	_, hint, err = matchReminderToCancel(items, "йога")
	if !errors.Is(err, notifapp.ErrReminderNotFound) || hint != "йога" {
		t.Fatalf("miss: hint=%q err=%v", hint, err)
	}

	_, _, err = matchReminderToCancel(nil, "x")
	if !errors.Is(err, notifapp.ErrReminderNotFound) {
		t.Fatalf("empty list: %v", err)
	}

	_, _, err = matchReminderToCancel([]notifapp.ReminderDTO{{ID: "not-a-uuid", Message: "x"}}, "")
	if err == nil {
		t.Fatal("invalid uuid must fail")
	}
}

func TestEnsureFutureFireAtEqualNowRolls(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 14, 19, 0, 0, 0, time.UTC)
	// «вечером» at exactly 19:00 must not stay in the past/equal.
	got := ensureFutureFireAt(now, now)
	if !got.After(now) {
		t.Fatalf("equal now must roll: %s", got)
	}
	if got.Day() != 15 || got.Hour() != 19 {
		t.Fatalf("expected tomorrow 19:00, got %s", got)
	}
	// Far past rolls more than one day.
	stale := time.Date(2026, 7, 12, 9, 0, 0, 0, time.UTC)
	got = ensureFutureFireAt(stale, now)
	if got.Day() != 15 || got.Hour() != 9 {
		t.Fatalf("multi-day roll: got %s", got)
	}
}

func TestSplitAndNames(t *testing.T) {
	t.Parallel()
	got := splitAndNames("Карьера и  Деньги и ")
	if len(got) != 2 || got[0] != "Карьера" || got[1] != "Деньги" {
		t.Fatalf("got %#v", got)
	}
	if len(splitAndNames("")) != 0 {
		t.Fatalf("empty: %#v", splitAndNames(""))
	}
	if got := splitAndNames("один"); len(got) != 1 || got[0] != "один" {
		t.Fatalf("single: %#v", got)
	}
}

func TestAtoiIntent(t *testing.T) {
	t.Parallel()
	if atoiIntent("09") != 9 || atoiIntent("30") != 30 {
		t.Fatal("digits")
	}
	if atoiIntent("") != 0 || atoiIntent("1a") != 0 || atoiIntent("-1") != 0 {
		t.Fatal("invalid must be 0")
	}
}
