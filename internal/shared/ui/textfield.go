package ui

import (
	"github.com/charmbracelet/x/ansi"

	"settings-cli/internal/shared/styles"
)

// TextFieldOpts configura un campo de texto. El buffer y el foco los
// mantiene quien lo dibuja.
type TextFieldOpts struct {
	Label       string
	Value       string
	Placeholder string
	Focused     bool
	// Width es el ancho total del campo, etiqueta incluida.
	Width int
}

// TextField renderiza una entrada de una línea. El cursor solo aparece con
// foco, para distinguirla de un campo inactivo con texto.
func TextField(t styles.Theme, opts TextFieldOpts) string {
	label := ""
	if opts.Label != "" {
		label = t.Input.Label.Render(opts.Label + " ")
	}

	box := opts.Width - ansi.StringWidth(label)
	if box < 4 {
		return label
	}

	style := t.Input.Blurred
	if opts.Focused {
		style = t.Input.Focused
	}

	// El cursor ocupa una celda, así que el texto se corta a box-1.
	text, cursor := opts.Value, ""
	if opts.Focused {
		cursor = t.Input.Cursor.Render(" ")
	}
	textWidth := box - ansi.StringWidth(ansi.Strip(cursor))

	// También con foco: un campo vacío sin pista no dice qué se espera.
	if text == "" && opts.Placeholder != "" {
		return label + style.Width(box).Render(
			cursor+t.Input.Placeholder.Render(fitLeft(t, opts.Placeholder, textWidth)),
		)
	}

	// Si no cabe se muestra la cola, donde está escribiendo el usuario.
	return label + style.Width(box).Render(tail(text, textWidth)+cursor)
}

// tail devuelve las últimas width columnas de s.
func tail(s string, width int) string {
	if width <= 0 {
		return ""
	}
	w := ansi.StringWidth(s)
	if w <= width {
		return s
	}
	return ansi.TruncateLeft(s, w-width, "")
}

func fitLeft(t styles.Theme, s string, width int) string {
	if width <= 0 {
		return ""
	}
	if ansi.StringWidth(s) <= width {
		return s
	}
	return ansi.Truncate(s, width, t.Icon.Ellipsis)
}
