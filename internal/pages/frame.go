// Package pages holds the app's screens: one file per menu entry. Each page
// composes its sections from internal/ui.
//
// This file is not a page: it is the common frame and the text helpers every
// page shares.
package pages

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"

	"settings-cli/internal/shared/styles"
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

// --- text -----------------------------------------------------------------

// plural formats a count with its noun.
func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return strconv.Itoa(n) + " " + many
}
