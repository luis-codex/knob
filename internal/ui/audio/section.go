package audio

import (
	"strings"

	"charm.land/lipgloss/v2"

	"settings-cli/internal/shared/components"
	"settings-cli/internal/shared/styles"
)

// Section es una lista con encabezado dentro de la pantalla de sonido.
//
// La lista se le pasa desde fuera porque guarda el cursor: la sección la pinta,
// pero no la posee.
type Section struct {
	Label   string
	Meters  []Meter
	Empty   string
	List    *components.List
	Focused bool
}

// Chrome son las filas que gasta una sección aparte de su lista: el encabezado
// y la línea en blanco que la cierra.
const Chrome = 2

// Render devuelve las filas de la sección.
func (s Section) Render(t styles.Theme, width, height int) []string {
	// La etiqueta va en mayúsculas y pasa a acento cuando tiene el foco.
	style := t.Body.Section
	if s.Focused {
		style = style.Foreground(t.Color.Accent)
	}
	heading := style.Render(strings.ToUpper(s.Label))

	if len(s.Meters) == 0 {
		return []string{heading, t.List.Empty.Render(s.Empty), ""}
	}

	rows := s.List.Render(t, len(s.Meters), width, height, s.Focused,
		func(i int, base lipgloss.Style, w int) string {
			return Row(t, s.Meters[i], base, w)
		},
	)

	return append(append([]string{heading}, rows...), "")
}
