package components

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"knob/internal/shared/styles"
)

// Header is the top bar: the application name and, on the right, the section
// you are in.
type Header struct {
	Name    string
	Section string
}

func NewHeader() Header {
	return Header{Name: "knob"}
}

// WithSection returns a copy noting where the user is.
func (c Header) WithSection(section string) Header {
	c.Section = section
	return c
}

func (c Header) View(t styles.Theme, width, height int) string {
	name := t.Header.Title.Render(c.Name)
	inner := width - t.Header.Base.GetHorizontalFrameSize()

	line := name
	if c.Section != "" {
		crumb := t.Header.Crumb.Render(c.Section)
		gap := inner - ansi.StringWidth(name) - ansi.StringWidth(crumb)
		if gap >= 1 {
			line = name + strings.Repeat(" ", gap) + crumb
		}
	}

	return t.Header.Base.Width(width).Height(height).MaxHeight(height).Render(line)
}
