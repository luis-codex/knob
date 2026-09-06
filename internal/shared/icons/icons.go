// Package icons gathers the interface's symbols in one place.
//
// They are the default set: the theme picks them up in styles.Icons, so
// changing them for a specific terminal does not require touching this
// package.
//
// They must all take one cell except the ones documented as double: the row
// grid counts on it.
package icons

// Status markers in a list.
const (
	// Cursor points at the row under the cursor.
	Cursor = "›"
	// Active points at the element in use: the default device.
	Active = "●"
	// Paused points at a stopped audio stream.
	Paused = "⏸"
	// Playing points at a playing stream. It blinks, so the screen turns it on
	// and off.
	Playing = "●"
)

// Layout and bar strokes.
const (
	BarOn       = "▰"
	BarOff      = "·"
	DividerH    = "─"
	DividerV    = "│"
	ScrollThumb = "┃"
	ScrollTrack = "│"
	Ellipsis    = "…"
)

// Keys. Written with the symbol when one exists and with the name when not:
// "esc" is easier to recognize than any glyph.
const (
	// Enter takes one cell.
	Enter = "↵"
	// UpDown and LeftRight take two.
	UpDown    = "↑↓"
	LeftRight = "←→"
	Tab       = "⇥"
	Escape    = "esc"
	Space     = "space"
	Delete    = "del"
)

// Separator goes between one keyboard hint and the next.
const Separator = "·"
