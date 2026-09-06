// Package pulse implements the audio port against the sound server by
// invoking `pactl -f json`. It works for both PulseAudio and PipeWire, which
// exposes the same protocol.
//
// The JSON output is used rather than the text one: it brings the booleans
// already typed, the default names directly, and avoids hand-parsing a format
// that changes between versions.
package pulse

import (
	"context"
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"

	"knob/internal/domain/audio"
	"knob/internal/domain/errs"
)

// binary is the executable that gets invoked. A variable so it can be swapped
// in tests.
var binary = "pactl"

func run(ctx context.Context, args ...string) ([]byte, error) {
	out, err := exec.CommandContext(ctx, binary, append([]string{"-f", "json"}, args...)...).Output()
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errs.Wrap(errs.KindConflict, "could not talk to the sound server", err)
	}
	return out, nil
}

// kindOf maps the direction to pactl's vocabulary.
func kindOf(direction audio.Direction) string {
	if direction == audio.Input {
		return "source"
	}
	return "sink"
}

// jsonChannel is a channel's volume. Only the percentage matters: the raw
// value and the decibels are a server detail.
type jsonChannel struct {
	ValuePercent string `json:"value_percent"`
}

type jsonDevice struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Mute        bool                   `json:"mute"`
	Volume      map[string]jsonChannel `json:"volume"`
	// MonitorSource, on a source, is the output it is a copy of. Empty means
	// it is a real input.
	MonitorSource string `json:"monitor_source"`
}

type jsonInfo struct {
	DefaultSink   string `json:"default_sink_name"`
	DefaultSource string `json:"default_source_name"`
}

// parsePercent reads "45%".
func parsePercent(s string) (int, bool) {
	digits, ok := strings.CutSuffix(strings.TrimSpace(s), "%")
	if !ok {
		return 0, false
	}

	value, err := strconv.Atoi(digits)
	if err != nil {
		return 0, false
	}
	return value, true
}

// volumeOf sums up the per-channel volume into a single level.
//
// It takes the maximum rather than "the first": the JSON brings them in a map,
// and Go's map iteration is random, so keeping any one of them would give a
// different level on each read of an unbalanced device. The maximum is also
// what mixers show.
func volumeOf(channels map[string]jsonChannel) (audio.Volume, bool) {
	level, found := 0, false

	for _, channel := range channels {
		value, ok := parsePercent(channel.ValuePercent)
		if !ok {
			continue
		}
		if !found || value > level {
			level, found = value, true
		}
	}

	if !found {
		return audio.Volume{}, false
	}
	return audio.ClampVolume(level), true
}

// toDevice translates a pactl device to the domain. It returns false if the
// block does not describe a usable device.
func toDevice(d jsonDevice, direction audio.Direction, defaultName string) (audio.Device, bool) {
	// Monitors are the copy of an output, not a real input: dropping them here
	// keeps a nonexistent microphone out of the list.
	if direction == audio.Input && d.MonitorSource != "" {
		return audio.Device{}, false
	}

	id, err := audio.NewID(d.Name)
	if err != nil {
		return audio.Device{}, false
	}

	name, err := audio.NewName(firstNonEmpty(d.Description, d.Name))
	if err != nil {
		return audio.Device{}, false
	}

	volume, ok := volumeOf(d.Volume)
	if !ok {
		return audio.Device{}, false
	}

	device, err := audio.Restore(id, name, direction, volume, d.Mute, d.Name == defaultName)
	if err != nil {
		return audio.Device{}, false
	}
	return device, true
}

func decodeDevices(raw []byte) ([]jsonDevice, error) {
	var devices []jsonDevice
	if err := json.Unmarshal(raw, &devices); err != nil {
		return nil, errs.Wrap(errs.KindConflict, "unreadable response from the sound server", err)
	}
	return devices, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
