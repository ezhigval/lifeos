package app_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/memory/app"
	"github.com/valentinezhov/lifeos/internal/memory/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type memoryStoreFake struct {
	byKey map[string]domain.Memory
}

func (s *memoryStoreFake) keyOf(userID ids.UserID, kind domain.Kind, key string) string {
	return userID.String() + "|" + string(kind) + "|" + key
}

func (s *memoryStoreFake) Upsert(_ context.Context, mem domain.Memory) (domain.Memory, error) {
	if s.byKey == nil {
		s.byKey = map[string]domain.Memory{}
	}
	k := s.keyOf(mem.UserID, mem.Kind, mem.Key)
	if existing, ok := s.byKey[k]; ok {
		existing.Value = mem.Value
		existing.Confidence = mem.Confidence
		existing.Source = mem.Source
		existing.UpdatedAt = mem.UpdatedAt
		s.byKey[k] = existing
		return existing, nil
	}
	s.byKey[k] = mem
	return mem, nil
}

func (s *memoryStoreFake) List(_ context.Context, userID ids.UserID, limit int) ([]domain.Memory, error) {
	out := make([]domain.Memory, 0)
	for _, mem := range s.byKey {
		if mem.UserID == userID {
			out = append(out, mem)
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *memoryStoreFake) Recall(_ context.Context, userID ids.UserID, query string, limit int) ([]domain.Memory, error) {
	q := strings.ToLower(query)
	out := make([]domain.Memory, 0)
	for _, mem := range s.byKey {
		if mem.UserID != userID {
			continue
		}
		if q == "" ||
			strings.Contains(strings.ToLower(mem.Key), q) ||
			strings.Contains(strings.ToLower(mem.Value), q) {
			out = append(out, mem)
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *memoryStoreFake) Delete(_ context.Context, userID ids.UserID, memoryID ids.MemoryID) error {
	for k, mem := range s.byKey {
		if mem.UserID == userID && mem.ID == memoryID {
			delete(s.byKey, k)
			return nil
		}
	}
	return domain.ErrNotFound
}

func TestUpsertMemoryInsertAndUpdate(t *testing.T) {
	t.Parallel()
	store := &memoryStoreFake{}
	uc := app.NewUpsertMemory(store)
	userID := ids.NewUserID()

	first, err := uc.Execute(context.Background(), app.UpsertMemoryInput{
		UserID: userID,
		Kind:   domain.KindPreference,
		Key:    "currency",
		Value:  "RUB",
	})
	if err != nil {
		t.Fatal(err)
	}
	if first.ID.IsZero() || first.Value != "RUB" || first.Confidence != domain.DefaultConfidence {
		t.Fatalf("first = %+v", first)
	}

	second, err := uc.Execute(context.Background(), app.UpsertMemoryInput{
		UserID:     userID,
		Kind:       domain.KindPreference,
		Key:        "currency",
		Value:      "USD",
		Confidence: 0.9,
		Source:     "user",
	})
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != first.ID {
		t.Fatalf("upsert should keep id: %s vs %s", first.ID, second.ID)
	}
	if second.Value != "USD" || second.Confidence != 0.9 || second.Source != "user" {
		t.Fatalf("second = %+v", second)
	}
	if !second.UpdatedAt.After(first.CreatedAt.Add(-time.Second)) {
		t.Fatalf("updated_at should be set: %+v", second)
	}
	if len(store.byKey) != 1 {
		t.Fatalf("expected single row, got %d", len(store.byKey))
	}
}

func TestListAndRecallMemories(t *testing.T) {
	t.Parallel()
	store := &memoryStoreFake{}
	uc := app.NewUpsertMemory(store)
	userID := ids.NewUserID()

	_, _ = uc.Execute(context.Background(), app.UpsertMemoryInput{
		UserID: userID, Kind: domain.KindFact, Key: "city", Value: "Moscow",
	})
	_, _ = uc.Execute(context.Background(), app.UpsertMemoryInput{
		UserID: userID, Kind: domain.KindAlias, Key: "boss", Value: "Alex",
	})

	listUC := app.NewListMemories(store)
	all, err := listUC.Execute(context.Background(), userID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("list = %d", len(all))
	}

	recallUC := app.NewRecall(store)
	hits, err := recallUC.Execute(context.Background(), userID, "mos", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].Key != "city" {
		t.Fatalf("recall = %+v", hits)
	}
}

func TestDeleteMemory(t *testing.T) {
	t.Parallel()
	store := &memoryStoreFake{}
	uc := app.NewUpsertMemory(store)
	userID := ids.NewUserID()
	mem, err := uc.Execute(context.Background(), app.UpsertMemoryInput{
		UserID: userID, Kind: domain.KindPattern, Key: "morning", Value: "coffee first",
	})
	if err != nil {
		t.Fatal(err)
	}
	del := app.NewDeleteMemory(store)
	if err := del.Execute(context.Background(), userID, mem.ID); err != nil {
		t.Fatal(err)
	}
	if len(store.byKey) != 0 {
		t.Fatalf("expected empty store, got %+v", store.byKey)
	}
}

func TestUpsertMemoryValidation(t *testing.T) {
	t.Parallel()
	uc := app.NewUpsertMemory(&memoryStoreFake{})
	_, err := uc.Execute(context.Background(), app.UpsertMemoryInput{
		UserID: ids.NewUserID(), Kind: "nope", Key: "k", Value: "v",
	})
	if err != domain.ErrInvalidKind {
		t.Fatalf("kind err = %v", err)
	}
}
