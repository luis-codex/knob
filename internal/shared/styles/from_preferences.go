package styles

import (
	"image/color"

	"charm.land/lipgloss/v2"

	"knob/internal/domain/preferences"
)

// FromPreferences turns the stored theme into the palettes For picks between.
// An unset domain color stays nil, which Merge already reads as "keep the
// default" -- this replaces the TOML-specific toPalette that used to live in
// infrastructure/config.
func FromPreferences(t preferences.Theme) Custom {
	return Custom{
		Light: paletteFromPreferences(t.Light),
		Dark:  paletteFromPreferences(t.Dark),
	}
}

func paletteFromPreferences(p preferences.Palette) Palette {
	toColor := func(c preferences.Color) color.Color {
		if c.IsZero() {
			return nil
		}
		return lipgloss.Color(c.String())
	}

	return Palette{
		Text:         toColor(p.Text),
		Muted:        toColor(p.Muted),
		Faint:        toColor(p.Faint),
		Line:         toColor(p.Line),
		LineStrong:   toColor(p.LineStrong),
		Accent:       toColor(p.Accent),
		AccentStrong: toColor(p.AccentStrong),
		OnAccent:     toColor(p.OnAccent),
		Bg:           toColor(p.Bg),
		BgSelected:   toColor(p.BgSelected),
		Success:      toColor(p.Success),
		Warning:      toColor(p.Warning),
		Danger:       toColor(p.Danger),
	}
}
