// Package bluetooth renders the Bluetooth screen's sections.
//
// It does not know the domain: it receives view models built by the page, so
// the render does not depend on how a device is modeled.
package bluetooth

import (
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"settings-cli/internal/shared/components"
	"settings-cli/internal/shared/styles"
)

// nameColumn is the width reserved for the name; the rest is left for kind and
// status.
const nameColumn = 22

// Device is one row of the list: what is shown of a device.
type Device struct {
	Name   string
	Kind   string
	Status string
}

// Row composes name, kind and status on one line. It carries no styles of its
// own: components.List tints the whole row and an inner color would break the
// selected row's background.
func Row(t styles.Theme, d Device, width int) string {
	name := pad(truncate(t, d.Name, nameColumn), nameColumn)

	gap := width - ansi.StringWidth(name) - ansi.StringWidth(d.Kind) - ansi.StringWidth(d.Status)
	return name + d.Kind + spaces(max(gap, 1)) + d.Status
}

// Adapter is the radio's switch, aligned to the edges.
func Adapter(t styles.Theme, enabled bool, width int) string {
	const label = "Bluetooth"

	state, style := "off", t.Body.Muted
	if enabled {
		state, style = "on", t.Body.Badge
	}

	gap := width - ansi.StringWidth(label) - ansi.StringWidth(state)
	return t.Body.Label.Render(label) + spaces(max(gap, 1)) + style.Render(state)
}

// Rows draws the list with the cursor the page lends it: the section renders,
// but the cursor is not its own.
func Rows(t styles.Theme, devices []Device, list *components.List, width, height int, focused bool) []string {
	return list.Render(t, len(devices), width, height, focused,
		func(i int, base lipgloss.Style, w int) string {
			return base.Width(w).Render(truncate(t, Row(t, devices[i], w), w))
		},
	)
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
