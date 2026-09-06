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

// Chrome is the rows a section spends apart from its list: the heading and the
// blank line that closes it.
const Chrome = 2

// Render returns the section's rows.
func (s Section) Render(t styles.Theme, width, height int) []string {
	// The label is uppercase and switches to accent when it has focus.
	style := t.Body.Section
	if s.Focused {
		style = style.Foreground(t.Color.Accent)
	}
	heading := style.Render(strings.ToUpper(s.Label))

	if len(s.Meters) == 0 {
		return []string{heading, t.List.Empty.Render(s.Empty), ""}
	}

	rows := s.List.Render(t, len(s.Meters), width, height, s.Focused,
		func(i int, base lipgloss.Style, w int) string {
			return Row(t, s.Meters[i], base, w)
		},
	)

	return append(append([]string{heading}, rows...), "")
}
