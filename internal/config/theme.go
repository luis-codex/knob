// Package config reads the user's configuration and translates it to the
// interface's vocabulary.
//
// It does not live under infrastructure because it implements no domain port:
// it does not translate inward, but toward presentation. It is a helper for
// the composition root, which is its only client.
//
// The styles package stays as pure tokens and receives the colors already
// resolved: reading the disk is this package's job.
package config

import (
	"errors"
	"fmt"
	"image/color"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"charm.land/lipgloss/v2"
	"github.com/BurntSushi/toml"

	"settings-cli/internal/shared/styles"
)

const (
	// appDir is the folder under the user's config directory.
	appDir = "settings-cli"
	// themeFile is the theme file inside that folder.
	themeFile = "theme.toml"
)

// hexPattern accepts #RGB and #RRGGBB, which is what lipgloss understands.
var hexPattern = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// palette is the colors exactly as they come from the file, unvalidated.
type palette struct {
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

type file struct {
	Light palette `toml:"light"`
	Dark  palette `toml:"dark"`
}

// ThemePath is where the file is looked for. It respects XDG_CONFIG_HOME.
func ThemePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appDir, themeFile), nil
}

// LoadTheme reads the user's theme.
//
// It always returns a usable palette: errors accompany the result rather than
// replace it. A badly written theme must not leave anyone unable to open their
// settings, so whatever fails stays at the default value and is reported.
func LoadTheme() (styles.Custom, []error) {
	path, err := ThemePath()
	if err != nil {
		return styles.Custom{}, []error{fmt.Errorf("could not locate the configuration: %w", err)}
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		// Having no theme is normal, not a failure.
		if errors.Is(err, fs.ErrNotExist) {
			return styles.Custom{}, nil
		}
		return styles.Custom{}, []error{fmt.Errorf("could not read %s: %w", path, err)}
	}

	var parsed file
	if err := toml.Unmarshal(raw, &parsed); err != nil {
		return styles.Custom{}, []error{fmt.Errorf("%s is not valid TOML: %w", path, err)}
	}

	light, lightErrs := toPalette("light", parsed.Light)
	dark, darkErrs := toPalette("dark", parsed.Dark)

	return styles.Custom{Light: light, Dark: dark}, append(lightErrs, darkErrs...)
}

// toPalette turns the hex strings into colors. One invalid color does not
// invalidate the rest: it is dropped, reported, and the rest carry on.
func toPalette(section string, p palette) (styles.Palette, []error) {
	var errs []error

	parse := func(key, value string) color.Color {
		if value == "" {
			return nil // absent: the default stays
		}
		if !hexPattern.MatchString(value) {
			errs = append(errs, fmt.Errorf("[%s] %s: %q is not a hex color", section, key, value))
			return nil
		}
		return lipgloss.Color(value)
	}

	// A table instead of thirteen conditionals: adding a color to the theme is
	// one line here and one in the file struct.
	var out styles.Palette
	fields := []struct {
		key   string
		value string
		dst   *color.Color
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
		if c := parse(f.key, f.value); c != nil {
			*f.dst = c
		}
	}

	return out, errs
}

// exampleTheme is the template written by WriteExampleTheme. It carries every
// color commented out with its default value: that way you can see what is
// tunable without reading the code.
const exampleTheme = `# settings-cli theme.
#
# Uncomment only what you want to change: anything missing stays at the default
# value. Colors are hexadecimal, #RGB or #RRGGBB.

[dark]
# text          = "#EAEAE8"   # primary text
# muted         = "#A8A8A5"   # secondary text
# faint         = "#80807E"   # labels and metadata
# line          = "#28282B"   # separators
# line_strong   = "#3E3E42"   # modal border
accent          = "#516BEB"   # keys, selection, primary button
# accent_strong = "#7D91F2"   # emphasized accent
# on_accent     = "#0D0A0C"   # text on the accent
# bg            = "#0D0A0C"   # modal background
# bg_selected   = "#111634"   # selected-row background
# success       = "#28DC82"
# warning       = "#FBBF24"
# danger        = "#FF5C5C"

[light]
# text          = "#1A1A1A"
# muted         = "#6B6B68"
# faint         = "#94948F"
# line          = "#E2E2E0"
# line_strong   = "#C9C9C6"
# accent        = "#3B4FC4"
# accent_strong = "#516BEB"
# on_accent     = "#FFFFFF"
# bg            = "#FAFAF8"
# bg_selected   = "#E7EBFD"
# success       = "#12864F"
# warning       = "#B45309"
# danger        = "#C0392B"
`

// WriteExampleTheme drops the template in place and returns the path.
//
// It does not overwrite an existing theme: clobbering what the user has
// already tuned would be worse than doing nothing.
func WriteExampleTheme() (string, error) {
	path, err := ThemePath()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(path); err == nil {
		return path, fs.ErrExist
	} else if !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(exampleTheme), 0o644); err != nil {
		return "", err
	}
	return path, nil
}
