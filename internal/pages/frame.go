// Package pages holds the app's screens: one file per menu entry. Each page
// composes its sections from internal/ui.
//
// This file is not a page: it is the common frame and the text helpers every
// page shares.
package pages

import (
	"strings"

	"charm.land/lipgloss/v2"

	"knob/internal/shared/styles"
)

// frameChrome is the rows frame adds on its own: the title, the rule that
// separates it and the blank line. Pages subtract it to measure how much
// height is left for their content.
const frameChrome = 3

// frame renders the common skeleton: title, separator and body.
func frame(t styles.Theme, width, height int, title string, rows ...string) string {
	inner := width - t.Body.Base.GetHorizontalFrameSize()
	if inner <= 0 {
		return ""
	}

	// A rule under the title separates the header from the content, like the
	// top border of bun.com's cards. Without it the page is a flat block of
	// text.
	lines := append([]string{
		t.Body.Title.Render(title),
		t.Divider.Render(strings.Repeat(t.Icon.DividerH, max(inner, 0))),
		"",
	}, rows...)

	// Height pads but does not clip; without MaxHeight the body would overflow.
	return t.Body.Base.Width(width).Height(height).MaxHeight(height).Render(
		lipgloss.JoinVertical(lipgloss.Left, lines...),
	)
}

// fit clips s to width columns, respecting the ANSI codes.
func fit(s string, width int) string {
	if width <= 0 {
		return ""
	}
	return lipgloss.NewStyle().MaxWidth(width).Render(s)
}

// pad right-fills s with spaces to width columns; it is fit's counterpart, for
// short rows that must reach the edge so a trailing scrollbar lands there.
func pad(s string, width int) string {
	if gap := width - lipgloss.Width(s); gap > 0 {
		return s + strings.Repeat(" ", gap)
	}
	return s
}
