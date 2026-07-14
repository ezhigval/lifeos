package rulebased

import "testing"

func TestParseStepCount(t *testing.T) {
	tests := []struct {
		in   string
		want int32
	}{
		{"8000", 8000},
		{"10 000", 10000},
		{"  12345  ", 12345},
	}
	for _, tc := range tests {
		got, err := ParseStepCount(tc.in)
		if err != nil {
			t.Fatalf("ParseStepCount(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("ParseStepCount(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
