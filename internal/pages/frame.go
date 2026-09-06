// Package pages contiene las pantallas de la app: un fichero por entrada del
// menú. Cada página compone sus secciones desde internal/ui.
//
// Este fichero no es una página: son el marco común y los helpers de texto que
// comparten todas.
package pages

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"settings-cli/internal/shared/styles"
)

// frameChrome son las filas que frame añade por su cuenta: el título, la
// regla que lo separa y la línea en blanco. Las páginas lo restan para medir
// cuánto alto le queda a su contenido.
const frameChrome = 3

// frame renderiza el esqueleto común: título, separación y cuerpo.
func frame(t styles.Theme, width, height int, title string, rows ...string) string {
	inner := width - t.Body.Base.GetHorizontalFrameSize()
	if inner <= 0 {
		return ""
	}

	// Una regla bajo el título separa la cabecera del contenido, como el
	// borde superior de las tarjetas de bun.com. Sin ella la página es un
	// bloque plano de texto.
	lines := append([]string{
		t.Body.Title.Render(title),
		t.Divider.Render(strings.Repeat(t.Icon.DividerH, max(inner, 0))),
		"",
	}, rows...)

	// Height rellena pero no recorta; sin MaxHeight el body desbordaría.
	return t.Body.Base.Width(width).Height(height).MaxHeight(height).Render(
		lipgloss.JoinVertical(lipgloss.Left, lines...),
	)
}

// fit recorta s a width columnas respetando los códigos ANSI.
func fit(s string, width int) string {
	if width <= 0 {
		return ""
	}
	return lipgloss.NewStyle().MaxWidth(width).Render(s)
}

// ellipsis recorta s a width columnas con el símbolo del tema, para cortes
// que lee una persona. Para rellenar filas basta con fit.
func ellipsis(t styles.Theme, s string, width int) string {
	if width <= 0 {
		return ""
	}
	if ansi.StringWidth(s) <= width {
		return s
	}
	return ansi.Truncate(s, width, t.Icon.Ellipsis)
}

// --- texto ------------------------------------------------------------------

// spaces devuelve n espacios. Rellena columnas de una rejilla, donde el hueco
// debe existir aunque no haya contenido.
func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(" ", n)
}

// pad y padLeft ajustan s a width alineándolo a un lado o al otro.
func pad(s string, width int) string {
	return s + spaces(width-ansi.StringWidth(s))
}

func padLeft(s string, width int) string {
	return spaces(width-ansi.StringWidth(s)) + s
}

// plural formatea un recuento con su sustantivo.
func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return strconv.Itoa(n) + " " + many
}
