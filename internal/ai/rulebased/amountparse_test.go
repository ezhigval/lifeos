package rulebased

import "testing"

func TestParseRublesAmount(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want int64
	}{
		{"50 тысяч", 5_000_000},
		{"50к", 5_000_000},
		{"50000", 5_000_000},
		{"50 000 руб", 5_000_000},
		{"10 тысяч", 1_000_000},
	}
	for _, tc := range cases {
		got, err := ParseRublesAmount(tc.in)
		if err != nil {
			t.Fatalf("ParseRublesAmount(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("ParseRublesAmount(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
