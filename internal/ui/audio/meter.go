// Package audio renders the sound screen's sections.
//
// It does not know the domain: it receives Meter, a view model built by the
// page. That way the render does not depend on whether there is a device or a
// stream behind it, and it can be tested without standing up half a system.
package audio

import (
	"strconv"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"knob/internal/shared/styles"
)

const (
	// BarSegments is the number of bar segments. They span 0-100%; above that
	// the bar is full.
	BarSegments = 13
	// NominalLevel is 100%, the maximum without amplification.
	NominalLevel = 100

	// The row is a grid of fixed columns. If any depended on the content --
	// the level takes 3 or 4 characters -- the bar would jump from one row to
	// the next and nothing would line up.
	markerColumn = 2
	gapColumn    = 2
	levelColumn  = 5 // "100%", "45%" or "mute", on the right
	maxNameWidth = 38
	minNameWidth = 12
)

// Meter is one row of the meter: what to render, without knowing where it
// comes from.
type Meter struct {
	// Marker is the indicator's symbol; empty if there is none. Accent decides
	// whether it is tinted, rather than comparing the symbol, which would tie
	// the color to the glyph.
	Marker string
	Accent bool

	Name  string
	Level int
	Muted bool
}

// Row composes the row span by span.
//
// Each piece has base applied, spaces included: a mid-string ANSI reset would
// take the selected row's background from there on.
func Row(t styles.Theme, m Meter, base lipgloss.Style, width int) string {
	level := padLeft(levelText(m), levelColumn)

	// With no room for the bar it is dropped: clipping the name further would
	// make it unreadable, and a row that overflows wraps and throws off the
	// height.
	if !fitsBar(width) {
		name := pad(truncate(t, m.Name, compactNameWidth(width)), compactNameWidth(width))
		return marker(t, m, base) + base.Render(name+spaces(gapColumn)) +
			base.Foreground(t.Color.Text).Bold(true).Render(level)
	}

	column := nameWidth(width)
	name := pad(truncate(t, m.Name, column), column)

	// The slack goes at the end and not between columns: that way the control
	// block stays anchored to the name instead of drifting away on wide
	// terminals.
	used := markerColumn + column + gapColumn + barWidth() + gapColumn + levelColumn

	return marker(t, m, base) + base.Render(name+spaces(gapColumn)) +
		bar(t, m, base) +
		base.Render(spaces(gapColumn)) +
		base.Foreground(t.Color.Text).Bold(true).Render(level) +
		base.Render(spaces(max(width-used, 0)))
}

// marker renders the indicator and its gap. With no symbol it leaves the slot,
// so the grid does not shift between rows.
func marker(t styles.Theme, m Meter, base lipgloss.Style) string {
	if m.Marker == "" {
		return base.Render(spaces(markerColumn))
	}

	style := base
	if m.Accent {
		style = base.Foreground(t.Color.Accent)
	}
	return style.Render(m.Marker) + base.Render(spaces(markerColumn-1))
}

// bar draws the full and empty segments. A muted meter shows them all empty
// even if it has volume: what matters is what is heard, not what is set.
func bar(t styles.Theme, m Meter, base lipgloss.Style) string {
	filled := 0
	if !m.Muted {
		filled = min(m.Level*BarSegments/NominalLevel, BarSegments)
	}

	on := base.Foreground(t.Color.Muted)
	off := base.Foreground(t.Color.Line)

	out := ""
	for i := range BarSegments {
		if i > 0 {
			out += base.Render(" ")
		}
		if i < filled {
			out += on.Render(t.Icon.BarOn)
			continue
		}
		out += off.Render(t.Icon.BarOff)
	}
	return out
}

// levelText always fits in levelColumn: "muted" would shift the grid.
func levelText(m Meter) string {
	if m.Muted {
		return "mute"
	}
	return strconv.Itoa(m.Level) + "%"
}

// barWidth is what the bar takes: one character per segment plus the
// separators.
func barWidth() int { return BarSegments*2 - 1 }

// fitsBar reports whether the full grid fits with a readable name.
func fitsBar(width int) bool { return width-fixedColumns() >= minNameWidth }

// fixedColumns is everything that is not the name.
func fixedColumns() int {
	return markerColumn + gapColumn + barWidth() + gapColumn + levelColumn
}

// nameWidth is what is left for the name, capped so that on wide terminals the
// control does not fly to the other end.
func nameWidth(width int) int {
	return min(max(width-fixedColumns(), minNameWidth), maxNameWidth)
}

// compactNameWidth is the name when there is no bar.
func compactNameWidth(width int) int {
	return max(width-markerColumn-gapColumn-levelColumn, 1)
}

func truncate(t styles.Theme, s string, width int) string {
	if width <= 0 {
		return ""
	}
	if ansi.StringWidth(s) <= width {
		return s
	}
	return ansi.Truncate(s, width, t.Icon.Ellipsis)
}

func pad(s string, width int) string { return s + spaces(width-ansi.StringWidth(s)) }

func padLeft(s string, width int) string { return spaces(width-ansi.StringWidth(s)) + s }

func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	out := make([]byte, n)
	for i := range out {
		out[i] = ' '
	}
	return string(out)
}
