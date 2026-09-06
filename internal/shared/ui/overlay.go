package ui

import "charm.land/lipgloss/v2"

// Overlay composes top centered over base, which stays visible around it.
//
// It uses Compositor and not Canvas: Canvas.Compose() ignores the layer's X/Y.
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

	// The compositor's bounds are the union of the layers: if top does not
	// fit, the result comes out larger than the area and throws off the whole
	// layout.
	return lipgloss.NewStyle().MaxWidth(width).MaxHeight(height).Render(composed)
}
