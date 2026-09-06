package pages

import (
	"settings-cli/internal/shared/styles"
)

// wifiNote avisa de que la pantalla aún no gestiona nada. Se quita cuando
// exista el adaptador contra el gestor de red.
const wifiNote = "En desarrollo · aún no gestiona redes"

// WiFi es la pantalla de redes. De momento solo informa: falta el puerto
// contra el gestor de red del sistema.
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
		t.Body.Muted.Render("Hará falta un adaptador contra NetworkManager o iwd,"),
		t.Body.Muted.Render("igual que bluez lo es para Bluetooth."),
	)
}
