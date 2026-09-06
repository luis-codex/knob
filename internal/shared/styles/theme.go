// Package styles concentrates the visual tokens: spacing, sizes, semantic
// colors and the styles already resolved. No other package declares colors or
// paddings.
package styles

import (
	"image/color"

	"charm.land/lipgloss/v2"

	"settings-cli/internal/shared/icons"
)

// Spacing and size tokens: the app's visual rhythm.
const (
	PadX = 1 // horizontal padding of each section
	PadY = 0 // sections are already separated by dividers

	SidebarWidth = 26
	HeaderHeight = 1
	FooterHeight = 1
	DividerSize  = 1 // thickness of a divider

	DialogWidth = 50
)

// Palette names the colors by role (Accent, Muted) and not by hue, so
// changing the palette does not force a rename.
//
// The dark values are bun.com's: an almost-black warm-tinted background, three
// text levels and a magenta used sparingly.
type Palette struct {
	Text  color.Color
	Muted color.Color
	// Faint is the third text level: labels and metadata that are only read if
	// you look for them.
	Faint color.Color

	Line color.Color
	// LineStrong frames what does need to stand out, like a modal.
	LineStrong color.Color

	Accent color.Color
	// AccentStrong is the accent emphasized or over an accent background.
	AccentStrong color.Color
	OnAccent     color.Color // text readable over Accent

	Bg         color.Color
	BgSelected color.Color

	Success color.Color
	Warning color.Color
	Danger  color.Color
}

type headerStyles struct {
	Base  lipgloss.Style
	Title lipgloss.Style
	// Crumb is the current section, to the right of the name.
	Crumb lipgloss.Style
}

type navStyles struct {
	Base     lipgloss.Style
	Group    lipgloss.Style
	Item     lipgloss.Style
	Selected lipgloss.Style
	// Blurred is the selected item without focus.
	Blurred lipgloss.Style
}

type inputStyles struct {
	Label       lipgloss.Style
	Blurred     lipgloss.Style
	Focused     lipgloss.Style
	Placeholder lipgloss.Style
	Cursor      lipgloss.Style
}

type scrollbarStyles struct {
	Track lipgloss.Style
	Thumb lipgloss.Style
}

type listStyles struct {
	Item     lipgloss.Style
	Selected lipgloss.Style
	Empty    lipgloss.Style
}

type bodyStyles struct {
	Base   lipgloss.Style
	Title  lipgloss.Style
	Label  lipgloss.Style
	Value  lipgloss.Style
	Muted  lipgloss.Style
	Badge  lipgloss.Style
	Danger lipgloss.Style
	Note   lipgloss.Style
	// Section is the label of a section within the page.
	Section lipgloss.Style
	// Hint is the keyboard hints within the page. The key is in plain text and
	// not in accent: the footer already carries the accent, and two things
	// shouting at once establish no hierarchy.
	Hint HintStyles
}

// HintStyles renders a list of keyboard hints.
type HintStyles struct {
	// Key is the key itself, which is what the user looks for.
	Key lipgloss.Style
	// Action is what that key does.
	Action lipgloss.Style
	// Sep separates one hint from the next.
	Sep lipgloss.Style
}

type footerStyles struct {
	Base lipgloss.Style
	Hint HintStyles
}

type buttonStyles struct {
	Blurred lipgloss.Style
	Focused lipgloss.Style
	// Danger is the primary button of a destructive action with focus.
	Danger lipgloss.Style
}

type dialogStyles struct {
	Box   lipgloss.Style
	Title lipgloss.Style
	Text  lipgloss.Style
	Label lipgloss.Style
	Error lipgloss.Style
	Hint  lipgloss.Style
}

// Icons are the interface's symbols.
//
// They live in the theme and not loose in the code so that changing the
// symbol set is touching a single place. They must all take one cell: the row
// grid counts on it.
type Icons struct {
	// Cursor marks the row under the cursor.
	Cursor string
	// Active marks the element in use: the default device.
	Active string
	// Paused marks a stopped audio stream.
	Paused string
	// Playing marks a playing stream; the screen makes it blink.
	Playing string
	// BarOn and BarOff are the full and empty segments of a bar.
	BarOn  string
	BarOff string
	// DividerH and DividerV separate the layout regions.
	DividerH string
	DividerV string
	// ScrollThumb and ScrollTrack are the scrollbar.
	ScrollThumb string
	ScrollTrack string
	// Ellipsis closes a clipped text.
	Ellipsis string
}

func defaultIcons() Icons {
	return Icons{
		Cursor:      icons.Cursor,
		Active:      icons.Active,
		Paused:      icons.Paused,
		Playing:     icons.Playing,
		BarOn:       icons.BarOn,
		BarOff:      icons.BarOff,
		DividerH:    icons.DividerH,
		DividerV:    icons.DividerV,
		ScrollThumb: icons.ScrollThumb,
		ScrollTrack: icons.ScrollTrack,
		Ellipsis:    icons.Ellipsis,
	}
}

// Theme is the styles resolved for the current terminal background.
type Theme struct {
	Color Palette
	Icon  Icons

	Header    headerStyles
	Nav       navStyles
	Body      bodyStyles
	Footer    footerStyles
	Dialog    dialogStyles
	Button    buttonStyles
	Input     inputStyles
	List      listStyles
	Scrollbar scrollbarStyles
	Divider   lipgloss.Style
}

// Custom are the palettes the user defines for each background type.
type Custom struct {
	Light Palette
	Dark  Palette
}

// For picks the palette that applies based on the terminal background.
func (c Custom) For(isDark bool) Palette {
	if isDark {
		return c.Dark
	}
	return c.Light
}

// Merge returns the palette with custom's non-nil colors substituted in.
//
// The fields are interfaces, so the zero value of Palette means "customize
// nothing": the user declares only the colors they want to change.
func (p Palette) Merge(custom Palette) Palette {
	set := func(dst *color.Color, src color.Color) {
		if src != nil {
			*dst = src
		}
	}

	set(&p.Text, custom.Text)
	set(&p.Muted, custom.Muted)
	set(&p.Faint, custom.Faint)
	set(&p.Line, custom.Line)
	set(&p.LineStrong, custom.LineStrong)
	set(&p.Accent, custom.Accent)
	set(&p.AccentStrong, custom.AccentStrong)
	set(&p.OnAccent, custom.OnAccent)
	set(&p.Bg, custom.Bg)
	set(&p.BgSelected, custom.BgSelected)
	set(&p.Success, custom.Success)
	set(&p.Warning, custom.Warning)
	set(&p.Danger, custom.Danger)

	return p
}

// New builds the theme with the default palette.
func New(isDark bool) Theme {
	return NewWithPalette(isDark, Palette{})
}

// NewWithPalette builds the theme applying the user's colors over the default
// ones. isDark comes from tea.BackgroundColorMsg.IsDark().
func NewWithPalette(isDark bool, custom Palette) Theme {
	pick := lipgloss.LightDark(isDark)

	p := Palette{
		Text:  pick(lipgloss.Color("#1A1A1A"), lipgloss.Color("#EAEAE8")),
		Muted: pick(lipgloss.Color("#6B6B68"), lipgloss.Color("#A8A8A5")),
		Faint: pick(lipgloss.Color("#94948F"), lipgloss.Color("#80807E")),

		Line:       pick(lipgloss.Color("#E2E2E0"), lipgloss.Color("#28282B")),
		LineStrong: pick(lipgloss.Color("#C9C9C6"), lipgloss.Color("#3E3E42")),

		// By default the accent is not a color, it is contrast: white on dark
		// and black on light. That way the app imposes no identity and whoever
		// wants color puts it in their theme.toml.
		//
		// Previous palettes, in case they are wanted back:
		//	Magenta (bun.com):
		//	Accent:       pick(lipgloss.Color("#D6006E"), lipgloss.Color("#FF2E97")),
		//	AccentStrong: pick(lipgloss.Color("#FF2E97"), lipgloss.Color("#FF5CB0")),
		//	Blue:
		//	Accent:       pick(lipgloss.Color("#3B4FC4"), lipgloss.Color("#516BEB")),
		//	AccentStrong: pick(lipgloss.Color("#516BEB"), lipgloss.Color("#7D91F2")),
		Accent:       pick(lipgloss.Color("#1A1A1A"), lipgloss.Color("#EAEAE8")),
		AccentStrong: pick(lipgloss.Color("#000000"), lipgloss.Color("#FFFFFF")),
		OnAccent:     pick(lipgloss.Color("#FFFFFF"), lipgloss.Color("#0D0A0C")),

		Bg: pick(lipgloss.Color("#FAFAF8"), lipgloss.Color("#0D0A0C")),
		// With a neutral accent the selection background is neutral too: a grey
		// that separates the row without tinting it. With magenta it was
		// #FFE7F2 / #280016 and with blue #E7EBFD / #111634.
		BgSelected: pick(lipgloss.Color("#EDEDEA"), lipgloss.Color("#1F1F22")),

		Success: pick(lipgloss.Color("#12864F"), lipgloss.Color("#28DC82")),
		Warning: pick(lipgloss.Color("#B45309"), lipgloss.Color("#FBBF24")),
		Danger:  pick(lipgloss.Color("#C0392B"), lipgloss.Color("#FF5C5C")),
	}.Merge(custom)

	// No border of its own: the separation between areas is done by the dividers.
	section := lipgloss.NewStyle().Padding(PadY, PadX)

	return Theme{
		Color: p,
		Icon:  defaultIcons(),

		Header: headerStyles{
			Base:  section,
			Title: lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
			Crumb: lipgloss.NewStyle().Foreground(p.Faint),
		},
		Nav: navStyles{
			Base: section,
			// The heading outranks what it groups, so it goes in plain text:
			// fainter than its items made it disappear among them.
			Group:    lipgloss.NewStyle().Foreground(p.Text).Bold(true),
			Item:     lipgloss.NewStyle().Foreground(p.Muted),
			Selected: lipgloss.NewStyle().Foreground(p.AccentStrong).Background(p.BgSelected).Bold(true),
			Blurred:  lipgloss.NewStyle().Foreground(p.Faint).Background(p.BgSelected),
		},
		Body: bodyStyles{
			Base: section,
			// The page title goes in plain text, not accent: the magenta is
			// reserved for what needs looking at.
			Title:  lipgloss.NewStyle().Foreground(p.Text).Bold(true),
			Label:  lipgloss.NewStyle().Foreground(p.Text),
			Value:  lipgloss.NewStyle().Foreground(p.Muted),
			Muted:  lipgloss.NewStyle().Foreground(p.Muted),
			Badge:  lipgloss.NewStyle().Foreground(p.Success),
			Danger: lipgloss.NewStyle().Foreground(p.Danger),
			Note:   lipgloss.NewStyle().Foreground(p.Faint).Italic(true),
			// Same rule as Nav.Group: the label weighs more than its list.
			Section: lipgloss.NewStyle().Foreground(p.Text).Bold(true),
			Hint: HintStyles{
				Key:    lipgloss.NewStyle().Foreground(p.Text).Bold(true),
				Action: lipgloss.NewStyle().Foreground(p.Muted),
				Sep:    lipgloss.NewStyle().Foreground(p.Line),
			},
		},
		Footer: footerStyles{
			Base: section,
			// In the footer the key goes in accent: it is the emergency exit.
			Hint: HintStyles{
				Key:    lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
				Action: lipgloss.NewStyle().Foreground(p.Muted),
				Sep:    lipgloss.NewStyle().Foreground(p.Line),
			},
		},
		// The modal carries a border and background: that is what lifts it off
		// the body.
		Dialog: dialogStyles{
			Box: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(p.LineStrong).
				Background(p.Bg).
				Foreground(p.Text).
				Padding(1, 2),
			Title: lipgloss.NewStyle().Foreground(p.Text).Background(p.Bg).Bold(true),
			Text:  lipgloss.NewStyle().Foreground(p.Muted).Background(p.Bg),
			Label: lipgloss.NewStyle().Foreground(p.Faint).Background(p.Bg),
			Error: lipgloss.NewStyle().Foreground(p.Danger).Background(p.Bg),
			Hint:  lipgloss.NewStyle().Foreground(p.Faint).Background(p.Bg),
		},
		Button: buttonStyles{
			Blurred: lipgloss.NewStyle().Foreground(p.Muted).Background(p.BgSelected),
			// Full accent background with dark text: bun.com's badge.
			Focused: lipgloss.NewStyle().Foreground(p.OnAccent).Background(p.Accent).Bold(true),
			Danger:  lipgloss.NewStyle().Foreground(p.OnAccent).Background(p.Danger).Bold(true),
		},
		Input: inputStyles{
			Label:       lipgloss.NewStyle().Foreground(p.Faint),
			Blurred:     lipgloss.NewStyle().Foreground(p.Muted).Background(p.BgSelected),
			Focused:     lipgloss.NewStyle().Foreground(p.Text).Background(p.BgSelected),
			Placeholder: lipgloss.NewStyle().Foreground(p.Faint).Background(p.BgSelected).Italic(true),
			Cursor:      lipgloss.NewStyle().Foreground(p.OnAccent).Background(p.Accent),
		},
		List: listStyles{
			Item:     lipgloss.NewStyle().Foreground(p.Muted),
			Selected: lipgloss.NewStyle().Foreground(p.Text).Background(p.BgSelected).Bold(true),
			Empty:    lipgloss.NewStyle().Foreground(p.Faint).Italic(true),
		},
		Scrollbar: scrollbarStyles{
			Track: lipgloss.NewStyle().Foreground(p.Line),
			Thumb: lipgloss.NewStyle().Foreground(p.Accent),
		},
		Divider: lipgloss.NewStyle().Foreground(p.Line),
	}
}
