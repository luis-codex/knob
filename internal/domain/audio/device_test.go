package audio_test

import (
	"errors"
	"strings"
	"testing"

	"knob/internal/domain/audio"
	"knob/internal/domain/errs"
)

func mustDevice(t *testing.T, level int, muted bool) audio.Device {
	t.Helper()

	id, err := audio.NewID("alsa_output.hdmi-stereo")
	if err != nil {
		t.Fatalf("NewID: %v", err)
	}
	name, err := audio.NewName("HDMI Output")
	if err != nil {
		t.Fatalf("NewName: %v", err)
	}
	volume, err := audio.NewVolume(level)
	if err != nil {
		t.Fatalf("NewVolume(%d): %v", level, err)
	}

	device, err := audio.Restore(id, name, audio.Output, volume, muted, true)
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	return device
}

func TestNewVolume(t *testing.T) {
	tests := []struct {
		name  string
		level int
		err   error
	}{
		{name: "zero is valid", level: 0},
		{name: "nominal", level: 100},
		{name: "amplified", level: 150},
		{name: "negative", level: -1, err: audio.ErrInvalidVolume},
		{name: "past the max", level: 151, err: audio.ErrInvalidVolume},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := audio.NewVolume(tc.level)

			if tc.err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("error = %v, want %v", err, tc.err)
				}
				if !errs.IsInvalid(err) {
					t.Error("must arrive classified as invalid")
				}
				return
			}
			if err != nil || got.Level() != tc.level {
				t.Fatalf("NewVolume(%d) = %d, %v", tc.level, got.Level(), err)
			}
		})
	}
}

// The zero value of Volume is 0%, a legitimate level, not an invalid value.
func TestVolumeZeroIsValid(t *testing.T) {
	var zero audio.Volume
	if zero.Level() != 0 || zero.Amplified() {
		t.Errorf("zero value = %d/%v", zero.Level(), zero.Amplified())
	}
}

func TestClampVolume(t *testing.T) {
	tests := []struct{ in, want int }{
		{-50, 0},
		{0, 0},
		{75, 75},
		{150, 150},
		{200, audio.MaxVolume},
	}

	for _, tc := range tests {
		if got := audio.ClampVolume(tc.in); got.Level() != tc.want {
			t.Errorf("ClampVolume(%d) = %d, want %d", tc.in, got.Level(), tc.want)
		}
	}
}

func TestAmplified(t *testing.T) {
	for _, tc := range []struct {
		level int
		want  bool
	}{{0, false}, {100, false}, {101, true}, {150, true}} {
		v, err := audio.NewVolume(tc.level)
		if err != nil {
			t.Fatalf("NewVolume(%d): %v", tc.level, err)
		}
		if v.Amplified() != tc.want {
			t.Errorf("%d%%: Amplified = %v", tc.level, v.Amplified())
		}
	}
}

// AdjustVolume clamps at the ends: reaching the limit with a volume-up key is
// not an error.
func TestAdjustVolume(t *testing.T) {
	tests := []struct {
		name  string
		start int
		delta int
		want  int
	}{
		{name: "up", start: 50, delta: 10, want: 60},
		{name: "down", start: 50, delta: -10, want: 40},
		{name: "does not go below zero", start: 5, delta: -20, want: 0},
		{name: "does not go past the max", start: 145, delta: 20, want: audio.MaxVolume},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := mustDevice(t, tc.start, false).AdjustVolume(tc.delta)
			if got.Volume().Level() != tc.want {
				t.Errorf("%d %+d = %d, want %d", tc.start, tc.delta, got.Volume().Level(), tc.want)
			}
		})
	}
}

// Volume and mute are independent controls.
func TestVolumeAndMuteAreIndependent(t *testing.T) {
	muted := mustDevice(t, 50, true)

	louder := muted.SetVolume(audio.ClampVolume(80))
	if !louder.Muted() {
		t.Error("raising the volume must not clear mute")
	}
	if louder.Volume().Level() != 80 {
		t.Errorf("volume = %d", louder.Volume().Level())
	}

	unmuted := louder.SetMuted(false)
	if unmuted.Volume().Level() != 80 {
		t.Error("clearing mute must not touch the volume")
	}
}

func TestToggleMuted(t *testing.T) {
	device := mustDevice(t, 50, false)

	once := device.ToggleMuted()
	if !once.Muted() {
		t.Error("did not mute")
	}
	if device.Muted() {
		t.Error("mutated the original instead of returning a copy")
	}
	if twice := once.ToggleMuted(); twice.Muted() {
		t.Error("twice must return to the initial state")
	}
}

func TestRestore(t *testing.T) {
	id, _ := audio.NewID("x")
	name, _ := audio.NewName("X")

	t.Run("no identifier", func(t *testing.T) {
		if _, err := audio.Restore(audio.ID{}, name, audio.Input, audio.Volume{}, false, false); !errors.Is(err, audio.ErrInvalidID) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("no name", func(t *testing.T) {
		if _, err := audio.Restore(id, audio.Name{}, audio.Input, audio.Volume{}, false, false); !errors.Is(err, audio.ErrEmptyName) {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestNewID(t *testing.T) {
	if _, err := audio.NewID("   "); !errors.Is(err, audio.ErrInvalidID) {
		t.Errorf("empty id: %v", err)
	}
	got, err := audio.NewID("  alsa_output.x  ")
	if err != nil || got.String() != "alsa_output.x" {
		t.Errorf("NewID = %q, %v", got, err)
	}
}

func TestNewName(t *testing.T) {
	if _, err := audio.NewName(""); !errors.Is(err, audio.ErrEmptyName) {
		t.Errorf("empty name: %v", err)
	}
	if _, err := audio.NewName(strings.Repeat("a", audio.MaxNameLength+1)); !errors.Is(err, audio.ErrNameTooLong) {
		t.Errorf("long name: %v", err)
	}
}

func TestDirectionString(t *testing.T) {
	if audio.Output.String() != "output" || audio.Input.String() != "input" {
		t.Errorf("Direction.String = %q/%q", audio.Output, audio.Input)
	}
}
