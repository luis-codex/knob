package pulse

import "testing"

// relevant filters the lines of `pactl subscribe`. The mixer depends on
// sink-input events counting; the client does not, or the re-read would feed
// itself.
func TestRelevant(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"Event 'change' on sink #58", true},
		{"Event 'new' on sink-input #204", true},
		{"Event 'remove' on sink-input #204", true},
		{"Event 'change' on source #12", true},
		{"Event 'change' on server", true},
		{"Event 'new' on card #9", true},
		{"Event 'new' on client #77", false},
		{"Event 'change' on module #3", false},
		{"noise without the keyword", false},
		{"", false},
	}

	for _, tc := range tests {
		if got := relevant(tc.line); got != tc.want {
			t.Errorf("relevant(%q) = %v, want %v", tc.line, got, tc.want)
		}
	}
}
