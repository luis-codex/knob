package components

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"settings-cli/internal/shared/styles"
	"settings-cli/internal/shared/ui"
)

// NavItem is a leaf of the tree and selects a page.
type NavItem struct {
	ID    string
	Label string
}

// NavGroup is a heading with its sub-items. It is not selectable.
type NavGroup struct {
	Label string
	Items []NavItem
}

// Nav is the sidebar's navigation tree.
type Nav struct {
	Groups []NavGroup
	// Focused reports whether the sidebar has the keyboard focus.
	Focused bool
	cursor  int // index over the flattened list of items
}

func NewNav(groups ...NavGroup) Nav {
	return Nav{Groups: groups, Focused: true}
}

// WithFocus returns a copy with focus set or cleared.
func (n Nav) WithFocus(focused bool) Nav {
	n.Focused = focused
	return n
}

// items flattens every group's sub-items in screen order.
func (n Nav) items() []NavItem {
	var out []NavItem
	for _, g := range n.Groups {
		out = append(out, g.Items...)
	}
	return out
}

// Selected returns the item under the cursor, or an empty NavItem if there is
// none.
func (n Nav) Selected() NavItem {
	items := n.items()
	if len(items) == 0 {
		return NavItem{}
	}
	return items[clamp(n.cursor, 0, len(items)-1)]
}

func (n Nav) Next() Nav {
	if last := len(n.items()) - 1; n.cursor < last {
		n.cursor++
	}
	return n
}

func (n Nav) Prev() Nav {
	if n.cursor > 0 {
		n.cursor--
	}
	return n
}

func (n Nav) View(t styles.Theme, width, height int) string {
	inner := width - t.Nav.Base.GetHorizontalFrameSize()
	if inner <= 0 {
		return ""
	}

	lines, cursorLine := n.lines(t, inner)

	// The window clips lines, but the cursor indexes items: that is why lines
	// returns which line the selected one fell on.
	if offset := ui.CenteredOffset(cursorLine, len(lines), height); offset > 0 {
		lines = lines[offset:min(offset+height, len(lines))]
	}

	// Height pads but does not clip; without MaxHeight the menu would overflow.
	return t.Nav.Base.Width(width).Height(height).MaxHeight(height).Render(
		lipgloss.JoinVertical(lipgloss.Left, lines...),
	)
}

// lines renders the whole menu and returns the line of the selected item.
func (n Nav) lines(t styles.Theme, width int) (lines []string, cursorLine int) {
	item := 0 // index over the flattened list
	for gi, g := range n.Groups {
		if gi > 0 {
			lines = append(lines, "")
		}
		// Uppercase: they are grouping labels, not content.
		lines = append(lines, t.Nav.Group.Render(fit(t, strings.ToUpper(g.Label), width)))

		for _, entry := range g.Items {
			style, label := t.Nav.Item, "  "+entry.Label
			if item == n.cursor {
				cursorLine = len(lines)
				style, label = t.Nav.Selected, t.Icon.Cursor+" "+entry.Label
				if !n.Focused {
					style = t.Nav.Blurred
				}
			}
			// Width on the style so the background reaches the edge.
			lines = append(lines, style.Width(width).Render(fit(t, label, width)))
			item++
		}
	}
	return lines, cursorLine
}

// fit truncates while respecting the ANSI codes, which would break if cut by
// byte.
func fit(t styles.Theme, s string, width int) string {
	if width <= 0 {
		return ""
	}
	if ansi.StringWidth(s) <= width {
		return s
	}
	return ansi.Truncate(s, width, t.Icon.Ellipsis)
}

func clamp(v, lo, hi int) int {
	return min(max(v, lo), hi)
}
