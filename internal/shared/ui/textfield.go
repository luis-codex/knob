package ui

import (
	"github.com/charmbracelet/x/ansi"

	"knob/internal/shared/styles"
)

// TextFieldOpts configures a text field. The buffer and focus are kept by
// whoever draws it.
type TextFieldOpts struct {
	Label       string
	Value       string
	Placeholder string
	Focused     bool
	// Width is the field's total width, label included.
	Width int
}

// TextField renders a single-line input. The cursor only appears when focused,
// to tell it apart from an inactive field with text.
func TextField(t styles.Theme, opts TextFieldOpts) string {
	label := ""
	if opts.Label != "" {
		label = t.Input.Label.Render(opts.Label + " ")
	}

	box := opts.Width - ansi.StringWidth(label)
	if box < 4 {
		return label
	}

	style := t.Input.Blurred
	if opts.Focused {
		style = t.Input.Focused
	}

	// The cursor takes one cell, so the text is clipped to box-1.
	text, cursor := opts.Value, ""
	if opts.Focused {
		cursor = t.Input.Cursor.Render(" ")
	}
	textWidth := box - ansi.StringWidth(ansi.Strip(cursor))

	// Also when focused: an empty field with no hint does not say what is
	// expected.
	if text == "" && opts.Placeholder != "" {
		return label + style.Width(box).Render(
			cursor+t.Input.Placeholder.Render(fitLeft(t, opts.Placeholder, textWidth)),
		)
	}

	// If it does not fit, the tail is shown, where the user is typing.
	return label + style.Width(box).Render(tail(text, textWidth)+cursor)
}

// tail returns the last width columns of s.
func tail(s string, width int) string {
	if width <= 0 {
		return ""
	}
	w := ansi.StringWidth(s)
	if w <= width {
		return s
	}
	return ansi.TruncateLeft(s, w-width, "")
}

func fitLeft(t styles.Theme, s string, width int) string {
	if width <= 0 {
		return ""
	}
	if ansi.StringWidth(s) <= width {
		return s
	}
	return ansi.Truncate(s, width, t.Icon.Ellipsis)
}
