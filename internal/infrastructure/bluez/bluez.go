// Package bluez implements the Bluetooth ports against the system's BlueZ
// daemon by invoking bluetoothctl.
//
// The CLI is used rather than D-Bus directly to avoid pulling in
// dependencies: the trade-off is parsing text, which is what parseFields does.
package bluez

import (
	"bufio"
	"context"
	"os/exec"
	"strconv"
	"strings"

	"knob/internal/domain/bluetooth"
	"knob/internal/domain/errs"
)

// binary is the executable that gets invoked. A variable so it can be swapped
// in tests.
var binary = "bluetoothctl"

// run executes bluetoothctl and returns its output.
//
// bluetoothctl writes failures to stdout and returns 0 in several cases, so
// the exit code is not enough: the caller must inspect the content.
func run(ctx context.Context, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, binary, args...).CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return string(out), errs.Wrap(errs.KindConflict, "could not talk to BlueZ", err)
	}
	return string(out), nil
}

// parseFields extracts the "Key: value" lines from the output. Keys repeat
// (UUID), so it keeps the first one, which is the one that matters.
func parseFields(out string) map[string]string {
	fields := make(map[string]string)

	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		key, value, ok := strings.Cut(strings.TrimSpace(scanner.Text()), ": ")
		if !ok {
			continue
		}
		if _, seen := fields[key]; !seen {
			fields[key] = strings.TrimSpace(value)
		}
	}
	return fields
}

func yes(fields map[string]string, key string) bool {
	return fields[key] == "yes"
}

// parseBattery reads "Battery Percentage: 0x52 (82)": the useful value is the
// decimal in parentheses.
func parseBattery(raw string) bluetooth.Battery {
	open := strings.Index(raw, "(")
	closeIdx := strings.Index(raw, ")")
	if open < 0 || closeIdx < open {
		return bluetooth.UnknownBattery()
	}

	level, err := strconv.Atoi(raw[open+1 : closeIdx])
	if err != nil {
		return bluetooth.UnknownBattery()
	}

	battery, err := bluetooth.NewBattery(level)
	if err != nil {
		return bluetooth.UnknownBattery()
	}
	return battery
}

// icons maps the BlueZ icon to the domain kind.
var icons = map[string]bluetooth.Kind{
	"audio-headset":    bluetooth.KindHeadphones,
	"audio-headphones": bluetooth.KindHeadphones,
	"audio-card":       bluetooth.KindSpeaker,
	"input-mouse":      bluetooth.KindMouse,
	"input-keyboard":   bluetooth.KindKeyboard,
	"input-tablet":     bluetooth.KindKeyboard,
	"phone":            bluetooth.KindPhone,
}

func parseKind(icon string) bluetooth.Kind {
	if kind, ok := icons[icon]; ok {
		return kind
	}
	return bluetooth.KindUnknown
}
