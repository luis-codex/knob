package components

import (
	"charm.land/lipgloss/v2"

	"settings-cli/internal/shared/styles"
	"settings-cli/internal/shared/ui"
)

// List is a navigable, scrolling list. It keeps the cursor and offset; the row
// text is formatted by whoever uses it.
//
// It is a pointer and not a value: Render adjusts the offset, and it is the
// only place that knows the available height.
type List struct {
	cursor int
	offset int
	count  int
}

func NewList() *List { return &List{} }

func (l *List) Cursor() int { return l.cursor }

func (l *List) Next() {
	l.cursor++
	l.clampCursor()
}

func (l *List) Prev() {
	l.cursor--
	l.clampCursor()
}

// SetCount re-adjusts the cursor after the data changes. Mandatory when
// filtering or deleting, or the cursor points to an index that no longer
// exists.
func (l *List) SetCount(n int) {
	l.count = n
	l.clampCursor()
}

func (l *List) clampCursor() {
	if l.count <= 0 {
		l.cursor = 0
		return
	}
	l.cursor = clamp(l.cursor, 0, l.count-1)
}

// RowFunc formats a row's content.
//
// It receives the style already resolved (normal or selected) and the exact
// width it must occupy. Whoever colors spans on their own must apply base to
// all of them: a mid-string ANSI reset takes the rest's background with it.
type RowFunc func(index int, base lipgloss.Style, width int) string

// PlainRows adapts plain, uncolored text rows, which is the common case.
func PlainRows(t styles.Theme, rows []string) RowFunc {
	return func(i int, base lipgloss.Style, width int) string {
		return base.Width(width).Render(fit(t, rows[i], width))
	}
}

// Render returns the visible rows with their scrollbar. showCursor turns off
// the highlight when focus is elsewhere.
func (l *List) Render(t styles.Theme, count, width, height int, showCursor bool, row RowFunc) []string {
	l.SetCount(count)
	if count == 0 || height <= 0 || width <= 0 {
		return nil
	}

	l.offset = ui.ScrollOffset(l.offset, l.cursor, count, height)
	bar := ui.Scrollbar(t, count, l.offset, height)

	// The scrollbar slot is always reserved, bar or no bar: if it depended on
	// the list overflowing, the rows would change width as it grows and two
	// adjacent lists would not line up.
	const scrollbarColumns = 2
	itemWidth := width - scrollbarColumns

	const prefixWidth = 2

	end := min(l.offset+height, count)
	out := make([]string, 0, end-l.offset)
	for i := l.offset; i < end; i++ {
		base, prefix := t.List.Item, "  "
		if i == l.cursor && showCursor {
			base, prefix = t.List.Selected, t.Icon.Cursor+" "
		}
		out = append(out, base.Render(prefix)+row(i, base, itemWidth-prefixWidth))
	}

	return ui.JoinScrollbar(out, bar)
}
