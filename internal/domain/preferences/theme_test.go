package preferences_test

import (
	"errors"
	"testing"

	"knob/internal/domain/errs"
	"knob/internal/domain/preferences"
)

func TestNewColor(t *testing.T) {
	tests := []struct {
		name string
		hex  string
		err  error
	}{
		{name: "short form", hex: "#516"},
		{name: "long form", hex: "#516BEB"},
		{name: "missing hash", hex: "516BEB", err: preferences.ErrInvalidColor},
		{name: "wrong length", hex: "#516BE", err: preferences.ErrInvalidColor},
		{name: "not hex", hex: "#GGGGGG", err: preferences.ErrInvalidColor},
		{name: "empty", hex: "", err: preferences.ErrInvalidColor},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := preferences.NewColor(tc.hex)

			if tc.err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("error = %v, want %v", err, tc.err)
				}
				if !errs.IsInvalid(err) {
					t.Error("must arrive classified as invalid")
				}
				return
			}
			if err != nil || got.String() != tc.hex {
				t.Fatalf("NewColor(%q) = %q, %v", tc.hex, got.String(), err)
			}
		})
	}
}

// The zero value of Color means "unset", not an invalid color.
func TestColorZeroIsUnset(t *testing.T) {
	var zero preferences.Color
	if !zero.IsZero() || zero.String() != "" {
		t.Errorf("zero value = %q, IsZero=%v", zero.String(), zero.IsZero())
	}
}

func TestNewMode(t *testing.T) {
	for _, valid := range []string{preferences.ModeAuto, preferences.ModeLight, preferences.ModeDark} {
		if _, err := preferences.NewMode(valid); err != nil {
			t.Errorf("NewMode(%q): %v", valid, err)
		}
	}

	if _, err := preferences.NewMode("system"); !errors.Is(err, preferences.ErrInvalidMode) {
		t.Errorf("error = %v, want %v", err, preferences.ErrInvalidMode)
	}
}

// The zero value of Mode is auto: an absent key must keep today's only
// behaviour, detecting the terminal's background.
func TestModeZeroIsAuto(t *testing.T) {
	var zero preferences.Mode
	if !zero.IsAuto() || zero.IsDark() || zero.IsLight() {
		t.Errorf("zero value: auto=%v dark=%v light=%v", zero.IsAuto(), zero.IsDark(), zero.IsLight())
	}
	if zero.String() != preferences.ModeAuto {
		t.Errorf("String() = %q, want %q", zero.String(), preferences.ModeAuto)
	}
}

func TestModeResolve(t *testing.T) {
	auto, _ := preferences.NewMode(preferences.ModeAuto)
	light, _ := preferences.NewMode(preferences.ModeLight)
	dark, _ := preferences.NewMode(preferences.ModeDark)

	tests := []struct {
		name         string
		mode         preferences.Mode
		detectedDark bool
		want         bool
	}{
		{name: "auto follows the terminal, dark", mode: auto, detectedDark: true, want: true},
		{name: "auto follows the terminal, light", mode: auto, detectedDark: false, want: false},
		{name: "dark overrides a light terminal", mode: dark, detectedDark: false, want: true},
		{name: "light overrides a dark terminal", mode: light, detectedDark: true, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.mode.Resolve(tc.detectedDark); got != tc.want {
				t.Errorf("Resolve(%v) = %v, want %v", tc.detectedDark, got, tc.want)
			}
		})
	}
}

func TestDefaultTheme(t *testing.T) {
	if mode := preferences.Default().Theme.Mode; !mode.IsAuto() {
		t.Errorf("default mode = %q, want auto", mode.String())
	}
}
