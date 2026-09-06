package components

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// TextInput is a single-line text buffer. State only: the drawing is done by
// ui.TextField.
type TextInput struct {
	value string
}

func NewTextInput(value string) TextInput {
	return TextInput{value: value}
}

func (i TextInput) Value() string { return i.value }

// Trimmed is the value with leading and trailing spaces removed.
func (i TextInput) Trimmed() string { return strings.TrimSpace(i.value) }

func (i TextInput) Insert(s string) TextInput {
	i.value += s
	return i
}

func (i TextInput) Backspace() TextInput {
	if r := []rune(i.value); len(r) > 0 {
		i.value = string(r[:len(r)-1])
	}
	return i
}

func (i TextInput) Clear() TextInput {
	i.value = ""
	return i
}

// TypeKey applies a keypress to the buffer. false means it was not text and
// the caller must treat it as a command.
func (i TextInput) TypeKey(msg tea.KeyPressMsg) (TextInput, bool) {
	if msg.String() == "backspace" {
		return i.Backspace(), true
	}

	// Text only carries content on printable keys. Those with a modifier are
	// dropped so a shortcut does not end up typed.
	if msg.Text != "" && msg.Mod&(tea.ModCtrl|tea.ModAlt|tea.ModMeta) == 0 {
		return i.Insert(msg.Text), true
	}

	return i, false
}
