package pulse

import "testing"

// relevant filtra las líneas de `pactl subscribe`. El mezclador depende de que
// los sink-input cuenten; el cliente, no, o la relectura se realimentaría.
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
		{"ruido sin la palabra clave", false},
		{"", false},
	}

	for _, tc := range tests {
		if got := relevant(tc.line); got != tc.want {
			t.Errorf("relevant(%q) = %v, se esperaba %v", tc.line, got, tc.want)
		}
	}
}
