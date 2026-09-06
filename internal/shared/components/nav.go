package components

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"settings-cli/internal/shared/styles"
	"settings-cli/internal/shared/ui"
)

// NavItem es una hoja del árbol y selecciona una página.
type NavItem struct {
	ID    string
	Label string
}

// NavGroup es un encabezado con sus sub-items. No es seleccionable.
type NavGroup struct {
	Label string
	Items []NavItem
}

// Nav es el árbol de navegación del sidebar.
type Nav struct {
	Groups []NavGroup
	// Focused indica si el sidebar tiene el foco del teclado.
	Focused bool
	cursor  int // índice sobre la lista aplanada de items
}

func NewNav(groups ...NavGroup) Nav {
	return Nav{Groups: groups, Focused: true}
}

// WithFocus devuelve una copia con el foco puesto o quitado.
func (n Nav) WithFocus(focused bool) Nav {
	n.Focused = focused
	return n
}

// items aplana los sub-items de todos los grupos en orden de pantalla.
func (n Nav) items() []NavItem {
	var out []NavItem
	for _, g := range n.Groups {
		out = append(out, g.Items...)
	}
	return out
}

// Selected devuelve el item bajo el cursor, o un NavItem vacío si no hay.
func (n Nav) Selected() NavItem {
	items := n.items()
	if len(items) == 0 {
		return NavItem{}
	}
	return items[clamp(n.cursor, 0, len(items)-1)]
}

func (n Nav) Next() Nav {
	if last := len(n.items()) - 1; n.cursor < last {
		n.cursor++
	}
	return n
}

func (n Nav) Prev() Nav {
	if n.cursor > 0 {
		n.cursor--
	}
	return n
}

func (n Nav) View(t styles.Theme, width, height int) string {
	inner := width - t.Nav.Base.GetHorizontalFrameSize()
	if inner <= 0 {
		return ""
	}

	lines, cursorLine := n.lines(t, inner)

	// La ventana recorta líneas, pero el cursor indexa items: por eso lines
	// devuelve en qué línea cayó el seleccionado.
	if offset := ui.CenteredOffset(cursorLine, len(lines), height); offset > 0 {
		lines = lines[offset:min(offset+height, len(lines))]
	}

	// Height rellena pero no recorta; sin MaxHeight el menú desbordaría.
	return t.Nav.Base.Width(width).Height(height).MaxHeight(height).Render(
		lipgloss.JoinVertical(lipgloss.Left, lines...),
	)
}

// lines renderiza el menú completo y devuelve la línea del item seleccionado.
func (n Nav) lines(t styles.Theme, width int) (lines []string, cursorLine int) {
	item := 0 // índice sobre la lista aplanada
	for gi, g := range n.Groups {
		if gi > 0 {
			lines = append(lines, "")
		}
		// Mayúsculas: son etiquetas de agrupación, no contenido.
		lines = append(lines, t.Nav.Group.Render(fit(t, strings.ToUpper(g.Label), width)))

		for _, entry := range g.Items {
			style, label := t.Nav.Item, "  "+entry.Label
			if item == n.cursor {
				cursorLine = len(lines)
				style, label = t.Nav.Selected, t.Icon.Cursor+" "+entry.Label
				if !n.Focused {
					style = t.Nav.Blurred
				}
			}
			// Width en el estilo para que el fondo llegue hasta el borde.
			lines = append(lines, style.Width(width).Render(fit(t, label, width)))
			item++
		}
	}
	return lines, cursorLine
}

// fit trunca respetando los códigos ANSI, que cortados por byte se romperían.
func fit(t styles.Theme, s string, width int) string {
	if width <= 0 {
		return ""
	}
	if ansi.StringWidth(s) <= width {
		return s
	}
	return ansi.Truncate(s, width, t.Icon.Ellipsis)
}

func clamp(v, lo, hi int) int {
	return min(max(v, lo), hi)
}
