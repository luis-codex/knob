package ui

import (
	"strings"

	"settings-cli/internal/shared/styles"
)

// ButtonOpts configures a button. Focus is kept by whoever draws it.
type ButtonOpts struct {
	Text string
	// Focused marks the active button within its group.
	Focused bool
	// Danger tints the button as destructive. Applied only when focused.
	Danger bool
	// Inner horizontal padding. 0 uses the default (2).
	Padding int
}

const defaultButtonPadding = 2

// Button renders a button.
func Button(t styles.Theme, opts ButtonOpts) string {
	style := t.Button.Blurred
	switch {
	case opts.Focused && opts.Danger:
		style = t.Button.Danger
	case opts.Focused:
		style = t.Button.Focused
	}

	if opts.Padding == 0 {
		opts.Padding = defaultButtonPadding
	}

	return style.Padding(0, opts.Padding).Render(opts.Text)
}

// ButtonGroup renders a row of buttons. An empty spacing uses two spaces; "\n"
// stacks them vertically.
func ButtonGroup(t styles.Theme, buttons []ButtonOpts, spacing string) string {
	if len(buttons) == 0 {
		return ""
	}
	if spacing == "" {
		spacing = "  "
	}

	parts := make([]string, len(buttons))
	for i, b := range buttons {
		parts[i] = Button(t, b)
	}
	return strings.Join(parts, spacing)
}
