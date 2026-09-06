// Package audio pinta las secciones de la pantalla de sonido.
//
// No conoce el dominio: recibe Meter, un modelo de vista que arma la página.
// Así el render no depende de si detrás hay un dispositivo o un flujo, y se
// puede probar sin montar medio sistema.
package audio

import (
	"strconv"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"settings-cli/internal/shared/styles"
)

const (
	// BarSegments son los tramos de la barra. Reparten 0-100 %; por encima la
	// barra queda llena.
	BarSegments = 13
	// NominalLevel es el 100 %, el máximo sin amplificar.
	NominalLevel = 100

	// La fila es una rejilla de columnas fijas. Si alguna dependiera del
	// contenido —el nivel ocupa 3 o 4 caracteres—, la barra bailaría de una
	// fila a otra y nada quedaría alineado.
	markerColumn = 2
	gapColumn    = 2
	levelColumn  = 5 // "100%", "45%" o "mudo", a la derecha
	maxNameWidth = 38
	minNameWidth = 12
)

// Meter es una fila del medidor: lo que hay que pintar, sin saber de dónde
// viene.
type Meter struct {
	// Marker es el símbolo del indicador; vacío si no hay. Accent decide si se
	// tiñe, en vez de comparar el símbolo, que ataría el color al glifo.
	Marker string
	Accent bool

	Name  string
	Level int
	Muted bool
}

// Row compone la fila tramo a tramo.
//
// Cada trozo lleva base aplicado, incluidos los espacios: un reset ANSI
// intermedio se llevaría el fondo de la fila seleccionada de ahí en adelante.
func Row(t styles.Theme, m Meter, base lipgloss.Style, width int) string {
	level := padLeft(levelText(m), levelColumn)

	// Sin sitio para la barra se prescinde de ella: recortar más el nombre lo
	// dejaría ilegible, y una fila que desborda envuelve y descuadra el alto.
	if !fitsBar(width) {
		name := pad(truncate(t, m.Name, compactNameWidth(width)), compactNameWidth(width))
		return marker(t, m, base) + base.Render(name+spaces(gapColumn)) +
			base.Foreground(t.Color.Text).Bold(true).Render(level)
	}

	column := nameWidth(width)
	name := pad(truncate(t, m.Name, column), column)

	// El sobrante va al final y no entre columnas: así el bloque de control
	// queda anclado al nombre en vez de alejarse en terminales anchas.
	used := markerColumn + column + gapColumn + barWidth() + gapColumn + levelColumn

	return marker(t, m, base) + base.Render(name+spaces(gapColumn)) +
		bar(t, m, base) +
		base.Render(spaces(gapColumn)) +
		base.Foreground(t.Color.Text).Bold(true).Render(level) +
		base.Render(spaces(max(width-used, 0)))
}

// marker pinta el indicador y su separación. Sin símbolo deja el hueco, para
// que la rejilla no se mueva entre filas.
func marker(t styles.Theme, m Meter, base lipgloss.Style) string {
	if m.Marker == "" {
		return base.Render(spaces(markerColumn))
	}

	style := base
	if m.Accent {
		style = base.Foreground(t.Color.Accent)
	}
	return style.Render(m.Marker) + base.Render(spaces(markerColumn-1))
}

// bar dibuja los tramos llenos y vacíos. Un medidor silenciado los muestra
// todos vacíos aunque tenga volumen: importa lo que se oye, no lo configurado.
func bar(t styles.Theme, m Meter, base lipgloss.Style) string {
	filled := 0
	if !m.Muted {
		filled = min(m.Level*BarSegments/NominalLevel, BarSegments)
	}

	on := base.Foreground(t.Color.Muted)
	off := base.Foreground(t.Color.Line)

	out := ""
	for i := range BarSegments {
		if i > 0 {
			out += base.Render(" ")
		}
		if i < filled {
			out += on.Render(t.Icon.BarOn)
			continue
		}
		out += off.Render(t.Icon.BarOff)
	}
	return out
}

// levelText cabe siempre en levelColumn: "silenciado" desplazaría la rejilla.
func levelText(m Meter) string {
	if m.Muted {
		return "mudo"
	}
	return strconv.Itoa(m.Level) + "%"
}

// barWidth es lo que ocupa la barra: un carácter por tramo más los
// separadores.
func barWidth() int { return BarSegments*2 - 1 }

// fitsBar indica si cabe la rejilla completa con el nombre legible.
func fitsBar(width int) bool { return width-fixedColumns() >= minNameWidth }

// fixedColumns es todo lo que no es el nombre.
func fixedColumns() int {
	return markerColumn + gapColumn + barWidth() + gapColumn + levelColumn
}

// nameWidth es lo que queda para el nombre, con tope para que en terminales
// anchas el control no se vaya al otro extremo.
func nameWidth(width int) int {
	return min(max(width-fixedColumns(), minNameWidth), maxNameWidth)
}

// compactNameWidth es el nombre cuando no hay barra.
func compactNameWidth(width int) int {
	return max(width-markerColumn-gapColumn-levelColumn, 1)
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

func padLeft(s string, width int) string { return spaces(width-ansi.StringWidth(s)) + s }

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
