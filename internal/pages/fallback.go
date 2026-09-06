package pages

import "settings-cli/internal/shared/styles"

// Fallback is not a menu entry: it covers an unknown navigation identifier so
// the layout never gets a nil page.
type Fallback struct {
	title string
}

func NewFallback(title string) Fallback {
	return Fallback{title: title}
}

func (p Fallback) View(t styles.Theme, width, height int) string {
	return frame(t, width, height, p.title,
		t.Body.Muted.Render("This section does not exist."),
	)
}
