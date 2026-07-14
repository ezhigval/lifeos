package domain

import (
	"errors"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var ErrEmptyCategoryName = errors.New("category name is required")

type Category struct {
	ID        ids.CategoryID
	UserID    ids.UserID
	Name      string
	Kind      Kind
	CreatedAt time.Time
}

func NewCategory(userID ids.UserID, name string, kind Kind, now time.Time) (Category, error) {
	if name == "" {
		return Category{}, ErrEmptyCategoryName
	}
	if !kind.Valid() {
		return Category{}, ErrInvalidKind
	}
	return Category{
		ID:        ids.NewCategoryID(),
		UserID:    userID,
		Name:      name,
		Kind:      kind,
		CreatedAt: now.UTC(),
	}, nil
}

var ErrCategoryNotFound = errors.New("category not found")
