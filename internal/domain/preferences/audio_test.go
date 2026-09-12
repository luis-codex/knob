package preferences_test

import (
	"errors"
	"testing"

	"knob/internal/domain/errs"
	"knob/internal/domain/preferences"
)

func TestNewVolumeStep(t *testing.T) {
	tests := []struct {
		name  string
		value int
		err   error
	}{
		{name: "minimum", value: preferences.MinVolumeStep},
		{name: "default", value: preferences.DefaultVolumeStep},
		{name: "maximum", value: preferences.MaxVolumeStep},
		{name: "below the minimum", value: preferences.MinVolumeStep - 1, err: preferences.ErrInvalidVolumeStep},
		{name: "above the maximum", value: preferences.MaxVolumeStep + 1, err: preferences.ErrInvalidVolumeStep},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := preferences.NewVolumeStep(tc.value)

			if tc.err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("error = %v, want %v", err, tc.err)
				}
				if !errs.IsInvalid(err) {
					t.Error("must arrive classified as invalid")
				}
				return
			}
			if err != nil || got.Value() != tc.value {
				t.Fatalf("NewVolumeStep(%d) = %d, %v", tc.value, got.Value(), err)
			}
		})
	}
}

func TestDefaultAudio(t *testing.T) {
	if got := preferences.Default().Audio.VolumeStep.Value(); got != preferences.DefaultVolumeStep {
		t.Errorf("default volume step = %d, want %d", got, preferences.DefaultVolumeStep)
	}
}
