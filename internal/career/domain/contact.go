package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var (
	ErrEmptyName = errors.New("contact name is required")
	ErrNotFound  = errors.New("contact not found")
)

type Contact struct {
	ID        ids.ContactID
	UserID    ids.UserID
	Name      string
	Company   string
	Role      string
	Notes     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewContact(userID ids.UserID, name, company, role, notes string, now time.Time) (Contact, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Contact{}, ErrEmptyName
	}
	now = now.UTC()
	return Contact{
		ID:        ids.NewContactID(),
		UserID:    userID,
		Name:      name,
		Company:   strings.TrimSpace(company),
		Role:      strings.TrimSpace(role),
		Notes:     strings.TrimSpace(notes),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func ParseContactLine(raw string) (name, company, role string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", ""
	}
	for _, sep := range []string{" — ", " - ", " / ", " | "} {
		if parts := strings.SplitN(raw, sep, 2); len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			if at := strings.LastIndex(right, " @ "); at > 0 {
				return left, strings.TrimSpace(right[at+3:]), strings.TrimSpace(right[:at])
			}
			if strings.Contains(right, "@") {
				chunks := strings.SplitN(right, "@", 2)
				return left, strings.TrimSpace(chunks[1]), strings.TrimSpace(chunks[0])
			}
			return left, right, ""
		}
	}
	return raw, "", ""
}
