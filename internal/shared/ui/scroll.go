package ui

import "knob/internal/shared/styles"

// ScrollOffset adjusts prev so that cursor lands inside the window, moving the
// minimum. It requires remembering the offset between renders, in exchange for
// the view not jumping when the cursor moves.
func ScrollOffset(prev, cursor, total, height int) int {
	if height <= 0 || total <= height {
		return 0
	}
	maxOffset := total - height

	offset := min(max(prev, 0), maxOffset)
	switch {
	case cursor < offset:
		offset = cursor
	case cursor >= offset+height:
		offset = cursor - height + 1
	}
	return min(max(offset, 0), maxOffset)
}

// CenteredOffset centers cursor in the window. It needs to remember nothing,
// in exchange for scrolling on every move: it is only good for short lists.
func CenteredOffset(cursor, total, height int) int {
	if height <= 0 || total <= height {
		return 0
	}
	return min(max(cursor-height/2, 0), total-height)
}

// Scrollbar returns a column of the given height showing which portion of
// total is on screen. Empty if everything fits.
func Scrollbar(t styles.Theme, total, offset, height int) []string {
	if height <= 0 || total <= height {
		return nil
	}

	// The thumb keeps the visible proportion, with a minimum of one row so it
	// does not vanish in very long lists.
	thumb := max(1, height*height/total)

	// The travel goes from 0 to height-thumb and is spread over the real range
	// of offsets. Scaling it over total would leave the thumb short of the
	// bottom.
	travel, maxOffset := height-thumb, total-height
	start := 0
	if maxOffset > 0 {
		start = min(offset*travel/maxOffset, travel)
	}

	col := make([]string, height)
	for i := range col {
		if i >= start && i < start+thumb {
			col[i] = t.Scrollbar.Thumb.Render(t.Icon.ScrollThumb)
			continue
		}
		col[i] = t.Scrollbar.Track.Render(t.Icon.ScrollTrack)
	}
	return col
}

// JoinScrollbar glues the column to the right of the rows. With no bar it pads
// with spaces: the caller reserves the slot in any case, so leaving it empty
// would throw off the width.
func JoinScrollbar(rows, bar []string) []string {
	out := make([]string, len(rows))
	for i, row := range rows {
		if i < len(bar) {
			out[i] = row + " " + bar[i]
			continue
		}
		out[i] = row + "  "
	}
	return out
}
