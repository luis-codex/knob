package tomlstore

import "knob/internal/domain/preferences"

// rawFile mirrors config.toml's layout. String fields use "" as their
// absent-key sentinel -- 0 and "" are not valid values for any of them, so
// the ambiguity the domain would otherwise have with a real zero value does
// not exist. Booleans cannot use that trick (false is a real value), so
// their fields are pointers: nil means absent, keep the default.
type rawFile struct {
	Audio     rawAudio     `toml:"audio"`
	Theme     rawTheme     `toml:"theme"`
	Interface rawInterface `toml:"interface"`
}

type rawAudio struct {
	VolumeStep int `toml:"volume_step"`
}

type rawTheme struct {
	Mode  string     `toml:"mode"`
	Dark  rawPalette `toml:"dark"`
	Light rawPalette `toml:"light"`
}

// rawPalette is the colors exactly as they come from the file, unvalidated.
type rawPalette struct {
	Text         string `toml:"text"`
	Muted        string `toml:"muted"`
	Faint        string `toml:"faint"`
	Line         string `toml:"line"`
	LineStrong   string `toml:"line_strong"`
	Accent       string `toml:"accent"`
	AccentStrong string `toml:"accent_strong"`
	OnAccent     string `toml:"on_accent"`
	Bg           string `toml:"bg"`
	BgSelected   string `toml:"bg_selected"`
	Success      string `toml:"success"`
	Warning      string `toml:"warning"`
	Danger       string `toml:"danger"`
}

type rawInterface struct {
	SidebarHidden *bool `toml:"sidebar_hidden"`
	Animations    *bool `toml:"animations"`
}

// fromRaw resolves raw into domain preferences. A value the domain rejects
// keeps its default and is reported instead of failing the whole file.
func fromRaw(raw rawFile) (preferences.Settings, []preferences.Rejected) {
	out := preferences.Default()
	var rejected []preferences.Rejected

	if v := raw.Audio.VolumeStep; v != 0 {
		if step, err := preferences.NewVolumeStep(v); err != nil {
			rejected = append(rejected, preferences.Rejected{Key: "audio.volume_step", Reason: err.Error()})
		} else {
			out.Audio.VolumeStep = step
		}
	}

	if v := raw.Theme.Mode; v != "" {
		if mode, err := preferences.NewMode(v); err != nil {
			rejected = append(rejected, preferences.Rejected{Key: "theme.mode", Reason: err.Error()})
		} else {
			out.Theme.Mode = mode
		}
	}

	dark, darkRejected := toPalette("theme.dark", raw.Theme.Dark)
	light, lightRejected := toPalette("theme.light", raw.Theme.Light)
	out.Theme.Dark = dark
	out.Theme.Light = light
	rejected = append(rejected, darkRejected...)
	rejected = append(rejected, lightRejected...)

	if raw.Interface.SidebarHidden != nil {
		out.Interface.SidebarHidden = *raw.Interface.SidebarHidden
	}
	if raw.Interface.Animations != nil {
		out.Interface.Animations = *raw.Interface.Animations
	}

	return out, rejected
}

// toPalette turns the hex strings into domain colors. One invalid color does
// not invalidate the rest: it is dropped, reported, and the rest carry on.
func toPalette(section string, p rawPalette) (preferences.Palette, []preferences.Rejected) {
	var rejected []preferences.Rejected
	var out preferences.Palette

	fields := []struct {
		key   string
		value string
		dst   *preferences.Color
	}{
		{"text", p.Text, &out.Text},
		{"muted", p.Muted, &out.Muted},
		{"faint", p.Faint, &out.Faint},
		{"line", p.Line, &out.Line},
		{"line_strong", p.LineStrong, &out.LineStrong},
		{"accent", p.Accent, &out.Accent},
		{"accent_strong", p.AccentStrong, &out.AccentStrong},
		{"on_accent", p.OnAccent, &out.OnAccent},
		{"bg", p.Bg, &out.Bg},
		{"bg_selected", p.BgSelected, &out.BgSelected},
		{"success", p.Success, &out.Success},
		{"warning", p.Warning, &out.Warning},
		{"danger", p.Danger, &out.Danger},
	}

	for _, f := range fields {
		if f.value == "" {
			continue // absent: the default stays
		}
		c, err := preferences.NewColor(f.value)
		if err != nil {
			rejected = append(rejected, preferences.Rejected{Key: section + "." + f.key, Reason: err.Error()})
			continue
		}
		*f.dst = c
	}

	return out, rejected
}

// toRaw is the inverse mapping, used only by write. Save always has a
// complete domain Settings, so there is no absent-key ambiguity to preserve
// here -- every field is written as it stands.
func toRaw(s preferences.Settings) rawFile {
	sidebarHidden := s.Interface.SidebarHidden
	animations := s.Interface.Animations

	return rawFile{
		Audio: rawAudio{VolumeStep: s.Audio.VolumeStep.Value()},
		Theme: rawTheme{
			Mode:  s.Theme.Mode.String(),
			Dark:  fromPalette(s.Theme.Dark),
			Light: fromPalette(s.Theme.Light),
		},
		Interface: rawInterface{
			SidebarHidden: &sidebarHidden,
			Animations:    &animations,
		},
	}
}

func fromPalette(p preferences.Palette) rawPalette {
	return rawPalette{
		Text:         p.Text.String(),
		Muted:        p.Muted.String(),
		Faint:        p.Faint.String(),
		Line:         p.Line.String(),
		LineStrong:   p.LineStrong.String(),
		Accent:       p.Accent.String(),
		AccentStrong: p.AccentStrong.String(),
		OnAccent:     p.OnAccent.String(),
		Bg:           p.Bg.String(),
		BgSelected:   p.BgSelected.String(),
		Success:      p.Success.String(),
		Warning:      p.Warning.String(),
		Danger:       p.Danger.String(),
	}
}
