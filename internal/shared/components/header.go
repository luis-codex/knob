package components

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"settings-cli/internal/shared/styles"
)

// Header es la barra superior: el nombre de la aplicación y, a la derecha, la
// sección en la que estás.
type Header struct {
	Name    string
	Section string
}

func NewHeader() Header {
	return Header{Name: "settings"}
}

// WithSection devuelve una copia indicando dónde está el usuario.
func (c Header) WithSection(section string) Header {
	c.Section = section
	return c
}

func (c Header) View(t styles.Theme, width, height int) string {
	name := t.Header.Title.Render(c.Name)
	inner := width - t.Header.Base.GetHorizontalFrameSize()

	line := name
	if c.Section != "" {
		crumb := t.Header.Crumb.Render(c.Section)
		gap := inner - ansi.StringWidth(name) - ansi.StringWidth(crumb)
		if gap >= 1 {
			line = name + strings.Repeat(" ", gap) + crumb
		}
	}

	return t.Header.Base.Width(width).Height(height).MaxHeight(height).Render(line)
}
