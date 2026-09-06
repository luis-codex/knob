package pages

import "settings-cli/internal/shared/styles"

// Fallback no es una entrada del menú: cubre un identificador de navegación
// desconocido, para que el layout nunca reciba una página nula.
type Fallback struct {
	title string
}

func NewFallback(title string) Fallback {
	return Fallback{title: title}
}

func (p Fallback) View(t styles.Theme, width, height int) string {
	return frame(t, width, height, p.title,
		t.Body.Muted.Render("Esta sección no existe."),
	)
}
