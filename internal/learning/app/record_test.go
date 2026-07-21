package app_test

import (
	"context"
	"testing"

	"github.com/valentinezhov/lifeos/internal/learning/app"
	"github.com/valentinezhov/lifeos/internal/learning/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type eventStoreFake struct {
	events []domain.Event
}

func (s *eventStoreFake) Insert(_ context.Context, ev domain.Event) error {
	s.events = append(s.events, ev)
	return nil
}

func TestRecordEvent(t *testing.T) {
	t.Parallel()
	store := &eventStoreFake{}
	uc := app.NewRecordEvent(store)
	ok := true
	err := uc.Execute(context.Background(), app.EventInput{
		AnonSubject:  domain.AnonSubject(ids.NewUserID(), "salt"),
		Type:         "tool_call",
		ToolOrIntent: "create_task",
		Success:      &ok,
		AskRounds:    1,
		Model:        "gpt-test",
		Meta:         map[string]any{"has_slots": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d", len(store.events))
	}
	ev := store.events[0]
	if ev.ID.IsZero() || ev.Type != "tool_call" || ev.AskRounds != 1 {
		t.Fatalf("event = %+v", ev)
	}
	if ev.Success == nil || !*ev.Success {
		t.Fatalf("success = %+v", ev.Success)
	}
}

func TestRecordEventRequiresAnonAndType(t *testing.T) {
	t.Parallel()
	uc := app.NewRecordEvent(&eventStoreFake{})
	if err := uc.Execute(context.Background(), app.EventInput{Type: "x"}); err == nil {
		t.Fatal("expected anon subject error")
	}
	if err := uc.Execute(context.Background(), app.EventInput{AnonSubject: "abc"}); err == nil {
		t.Fatal("expected type error")
	}
}
