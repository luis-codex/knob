// Package layouts composes screen regions: it computes measurements and joins
// them with dividers, without knowing what is inside.
package layouts

import (
	"charm.land/lipgloss/v2"

	"settings-cli/internal/shared/styles"
	"settings-cli/internal/shared/ui"
)

// Section renders into the area it is given. It is declared by this package,
// which is the one that consumes it.
type Section interface {
	View(t styles.Theme, width, height int) string
}

// Minimum measurements for the layout to make sense.
const (
	minWidth  = styles.SidebarWidth + styles.DividerSize + 20
	minHeight = styles.HeaderHeight + styles.FooterHeight + 2*styles.DividerSize + 3
)

// App is the main layout:
//
//	┌──────────────────────────────┐
//	│ header                       │
//	├──────────────────────────────┤
//	│ sidebar │ body               │
//	├──────────────────────────────┤
//	│ footer                       │
//	└──────────────────────────────┘
type App struct {
	Header  Section
	Sidebar Section
	Body    Section
	Footer  Section

	// Overlay is composed centered over the body, which stays visible beneath.
	Overlay Section
}

func (a App) View(t styles.Theme, width, height int) string {
	if width < minWidth || height < minHeight {
		// Fill the whole area: returning just the text leaves the frame with a
		// row of the wrong width.
		return t.Body.Base.
			Width(width).Height(height).
			MaxWidth(width).MaxHeight(height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Terminal too small")
	}

	// Vertical: header + divider + center + divider + footer.
	centerHeight := height - styles.HeaderHeight - styles.FooterHeight - 2*styles.DividerSize
	// Horizontal: sidebar + divider + body.
	sidebarWidth := styles.SidebarWidth
	bodyWidth := width - sidebarWidth - styles.DividerSize

	body := a.Body.View(t, bodyWidth, centerHeight)
	if a.Overlay != nil {
		body = ui.Overlay(body, bodyWidth, centerHeight, a.Overlay.View(t, bodyWidth, centerHeight))
	}

	center := lipgloss.JoinHorizontal(
		lipgloss.Top,
		a.Sidebar.View(t, sidebarWidth, centerHeight),
		ui.VDivider(t, centerHeight),
		body,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		a.Header.View(t, width, styles.HeaderHeight),
		ui.HDivider(t, width),
		center,
		ui.HDivider(t, width),
		a.Footer.View(t, width, styles.FooterHeight),
	)
}
