package domain_test

import (
	"testing"

	"github.com/valentinezhov/lifeos/internal/settings/domain"
)

func TestParseTimeOfDay(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in      string
		wantH   int
		wantM   int
		wantErr bool
	}{
		{"8", 8, 0, false},
		{"08:30", 8, 30, false},
		{"21:00", 21, 0, false},
		{"", 0, 0, true},
		{"25", 0, 0, true},
	}
	for _, tc := range cases {
		got, err := domain.ParseTimeOfDay(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("ParseTimeOfDay(%q) expected error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("ParseTimeOfDay(%q): %v", tc.in, err)
		}
		if got.Hour != tc.wantH || got.Minute != tc.wantM {
			t.Fatalf("ParseTimeOfDay(%q) = %+v", tc.in, got)
		}
	}
}
