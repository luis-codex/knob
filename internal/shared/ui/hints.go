package ui

import (
	"strings"

	"knob/internal/shared/icons"
	"knob/internal/shared/styles"
)

// Key is a keyboard hint: the key and what it does.
type Key struct {
	Name   string
	Action string
}

// Hints renders a list of hints.
//
// The key is rendered apart from its description because they are different
// things: someone who does not know how to quit looks for the key, not the
// phrase.
func Hints(s styles.HintStyles, keys ...Key) string {
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, s.Key.Render(k.Name)+" "+s.Action.Render(k.Action))
	}
	return strings.Join(parts, s.Sep.Render("  "+icons.Separator+"  "))
}
