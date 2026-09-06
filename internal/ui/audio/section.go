package audio

import (
	"strings"

	"charm.land/lipgloss/v2"

	"knob/internal/shared/components"
	"knob/internal/shared/styles"
)

// Section is a list with a heading inside the sound screen.
//
// The list is passed in from outside because it holds the cursor: the section
// renders it, but does not own it.
type Section struct {
	Label   string
	Meters  []Meter
	Empty   string
	List    *components.List
	Focused bool
}

// Render returns the section's rows: the heading, every meter row and a blank
// line to close it. It renders in full -- the page stacks the sections and
// windows the body as one, so the section itself does not scroll.
func (s Section) Render(t styles.Theme, width int) []string {
	// The label is uppercase and switches to accent when it has focus.
	style := t.Body.Section
	if s.Focused {
		style = style.Foreground(t.Color.Accent)
	}
	heading := style.Render(strings.ToUpper(s.Label))

	if len(s.Meters) == 0 {
		return []string{heading, t.List.Empty.Render(s.Empty), ""}
	}

	rows := s.List.RenderFull(t, len(s.Meters), width, s.Focused,
		func(i int, base lipgloss.Style, w int) string {
			return Row(t, s.Meters[i], base, w)
		},
	)

	return append(append([]string{heading}, rows...), "")
}
