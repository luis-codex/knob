// Package bluetooth pinta las secciones de la pantalla de Bluetooth.
//
// No conoce el dominio: recibe modelos de vista que arma la página, así que
// el render no depende de cómo esté modelado un dispositivo.
package bluetooth

import (
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"settings-cli/internal/shared/components"
	"settings-cli/internal/shared/styles"
)

// nameColumn es el ancho reservado al nombre; el resto queda para tipo y
// estado.
const nameColumn = 22

// Device es una fila de la lista: lo que se enseña de un dispositivo.
type Device struct {
	Name   string
	Kind   string
	Status string
}

// Row compone nombre, tipo y estado en una línea. Va sin estilos propios:
// components.List tiñe la fila entera y un color interno rompería el fondo de
// la seleccionada.
func Row(t styles.Theme, d Device, width int) string {
	name := pad(truncate(t, d.Name, nameColumn), nameColumn)

	gap := width - ansi.StringWidth(name) - ansi.StringWidth(d.Kind) - ansi.StringWidth(d.Status)
	return name + d.Kind + spaces(max(gap, 1)) + d.Status
}

// Adapter es el interruptor de la radio, alineado a los extremos.
func Adapter(t styles.Theme, enabled bool, width int) string {
	const label = "Bluetooth"

	state, style := "desactivado", t.Body.Muted
	if enabled {
		state, style = "activado", t.Body.Badge
	}

	gap := width - ansi.StringWidth(label) - ansi.StringWidth(state)
	return t.Body.Label.Render(label) + spaces(max(gap, 1)) + style.Render(state)
}

// Rows dibuja la lista con el cursor que le presta la página: la sección
// pinta, pero el cursor no es suyo.
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
