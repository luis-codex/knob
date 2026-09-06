package components

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"settings-cli/internal/shared/styles"
	"settings-cli/internal/shared/ui"
)

// DialogAction is what a keypress causes in the dialog.
type DialogAction int

const (
	// DialogNone: the dialog stays open, there is nothing to apply.
	DialogNone DialogAction = iota
	// DialogAccept: the user accepted (primary button).
	DialogAccept
	// DialogDismiss: the user cancelled (secondary button or esc).
	DialogDismiss
)

// Dialog is a confirmation modal (message) or a form modal (text field). It is
// a single type because they share frame, focus and keyboard: the only thing
// that changes is one row.
type Dialog struct {
	title   string
	message string

	// hasInput tells the two forms apart. By value, so copying the Dialog
	// does not share the buffer.
	hasInput    bool
	label       string
	placeholder string
	input       TextInput

	accept  string
	dismiss string
	// errMsg is a validation failure; shown inside without closing.
	errMsg string
	// dangerous renders the primary button as destructive.
	dangerous bool

	focus int // 0 = accept, 1 = dismiss
	open  bool
}

// NewConfirmDialog creates a confirmation dialog.
func NewConfirmDialog(title, message, accept string) Dialog {
	return Dialog{
		title:   title,
		message: message,
		accept:  accept,
		dismiss: "Cancel",
	}
}

// NewFormDialog creates a dialog with a text field.
func NewFormDialog(title, label, placeholder, accept string) Dialog {
	return Dialog{
		title:       title,
		label:       label,
		placeholder: placeholder,
		hasInput:    true,
		accept:      accept,
		dismiss:     "Cancel",
	}
}

// Dangerous marks the primary action as destructive.
func (d Dialog) Dangerous() Dialog {
	d.dangerous = true
	return d
}

// WithValue pre-fills the text field.
func (d Dialog) WithValue(value string) Dialog {
	d.input = NewTextInput(value)
	return d
}

// Open opens the dialog with focus on the primary button.
func (d Dialog) Open() Dialog {
	d.open, d.focus, d.errMsg = true, 0, ""
	return d
}

// WithError shows a failure inside the dialog, which stays open.
func (d Dialog) WithError(msg string) Dialog {
	d.errMsg = msg
	return d
}

func (d Dialog) Close() Dialog {
	d.open = false
	return d
}

func (d Dialog) IsOpen() bool { return d.open }

// Value is the field's text, already trimmed.
func (d Dialog) Value() string { return d.input.Trimmed() }

// HandleKey processes a keypress and returns the updated dialog and the
// resulting action. It consumes every key while open.
func (d Dialog) HandleKey(msg tea.KeyPressMsg) (Dialog, DialogAction) {
	d.errMsg = "" // any keypress dismisses the previous failure

	switch msg.String() {
	case "esc":
		return d, DialogDismiss

	case "enter":
		if d.focus == 0 {
			return d, DialogAccept
		}
		return d, DialogDismiss

	case "tab":
		return d.focusNext(), DialogNone

	case "shift+tab":
		return d.focusPrev(), DialogNone
	}

	// With a field, arrows and letters are typing and focus only moves with
	// tab. Without a field, any sideways move switches button.
	if d.hasInput {
		input, typed := d.input.TypeKey(msg)
		if typed {
			d.input = input
		}
		return d, DialogNone
	}

	switch msg.String() {
	case "right", "l":
		return d.focusNext(), DialogNone
	case "left", "h":
		return d.focusPrev(), DialogNone
	}
	return d, DialogNone
}

func (d Dialog) focusNext() Dialog {
	d.focus = (d.focus + 1) % 2
	return d
}

func (d Dialog) focusPrev() Dialog {
	d.focus = (d.focus + 1) % 2 // with two buttons this matches focusNext
	return d
}

// View receives the available area, not its own size: the dialog decides how
// much it takes and whoever composes it centers it.
func (d Dialog) View(t styles.Theme, width, height int) string {
	// In lipgloss v2, Width(n) is the TOTAL width: border and padding go
	// inside. Sizing the content to the outer width makes the frame wrap it.
	box := min(styles.DialogWidth, width-4)
	content := box - t.Dialog.Box.GetHorizontalFrameSize()
	if content < 16 {
		return ""
	}

	// Rows that fit inside the frame. A dialog taller than its slot would grow
	// the layout instead of staying inside.
	maxRows := height - 2 - t.Dialog.Box.GetVerticalFrameSize()
	if maxRows < 3 {
		return ""
	}

	// Each span separately: wrapping everything in one style makes an inner
	// span's reset eat the rest's color.
	rows := []string{
		t.Dialog.Title.Width(content).Render(d.title),
		t.Dialog.Text.Width(content).Render(""),
	}
	rows = append(rows, d.contentRows(t, content)...)
	if d.errMsg != "" {
		rows = append(rows, t.Dialog.Error.Width(content).Render(d.errMsg))
	}
	rows = append(rows,
		t.Dialog.Text.Width(content).Render(""),
		lipgloss.NewStyle().Width(content).Align(lipgloss.Center).Render(d.buttons(t)),
	)

	// The help is the first thing to drop when height is short: without it the
	// dialog is still usable, without the buttons it is not.
	hint := []string{
		t.Dialog.Text.Width(content).Render(""),
		// MaxWidth: a wrapping hint stretches the box and throws off the modal.
		t.Dialog.Hint.Width(content).MaxWidth(content).Align(lipgloss.Center).Render(d.hint()),
	}
	if len(rows)+len(hint) <= maxRows {
		rows = append(rows, hint...)
	}

	rows = trimBlanks(rows, maxRows)
	if len(rows) > maxRows {
		return ""
	}

	return t.Dialog.Box.Width(box).Render(
		lipgloss.JoinVertical(lipgloss.Left, rows...),
	)
}

// trimBlanks removes mid blank lines while there are extra rows. The dialog is
// squeezed before giving up on showing it.
func trimBlanks(rows []string, maxRows int) []string {
	for len(rows) > maxRows {
		removed := false
		for i := len(rows) - 2; i > 0; i-- {
			if strings.TrimSpace(ansi.Strip(rows[i])) != "" {
				continue
			}
			rows = append(rows[:i], rows[i+1:]...)
			removed = true
			break
		}
		if !removed {
			return rows
		}
	}
	return rows
}

// contentRows is the only real difference between the two forms of the dialog.
func (d Dialog) contentRows(t styles.Theme, width int) []string {
	if !d.hasInput {
		return []string{t.Dialog.Text.Width(width).Render(d.message)}
	}

	// The label goes above the field so the field takes the full width.
	return []string{
		t.Dialog.Label.Width(width).Render(d.label),
		ui.TextField(t, ui.TextFieldOpts{
			Value:       d.input.Value(),
			Placeholder: d.placeholder,
			Focused:     true,
			Width:       width,
		}),
	}
}

func (d Dialog) buttons(t styles.Theme) string {
	primary := ui.ButtonOpts{Text: d.accept, Focused: d.focus == 0, Danger: d.dangerous}
	secondary := ui.ButtonOpts{Text: d.dismiss, Focused: d.focus == 1}
	return ui.ButtonGroup(t, []ui.ButtonOpts{primary, secondary}, "  ")
}

func (d Dialog) hint() string {
	if d.hasInput {
		return "tab · enter accept · esc cancel"
	}
	return "←→ · enter accept · esc cancel"
}
