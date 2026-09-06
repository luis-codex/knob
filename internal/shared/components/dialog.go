package components

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"settings-cli/internal/shared/styles"
	"settings-cli/internal/shared/ui"
)

// DialogAction es lo que una pulsación provoca en el diálogo.
type DialogAction int

const (
	// DialogNone: el diálogo sigue abierto, no hay nada que aplicar.
	DialogNone DialogAction = iota
	// DialogAccept: el usuario aceptó (botón primario).
	DialogAccept
	// DialogDismiss: el usuario canceló (botón secundario o esc).
	DialogDismiss
)

// Dialog es un modal de confirmación (mensaje) o de formulario (campo de
// texto). Es un solo tipo porque comparten marco, foco y teclado: lo único
// que cambia es una fila.
type Dialog struct {
	title   string
	message string

	// hasInput distingue las dos formas. Por valor, para que copiar el
	// Dialog no comparta buffer.
	hasInput    bool
	label       string
	placeholder string
	input       TextInput

	accept  string
	dismiss string
	// errMsg es un fallo de validación; se muestra dentro sin cerrar.
	errMsg string
	// dangerous pinta el botón primario como destructivo.
	dangerous bool

	focus int // 0 = accept, 1 = dismiss
	open  bool
}

// NewConfirmDialog crea un diálogo de confirmación.
func NewConfirmDialog(title, message, accept string) Dialog {
	return Dialog{
		title:   title,
		message: message,
		accept:  accept,
		dismiss: "Cancelar",
	}
}

// NewFormDialog crea un diálogo con un campo de texto.
func NewFormDialog(title, label, placeholder, accept string) Dialog {
	return Dialog{
		title:       title,
		label:       label,
		placeholder: placeholder,
		hasInput:    true,
		accept:      accept,
		dismiss:     "Cancelar",
	}
}

// Dangerous marca la acción primaria como destructiva.
func (d Dialog) Dangerous() Dialog {
	d.dangerous = true
	return d
}

// WithValue precarga el campo de texto.
func (d Dialog) WithValue(value string) Dialog {
	d.input = NewTextInput(value)
	return d
}

// Open abre el diálogo con el foco en el botón primario.
func (d Dialog) Open() Dialog {
	d.open, d.focus, d.errMsg = true, 0, ""
	return d
}

// WithError muestra un fallo dentro del diálogo, que permanece abierto.
func (d Dialog) WithError(msg string) Dialog {
	d.errMsg = msg
	return d
}

func (d Dialog) Close() Dialog {
	d.open = false
	return d
}

func (d Dialog) IsOpen() bool { return d.open }

// Value es el texto del campo, ya recortado.
func (d Dialog) Value() string { return d.input.Trimmed() }

// HandleKey procesa una pulsación y devuelve el diálogo actualizado y la
// acción resultante. Consume todas las teclas mientras esté abierto.
func (d Dialog) HandleKey(msg tea.KeyPressMsg) (Dialog, DialogAction) {
	d.errMsg = "" // cualquier pulsación descarta el fallo anterior

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

	// Con campo, las flechas y letras son escritura y el foco solo se mueve
	// con tab. Sin campo, cualquier movimiento lateral cambia de botón.
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
	d.focus = (d.focus + 1) % 2 // con dos botones coincide con focusNext
	return d
}

// View recibe el área disponible, no su propio tamaño: el diálogo decide
// cuánto ocupa y quien lo compone lo centra.
func (d Dialog) View(t styles.Theme, width, height int) string {
	// En lipgloss v2, Width(n) es el ancho TOTAL: borde y padding van dentro.
	// Dimensionar el contenido al ancho externo hace que el marco lo envuelva.
	box := min(styles.DialogWidth, width-4)
	content := box - t.Dialog.Box.GetHorizontalFrameSize()
	if content < 16 {
		return ""
	}

	// Filas que caben dentro del marco. Un diálogo más alto que su hueco
	// haría crecer el layout en vez de quedarse dentro.
	maxRows := height - 2 - t.Dialog.Box.GetVerticalFrameSize()
	if maxRows < 3 {
		return ""
	}

	// Cada tramo por separado: envolver todo en un estilo hace que el reset
	// de un tramo interno se coma el color del resto.
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

	// La ayuda es lo primero que sobra cuando falta alto: sin ella el diálogo
	// sigue siendo usable, sin los botones no.
	hint := []string{
		t.Dialog.Text.Width(content).Render(""),
		// MaxWidth: una ayuda que envuelve estira la caja y descuadra el modal.
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

// trimBlanks quita líneas en blanco intermedias mientras sobren filas. Se
// aprieta el diálogo antes que renunciar a mostrarlo.
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

// contentRows es la única diferencia real entre las dos formas del diálogo.
func (d Dialog) contentRows(t styles.Theme, width int) []string {
	if !d.hasInput {
		return []string{t.Dialog.Text.Width(width).Render(d.message)}
	}

	// La etiqueta va sobre el campo para que este ocupe el ancho completo.
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
		return "tab · enter aceptar · esc cancelar"
	}
	return "←→ · enter aceptar · esc cancelar"
}
