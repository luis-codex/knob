// Package tomlstore implements preferences.Repository against a single TOML
// file, ~/.config/knob/config.toml (XDG_CONFIG_HOME-aware).
//
// It is an infrastructure adapter: the file is the store today, but nothing
// above this package knows that. Swapping it for a database later means
// writing a new adapter beside this one and changing the composition root,
// not touching domain or application.
package tomlstore

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"knob/internal/domain/preferences"
)

// appDir is knob's folder under the user's config directory; fileName is the
// store itself.
const (
	appDir   = "knob"
	fileName = "config.toml"
)

// Path is where the store is looked for. It respects XDG_CONFIG_HOME.
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appDir, fileName), nil
}

// Repository is the TOML-file adapter for preferences.Repository.
type Repository struct{}

func NewRepository() *Repository { return &Repository{} }

// Load reads the store. A missing file is a normal first run: the defaults
// are written so the file becomes the discoverable reference (there is no
// commented template anymore, the store is app-owned), and that write is
// best-effort -- a read-only home must not stop knob from starting.
//
// A value the store holds but the domain rejects falls back to its default
// and is named in Loaded.Rejected. The returned error is reserved for not
// being able to read the store at all: an unreadable directory, or a file
// that is not valid TOML. That distinction matters for Save, see below.
func (r *Repository) Load(ctx context.Context) (preferences.Loaded, error) {
	path, err := Path()
	if err != nil {
		return preferences.Loaded{Settings: preferences.Default()}, fmt.Errorf("could not locate the configuration: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return preferences.Loaded{Settings: preferences.Default()}, fmt.Errorf("could not read %s: %w", path, err)
		}
		out := preferences.Default()
		_ = write(path, out) // best-effort; a failure here is not fatal
		return preferences.Loaded{Settings: out}, nil
	}

	var raw rawFile
	if _, err := toml.Decode(string(data), &raw); err != nil {
		return preferences.Loaded{Settings: preferences.Default()}, fmt.Errorf("%s is not valid TOML: %w", path, err)
	}

	settings, rejected := fromRaw(raw)
	return preferences.Loaded{Settings: settings, Rejected: rejected}, nil
}

// Save persists s, replacing whatever was stored before.
//
// It refuses to write over a file that does not currently parse: a syntax
// error is harmless while the store is read-only, but once knob can write
// its own file, saving over one it could not read would silently destroy
// whatever the user was in the middle of fixing. The caller must not lose
// that file; they fix the syntax error by hand, then a save goes through.
func (r *Repository) Save(ctx context.Context, s preferences.Settings) error {
	path, err := Path()
	if err != nil {
		return fmt.Errorf("could not locate the configuration: %w", err)
	}

	if data, err := os.ReadFile(path); err == nil {
		var probe rawFile
		if _, err := toml.Decode(string(data), &probe); err != nil {
			return fmt.Errorf("%s has a syntax error; fix it by hand before saving: %w", path, err)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("could not read %s: %w", path, err)
	}

	return write(path, s)
}

// write marshals s and replaces path atomically: a crash mid-write must never
// leave a truncated file where the user's preferences used to be. The temp
// file gets a unique name (not path+".tmp") because Bubble Tea runs each
// tea.Cmd in its own goroutine, so two saves in flight at once must not race
// on the same temp path.
func write(path string, s preferences.Settings) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := toml.Marshal(toRaw(s))
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, "config-*.toml.tmp") // mode 0600, per os.CreateTemp
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }() // no-op once the rename below succeeds

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
