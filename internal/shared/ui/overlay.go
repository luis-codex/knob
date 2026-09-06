package ui

import "charm.land/lipgloss/v2"

// Overlay compone top centrado sobre base, que sigue visible alrededor.
//
// Usa Compositor y no Canvas: Canvas.Compose() ignora el X/Y de la capa.
func Overlay(base string, width, height int, top string) string {
	if width <= 0 || height <= 0 || top == "" {
		return base
	}

	topWidth, topHeight := lipgloss.Size(top)
	x := max(0, (width-topWidth)/2)
	y := max(0, (height-topHeight)/2)

	composed := lipgloss.NewCompositor(
		lipgloss.NewLayer(base).X(0).Y(0).Z(0),
		lipgloss.NewLayer(top).X(x).Y(y).Z(1),
	).Render()

	// Los límites del compositor son la unión de las capas: si top no cabe,
	// el resultado sale más grande que el área y descuadra el layout entero.
	return lipgloss.NewStyle().MaxWidth(width).MaxHeight(height).Render(composed)
}
