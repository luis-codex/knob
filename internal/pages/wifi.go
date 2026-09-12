package pages

import (
	"knob/internal/shared/styles"
)

// wifiNote warns that the screen does not manage anything yet. It goes away
// when the adapter against the network manager exists.
const wifiNote = "In development · does not manage networks yet"

// WiFi is the network screen. For now it only informs: the port against the
// system's network manager is missing.
type WiFi struct {
	title string
}

func NewWiFi(title string) WiFi {
	return WiFi{title: title}
}

func (p WiFi) View(t styles.Theme, width, height int) string {
	return frame(t, width, height, p.title,
		t.Body.Note.Render(wifiNote),
		"",
		t.Body.Muted.Render("It will need an adapter against NetworkManager or iwd."),
	)
}
