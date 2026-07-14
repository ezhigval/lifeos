package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var (
	ErrEmptySkill  = errors.New("skill name is required")
	ErrSkillNotFound = errors.New("skill not found")
)

type Skill struct {
	ID        ids.SkillID
	UserID    ids.UserID
	Name      string
	Level     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewSkill(userID ids.UserID, name, level string, now time.Time) (Skill, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Skill{}, ErrEmptySkill
	}
	now = now.UTC()
	return Skill{
		ID:        ids.NewSkillID(),
		UserID:    userID,
		Name:      name,
		Level:     strings.TrimSpace(level),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func ParseSkillLine(raw string) (name, level string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}
	for _, sep := range []string{" — ", " - ", " / ", " | ", ": "} {
		if parts := strings.SplitN(raw, sep, 2); len(parts) == 2 {
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		}
	}
	parts := strings.Fields(raw)
	if len(parts) >= 2 {
		return strings.Join(parts[:len(parts)-1], " "), parts[len(parts)-1]
	}
	return raw, ""
}
