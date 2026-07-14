package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var (
	ErrEmptyName   = errors.New("sphere name is required")
	ErrNotFound    = errors.New("sphere not found")
	ErrHasProjects = errors.New("sphere has linked projects")
)

var DefaultSphereNames = []string{
	"Деньги",
	"Карьера GO",
	"Цех ЧПУ",
	"Дом и быт",
	"Хобби и отдых",
}

type Sphere struct {
	ID        ids.SphereID
	UserID    ids.UserID
	Name      string
	SortOrder int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewSphere(userID ids.UserID, name string, sortOrder int32, now time.Time) (Sphere, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Sphere{}, ErrEmptyName
	}
	now = now.UTC()
	return Sphere{
		ID:        ids.NewSphereID(),
		UserID:    userID,
		Name:      name,
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *Sphere) Rename(name string, sortOrder int32, now time.Time) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyName
	}
	s.Name = name
	s.SortOrder = sortOrder
	s.UpdatedAt = now.UTC()
	return nil
}
