package ui

import "settings-cli/internal/shared/styles"

// ScrollOffset ajusta prev para que cursor quede dentro de la ventana,
// desplazando lo mínimo. Requiere recordar el offset entre renders, a cambio
// de que la vista no salte al mover el cursor.
func ScrollOffset(prev, cursor, total, height int) int {
	if height <= 0 || total <= height {
		return 0
	}
	maxOffset := total - height

	offset := min(max(prev, 0), maxOffset)
	switch {
	case cursor < offset:
		offset = cursor
	case cursor >= offset+height:
		offset = cursor - height + 1
	}
	return min(max(offset, 0), maxOffset)
}

// CenteredOffset centra cursor en la ventana. No necesita recordar nada, a
// cambio de desplazar en cada movimiento: solo vale para listas cortas.
func CenteredOffset(cursor, total, height int) int {
	if height <= 0 || total <= height {
		return 0
	}
	return min(max(cursor-height/2, 0), total-height)
}

// Scrollbar devuelve una columna de height filas indicando qué porción de
// total se está viendo. Vacía si cabe todo.
func Scrollbar(t styles.Theme, total, offset, height int) []string {
	if height <= 0 || total <= height {
		return nil
	}

	// El pulgar guarda la proporción visible, con un mínimo de una fila para
	// que no desaparezca en listas muy largas.
	thumb := max(1, height*height/total)

	// El recorrido va de 0 a height-thumb y se reparte sobre el rango real de
	// offsets. Escalarlo sobre total dejaría el pulgar sin llegar al fondo.
	travel, maxOffset := height-thumb, total-height
	start := 0
	if maxOffset > 0 {
		start = min(offset*travel/maxOffset, travel)
	}

	col := make([]string, height)
	for i := range col {
		if i >= start && i < start+thumb {
			col[i] = t.Scrollbar.Thumb.Render(t.Icon.ScrollThumb)
			continue
		}
		col[i] = t.Scrollbar.Track.Render(t.Icon.ScrollTrack)
	}
	return col
}

// JoinScrollbar pega la columna a la derecha de las filas. Sin barra rellena
// con espacios: el hueco lo reserva el llamante en todo caso, así que dejarlo
// vacío descuadraría el ancho.
func JoinScrollbar(rows, bar []string) []string {
	out := make([]string, len(rows))
	for i, row := range rows {
		if i < len(bar) {
			out[i] = row + " " + bar[i]
			continue
		}
		out[i] = row + "  "
	}
	return out
}
