package auth

import (
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

func TestTokenServiceIssueAndParse(t *testing.T) {
	t.Parallel()
	svc, err := NewTokenService("test-secret-key-32bytes-min!!", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	userID := ids.NewUserID()
	token, exp, err := svc.Issue(userID)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" || !exp.After(time.Now()) {
		t.Fatalf("token=%q exp=%v", token, exp)
	}
	got, err := svc.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if got != userID {
		t.Fatalf("got %s want %s", got, userID)
	}
}

func TestTokenServiceRejectsEmptySecret(t *testing.T) {
	t.Parallel()
	if _, err := NewTokenService("", time.Hour); err == nil {
		t.Fatal("expected error")
	}
}
