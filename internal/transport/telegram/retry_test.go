package telegram

import "testing"

func TestIsBenignEditError(t *testing.T) {
	t.Parallel()
	if !isBenignEditError("Bad Request: message is not modified") {
		t.Fatal("expected benign")
	}
	if isBenignEditError("chat not found") {
		t.Fatal("expected not benign")
	}
}
