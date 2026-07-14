package domain

import "testing"

func TestNewMoney(t *testing.T) {
	t.Parallel()
	m, err := NewMoney(5_000_000, "RUB")
	if err != nil || m.AmountCents != 5_000_000 {
		t.Fatalf("got %+v err=%v", m, err)
	}
	if _, err := NewMoney(0, "RUB"); err == nil {
		t.Fatal("expected error")
	}
}

func TestFormatMoney(t *testing.T) {
	t.Parallel()
	if got := FormatMoney(Money{AmountCents: 5_000_000}); got != "50000 ₽" {
		t.Fatalf("got %q", got)
	}
}
