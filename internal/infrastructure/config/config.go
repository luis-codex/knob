// Package config reads the user's configuration files — the theme and the
// behavioural settings — and resolves them to the values their consumers use.
//
// It is an infrastructure service: it touches the disk and hands the result to
// the composition root. Like the rest of infrastructure it implements no
// domain port; unlike the repositories it is not driven by the application, so
// it has no interface in domain either. Its only client is cmd/knob.
//
// Both loaders are read-only: knob never writes theme.toml or config.toml,
// they are hand-edited. WriteExampleTheme only scaffolds a commented template.
package config

import (
	"os"
	"path/filepath"
)

// appDir is knob's folder under the user's config directory.
const appDir = "knob"

// File names inside appDir: the theme and the behavioural settings.
const (
	themeFile    = "theme.toml"
	settingsFile = "config.toml"
)

// configPath joins name onto the user's config directory. It respects
// XDG_CONFIG_HOME.
func configPath(name string) (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appDir, name), nil
}

// ThemePath is where the theme file is looked for.
func ThemePath() (string, error) { return configPath(themeFile) }

// SettingsPath is where the behavioural-settings file is looked for.
func SettingsPath() (string, error) { return configPath(settingsFile) }
