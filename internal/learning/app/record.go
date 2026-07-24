package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/learning/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type EventStore interface {
	Insert(ctx context.Context, ev domain.Event) error
}

type RecordEvent struct {
	store EventStore
	now   func() time.Time
}

func NewRecordEvent(store EventStore) *RecordEvent {
	return &RecordEvent{
		store: store,
		now:   func() time.Time { return time.Now().UTC() },
	}
}

type EventInput struct {
	AnonSubject  string
	Type         string
	ToolOrIntent string
	Success      *bool
	AskRounds    int
	Model        string
	Meta         map[string]any
}

func (uc *RecordEvent) Execute(ctx context.Context, in EventInput) error {
	anon := strings.TrimSpace(in.AnonSubject)
	if anon == "" {
		return fmt.Errorf("anon subject is required")
	}
	eventType := strings.TrimSpace(in.Type)
	if eventType == "" {
		return fmt.Errorf("event type is required")
	}
	if in.AskRounds < 0 {
		return fmt.Errorf("ask rounds must be >= 0")
	}
	meta := in.Meta
	if meta == nil {
		meta = map[string]any{}
	}
	ev := domain.Event{
		ID:           ids.NewLearningEventID(),
		AnonSubject:  anon,
		Type:         eventType,
		ToolOrIntent: strings.TrimSpace(in.ToolOrIntent),
		Success:      in.Success,
		AskRounds:    in.AskRounds,
		Model:        strings.TrimSpace(in.Model),
		Meta:         meta,
		CreatedAt:    uc.now(),
	}
	if err := uc.store.Insert(ctx, ev); err != nil {
		return fmt.Errorf("record learning event: %w", err)
	}
	return nil
}
