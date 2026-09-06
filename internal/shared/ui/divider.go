package ui

import (
	"strings"

	"settings-cli/internal/shared/styles"
)

// HDivider dibuja una línea horizontal de ancho width.
func HDivider(t styles.Theme, width int) string {
	if width <= 0 {
		return ""
	}
	return t.Divider.Render(strings.Repeat(t.Icon.DividerH, width))
}

// VDivider dibuja una columna vertical de height filas.
func VDivider(t styles.Theme, height int) string {
	if height <= 0 {
		return ""
	}
	col := make([]string, height)
	for i := range col {
		col[i] = t.Icon.DividerV
	}
	return t.Divider.Render(strings.Join(col, "\n"))
}
