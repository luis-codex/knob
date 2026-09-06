package components

import (
	"charm.land/lipgloss/v2"

	"settings-cli/internal/shared/styles"
	"settings-cli/internal/shared/ui"
)

// List es una lista navegable con scroll. Guarda cursor y desplazamiento; el
// texto de las filas lo formatea quien la usa.
//
// Es un puntero y no un valor: Render ajusta el desplazamiento, y es el único
// punto que conoce la altura disponible.
type List struct {
	cursor int
	offset int
	count  int
}

func NewList() *List { return &List{} }

func (l *List) Cursor() int { return l.cursor }

func (l *List) Next() {
	l.cursor++
	l.clampCursor()
}

func (l *List) Prev() {
	l.cursor--
	l.clampCursor()
}

// SetCount reajusta el cursor tras cambiar los datos. Obligatorio al filtrar
// o borrar, o el cursor apunta a un índice que ya no existe.
func (l *List) SetCount(n int) {
	l.count = n
	l.clampCursor()
}

func (l *List) clampCursor() {
	if l.count <= 0 {
		l.cursor = 0
		return
	}
	l.cursor = clamp(l.cursor, 0, l.count-1)
}

// RowFunc formatea el contenido de una fila.
//
// Recibe el estilo ya resuelto (normal o seleccionada) y el ancho exacto que
// debe ocupar. Quien coloree tramos por su cuenta tiene que aplicar base a
// todos: un reset ANSI intermedio se lleva por delante el fondo del resto.
type RowFunc func(index int, base lipgloss.Style, width int) string

// PlainRows adapta filas de texto sin color, que es el caso habitual.
func PlainRows(t styles.Theme, rows []string) RowFunc {
	return func(i int, base lipgloss.Style, width int) string {
		return base.Width(width).Render(fit(t, rows[i], width))
	}
}

// Render devuelve las filas visibles con su barra de scroll. showCursor apaga
// el resaltado cuando el foco está en otro sitio.
func (l *List) Render(t styles.Theme, count, width, height int, showCursor bool, row RowFunc) []string {
	l.SetCount(count)
	if count == 0 || height <= 0 || width <= 0 {
		return nil
	}

	l.offset = ui.ScrollOffset(l.offset, l.cursor, count, height)
	bar := ui.Scrollbar(t, count, l.offset, height)

	// El hueco de la barra se reserva siempre, haya barra o no: si dependiera
	// de que la lista desborde, las filas cambiarían de ancho al crecer y dos
	// listas contiguas no quedarían alineadas.
	const scrollbarColumns = 2
	itemWidth := width - scrollbarColumns

	const prefixWidth = 2

	end := min(l.offset+height, count)
	out := make([]string, 0, end-l.offset)
	for i := l.offset; i < end; i++ {
		base, prefix := t.List.Item, "  "
		if i == l.cursor && showCursor {
			base, prefix = t.List.Selected, t.Icon.Cursor+" "
		}
		out = append(out, base.Render(prefix)+row(i, base, itemWidth-prefixWidth))
	}

	return ui.JoinScrollbar(out, bar)
}
