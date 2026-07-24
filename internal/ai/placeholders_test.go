package ai

import "testing"

func TestIsPlaceholderTitle(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want bool
	}{
		{"", true},
		{"разобрать входящие", true},
		{"  Разобрать Почту ", true},
		{"todo", true},
		{"купить хлеб", false},
		{"разобрать склад", false},
	}
	for _, tc := range cases {
		if got := IsPlaceholderTitle(tc.in); got != tc.want {
			t.Fatalf("IsPlaceholderTitle(%q)=%v want %v", tc.in, got, tc.want)
		}
	}
}
