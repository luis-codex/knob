package components

import (
	"settings-cli/internal/shared/styles"
	"settings-cli/internal/shared/ui"
)

// Footer shows the keys available in the current context.
type Footer struct {
	Keys []ui.Key
}

func NewFooter(keys ...ui.Key) Footer {
	return Footer{Keys: keys}
}

func (c Footer) View(t styles.Theme, width, height int) string {
	return t.Footer.Base.Width(width).Height(height).MaxHeight(height).
		MaxWidth(width).Render(ui.Hints(t.Footer.Hint, c.Keys...))
}
