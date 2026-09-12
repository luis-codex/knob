package preferences

import "regexp"

// hexPattern accepts #RGB and #RRGGBB, which is what the renderer understands.
var hexPattern = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// Color is a validated hex color. Its zero value means "unset": the field
// keeps whatever default the renderer uses, the same rule Merge already
// applies for the palette as a whole.
type Color struct {
	value string
}

func NewColor(hex string) (Color, error) {
	if !hexPattern.MatchString(hex) {
		return Color{}, ErrInvalidColor
	}
	return Color{value: hex}, nil
}

func (c Color) String() string { return c.value }

func (c Color) IsZero() bool { return c.value == "" }

// Mode picks which palette applies, or lets the terminal decide.
type Mode struct {
	value string
}

const (
	ModeAuto  = "auto"
	ModeLight = "light"
	ModeDark  = "dark"
)

func NewMode(s string) (Mode, error) {
	switch s {
	case ModeAuto, ModeLight, ModeDark:
		return Mode{value: s}, nil
	default:
		return Mode{}, ErrInvalidMode
	}
}

func (m Mode) String() string {
	if m.value == "" {
		return ModeAuto
	}
	return m.value
}

// IsAuto reports whether the terminal's background decides the palette. The
// zero value is auto: an absent key keeps today's only behaviour.
func (m Mode) IsAuto() bool { return m.value == "" || m.value == ModeAuto }

func (m Mode) IsDark() bool { return m.value == ModeDark }

func (m Mode) IsLight() bool { return m.value == ModeLight }

// Resolve picks light or dark: an explicit mode wins, auto defers to
// detectedDark (the terminal's own background, in whatever way the caller
// detected it).
func (m Mode) Resolve(detectedDark bool) bool {
	switch {
	case m.IsDark():
		return true
	case m.IsLight():
		return false
	default:
		return detectedDark
	}
}

// Palette names the colors by role, one variant per background. A zero Color
// in any field means "use the built-in default for that role".
type Palette struct {
	Text         Color
	Muted        Color
	Faint        Color
	Line         Color
	LineStrong   Color
	Accent       Color
	AccentStrong Color
	OnAccent     Color
	Bg           Color
	BgSelected   Color
	Success      Color
	Warning      Color
	Danger       Color
}

// Theme groups the theme preferences: the mode and the two palettes it
// chooses between.
type Theme struct {
	Mode  Mode
	Dark  Palette
	Light Palette
}

func defaultTheme() Theme {
	return Theme{Mode: Mode{value: ModeAuto}}
}
