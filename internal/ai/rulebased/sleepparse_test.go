package rulebased

import "testing"

func TestParseSleepHours(t *testing.T) {
	tests := []struct {
		in   string
		want float64
	}{
		{"7", 7},
		{"7.5", 7.5},
		{"7,5 часов", 7.5},
		{"8 ч", 8},
	}
	for _, tc := range tests {
		got, err := ParseSleepHours(tc.in)
		if err != nil {
			t.Fatalf("ParseSleepHours(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("ParseSleepHours(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestSleepHoursToMinutes(t *testing.T) {
	if got := SleepHoursToMinutes(7.5); got != 450 {
		t.Fatalf("got %d", got)
	}
}
