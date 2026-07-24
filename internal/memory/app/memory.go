package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/memory/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type MemoryStore interface {
	Upsert(ctx context.Context, mem domain.Memory) (domain.Memory, error)
	List(ctx context.Context, userID ids.UserID, limit int) ([]domain.Memory, error)
	Recall(ctx context.Context, userID ids.UserID, query string, limit int) ([]domain.Memory, error)
	Delete(ctx context.Context, userID ids.UserID, memoryID ids.MemoryID) error
}

type PrivacyStore interface {
	GetPrivacyFlags(ctx context.Context, userID ids.UserID) (domain.PrivacyFlags, error)
	SetPrivacyFlags(ctx context.Context, userID ids.UserID, flags domain.PrivacyFlags) error
}

type UpsertMemory struct {
	store MemoryStore
	now   func() time.Time
}

func NewUpsertMemory(store MemoryStore) *UpsertMemory {
	return &UpsertMemory{
		store: store,
		now:   func() time.Time { return time.Now().UTC() },
	}
}

type UpsertMemoryInput struct {
	UserID     ids.UserID
	Kind       domain.Kind
	Key        string
	Value      string
	Confidence float64
	Source     string
}

func (uc *UpsertMemory) Execute(ctx context.Context, in UpsertMemoryInput) (domain.Memory, error) {
	mem, err := domain.NewMemory(in.UserID, in.Kind, in.Key, in.Value, uc.now())
	if err != nil {
		return domain.Memory{}, err
	}
	if in.Confidence > 0 {
		mem.Confidence = in.Confidence
	}
	if s := strings.TrimSpace(in.Source); s != "" {
		mem.Source = s
	}
	out, err := uc.store.Upsert(ctx, mem)
	if err != nil {
		return domain.Memory{}, fmt.Errorf("upsert memory: %w", err)
	}
	return out, nil
}

type ListMemories struct {
	store MemoryStore
}

func NewListMemories(store MemoryStore) *ListMemories {
	return &ListMemories{store: store}
}

func (uc *ListMemories) Execute(ctx context.Context, userID ids.UserID, limit int) ([]domain.Memory, error) {
	if userID.IsZero() {
		return nil, domain.ErrInvalidUser
	}
	limit = normalizeLimit(limit, 50)
	out, err := uc.store.List(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list memories: %w", err)
	}
	return out, nil
}

type Recall struct {
	store MemoryStore
}

func NewRecall(store MemoryStore) *Recall {
	return &Recall{store: store}
}

func (uc *Recall) Execute(ctx context.Context, userID ids.UserID, query string, limit int) ([]domain.Memory, error) {
	if userID.IsZero() {
		return nil, domain.ErrInvalidUser
	}
	limit = normalizeLimit(limit, 20)
	query = strings.TrimSpace(query)
	out, err := uc.store.Recall(ctx, userID, query, limit)
	if err != nil {
		return nil, fmt.Errorf("recall memories: %w", err)
	}
	return out, nil
}

type DeleteMemory struct {
	store MemoryStore
}

func NewDeleteMemory(store MemoryStore) *DeleteMemory {
	return &DeleteMemory{store: store}
}

func (uc *DeleteMemory) Execute(ctx context.Context, userID ids.UserID, memoryID ids.MemoryID) error {
	if userID.IsZero() {
		return domain.ErrInvalidUser
	}
	if memoryID.IsZero() {
		return fmt.Errorf("memory id is required")
	}
	if err := uc.store.Delete(ctx, userID, memoryID); err != nil {
		return fmt.Errorf("delete memory: %w", err)
	}
	return nil
}

func normalizeLimit(limit, fallback int) int {
	if limit <= 0 {
		return fallback
	}
	if limit > 200 {
		return 200
	}
	return limit
}
