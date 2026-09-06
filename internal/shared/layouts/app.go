// Package layouts compone regiones de pantalla: calcula medidas y las une
// con dividers, sin saber qué hay dentro.
package layouts

import (
	"charm.land/lipgloss/v2"

	"settings-cli/internal/shared/styles"
	"settings-cli/internal/shared/ui"
)

// Section se renderiza en el área que le den. La declara este paquete, que
// es quien la consume.
type Section interface {
	View(t styles.Theme, width, height int) string
}

// Medidas mínimas para que el layout tenga sentido.
const (
	minWidth  = styles.SidebarWidth + styles.DividerSize + 20
	minHeight = styles.HeaderHeight + styles.FooterHeight + 2*styles.DividerSize + 3
)

// App es el layout principal:
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

	// Overlay se compone centrado sobre el body, que sigue visible debajo.
	Overlay Section
}

func (a App) View(t styles.Theme, width, height int) string {
	if width < minWidth || height < minHeight {
		// Rellena el área entera: devolver solo el texto deja el marco con
		// una fila del ancho equivocado.
		return t.Body.Base.
			Width(width).Height(height).
			MaxWidth(width).MaxHeight(height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Terminal demasiado pequeña")
	}

	// Vertical: header + divider + centro + divider + footer.
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
