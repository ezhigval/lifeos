package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var (
	ErrInvalidKind  = errors.New("invalid memory kind")
	ErrEmptyKey     = errors.New("memory key is required")
	ErrEmptyValue   = errors.New("memory value is required")
	ErrNotFound     = errors.New("memory not found")
	ErrInvalidUser  = errors.New("user id is required")
)

type Kind string

const (
	KindPreference Kind = "preference"
	KindFact       Kind = "fact"
	KindAlias      Kind = "alias"
	KindPattern    Kind = "pattern"
)

func (k Kind) Valid() bool {
	switch k {
	case KindPreference, KindFact, KindAlias, KindPattern:
		return true
	default:
		return false
	}
}

type Memory struct {
	ID         ids.MemoryID
	UserID     ids.UserID
	Kind       Kind
	Key        string
	Value      string
	Confidence float64
	Source     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

const DefaultConfidence = 0.7
const DefaultSource = "agent"

func NewMemory(userID ids.UserID, kind Kind, key, value string, now time.Time) (Memory, error) {
	if userID.IsZero() {
		return Memory{}, ErrInvalidUser
	}
	if !kind.Valid() {
		return Memory{}, ErrInvalidKind
	}
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if key == "" {
		return Memory{}, ErrEmptyKey
	}
	if value == "" {
		return Memory{}, ErrEmptyValue
	}
	ts := now.UTC()
	return Memory{
		ID:         ids.NewMemoryID(),
		UserID:     userID,
		Kind:       kind,
		Key:        key,
		Value:      value,
		Confidence: DefaultConfidence,
		Source:     DefaultSource,
		CreatedAt:  ts,
		UpdatedAt:  ts,
	}, nil
}

type PrivacyFlags struct {
	MemoryEnabled bool
	LearningOptIn bool
}
