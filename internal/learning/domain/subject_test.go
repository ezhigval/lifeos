package domain_test

import (
	"testing"

	"github.com/valentinezhov/lifeos/internal/learning/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

func TestAnonSubjectStableAndSaltSensitive(t *testing.T) {
	t.Parallel()
	userID := ids.NewUserID()
	salt := "test-salt"

	a := domain.AnonSubject(userID, salt)
	b := domain.AnonSubject(userID, salt)
	if a != b {
		t.Fatalf("AnonSubject not stable: %q vs %q", a, b)
	}
	if len(a) != 64 {
		t.Fatalf("expected sha256 hex length 64, got %d (%q)", len(a), a)
	}
	if a == userID.String() {
		t.Fatal("anon subject must not equal raw user id")
	}

	otherSalt := domain.AnonSubject(userID, "other-salt")
	if otherSalt == a {
		t.Fatal("different salt must change anon subject")
	}

	otherUser := domain.AnonSubject(ids.NewUserID(), salt)
	if otherUser == a {
		t.Fatal("different user must change anon subject")
	}
}
