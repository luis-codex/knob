package ui

import (
	"strings"

	"settings-cli/internal/shared/styles"
)

// ButtonOpts configura un botón. El foco lo mantiene quien lo dibuja.
type ButtonOpts struct {
	Text string
	// Focused marca el botón activo dentro de su grupo.
	Focused bool
	// Danger tiñe el botón como destructivo. Solo se aplica con foco.
	Danger bool
	// Padding horizontal interno. 0 usa el valor por defecto (2).
	Padding int
}

const defaultButtonPadding = 2

// Button renderiza un botón.
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

// ButtonGroup renderiza una fila de botones. spacing vacío usa dos espacios;
// "\n" los apila en vertical.
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
