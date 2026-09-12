package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/BurntSushi/toml"
)

// Settings bounds. A step below 1 makes the volume keys inert; above 50 a
// single press swings across the range.
const (
	minVolumeStep     = 1
	maxVolumeStep     = 50
	defaultVolumeStep = 5
)

// Settings are the behavioural knobs read from config.toml, already resolved
// to the types their consumers use.
type Settings struct {
	// VolumeStep is how much a left/right press moves the volume, in percent.
	VolumeStep int
}

// DefaultSettings is what knob uses with no config.toml: the values it shipped
// with before they were configurable.
func DefaultSettings() Settings {
	return Settings{
		VolumeStep: defaultVolumeStep,
	}
}

// rawSettings mirrors the config.toml layout. An absent key stays at its zero
// value, which LoadSettings reads as "keep the default".
type rawSettings struct {
	Audio struct {
		VolumeStep int `toml:"volume_step"`
	} `toml:"audio"`
}

// LoadSettings reads the user's behavioural settings.
//
// Like LoadTheme, it always returns a usable value: errors accompany the
// result rather than replace it. A missing file is a normal first run; a bad
// value keeps that field's default and is reported, so one typo does not
// leave anyone unable to start knob.
func LoadSettings() (Settings, []error) {
	out := DefaultSettings()

	path, err := SettingsPath()
	if err != nil {
		return out, []error{fmt.Errorf("could not locate the configuration: %w", err)}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return out, nil
		}
		return out, []error{fmt.Errorf("could not read %s: %w", path, err)}
	}

	var parsed rawSettings
	if err := toml.Unmarshal(data, &parsed); err != nil {
		return out, []error{fmt.Errorf("%s is not valid TOML: %w", path, err)}
	}

	var errs []error

	if v := parsed.Audio.VolumeStep; v != 0 {
		if v < minVolumeStep || v > maxVolumeStep {
			errs = append(errs, fmt.Errorf("[audio] volume_step: %d is not between %d and %d", v, minVolumeStep, maxVolumeStep))
		} else {
			out.VolumeStep = v
		}
	}

	return out, errs
}
