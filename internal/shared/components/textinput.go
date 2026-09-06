package components

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// TextInput es un buffer de texto de una línea. Solo estado: el dibujo lo
// hace ui.TextField.
type TextInput struct {
	value string
}

func NewTextInput(value string) TextInput {
	return TextInput{value: value}
}

func (i TextInput) Value() string { return i.value }

// Trimmed es el valor sin espacios en los extremos.
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

// TypeKey aplica una pulsación al buffer. false significa que no era texto y
// el llamante debe tratarla como comando.
func (i TextInput) TypeKey(msg tea.KeyPressMsg) (TextInput, bool) {
	if msg.String() == "backspace" {
		return i.Backspace(), true
	}

	// Text solo trae contenido en teclas imprimibles. Se descartan las que
	// llevan modificador para que un atajo no acabe escrito.
	if msg.Text != "" && msg.Mod&(tea.ModCtrl|tea.ModAlt|tea.ModMeta) == 0 {
		return i.Insert(msg.Text), true
	}

	return i, false
}
