package ui

import (
	"strings"

	"settings-cli/internal/shared/icons"
	"settings-cli/internal/shared/styles"
)

// Key es una pista de teclado: la tecla y lo que hace.
type Key struct {
	Name   string
	Action string
}

// Hints pinta una lista de pistas.
//
// La tecla se pinta aparte de su descripción porque son cosas distintas: quien
// no sepa cómo salir busca la tecla, no la frase.
func Hints(s styles.HintStyles, keys ...Key) string {
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, s.Key.Render(k.Name)+" "+s.Action.Render(k.Action))
	}
	return strings.Join(parts, s.Sep.Render("  "+icons.Separator+"  "))
}
