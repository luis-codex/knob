package audio_test

import (
	"errors"
	"strings"
	"testing"

	"settings-cli/internal/domain/audio"
	"settings-cli/internal/domain/errs"
)

func mustDevice(t *testing.T, level int, muted bool) audio.Device {
	t.Helper()

	id, err := audio.NewID("alsa_output.hdmi-stereo")
	if err != nil {
		t.Fatalf("NewID: %v", err)
	}
	name, err := audio.NewName("Salida HDMI")
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
		{name: "cero es válido", level: 0},
		{name: "nominal", level: 100},
		{name: "amplificado", level: 150},
		{name: "negativo", level: -1, err: audio.ErrInvalidVolume},
		{name: "pasado del máximo", level: 151, err: audio.ErrInvalidVolume},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := audio.NewVolume(tc.level)

			if tc.err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("error = %v, se esperaba %v", err, tc.err)
				}
				if !errs.IsInvalid(err) {
					t.Error("debe llegar clasificado como inválido")
				}
				return
			}
			if err != nil || got.Level() != tc.level {
				t.Fatalf("NewVolume(%d) = %d, %v", tc.level, got.Level(), err)
			}
		})
	}
}

// El valor cero de Volume es 0 %, un nivel legítimo, no un valor inválido.
func TestVolumeCeroEsValido(t *testing.T) {
	var zero audio.Volume
	if zero.Level() != 0 || zero.Amplified() {
		t.Errorf("valor cero = %d/%v", zero.Level(), zero.Amplified())
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
			t.Errorf("ClampVolume(%d) = %d, se esperaba %d", tc.in, got.Level(), tc.want)
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

// AdjustVolume acota en los extremos: llegar al tope con una tecla de subir
// volumen no es un error.
func TestAdjustVolume(t *testing.T) {
	tests := []struct {
		name  string
		start int
		delta int
		want  int
	}{
		{name: "sube", start: 50, delta: 10, want: 60},
		{name: "baja", start: 50, delta: -10, want: 40},
		{name: "no baja de cero", start: 5, delta: -20, want: 0},
		{name: "no pasa del máximo", start: 145, delta: 20, want: audio.MaxVolume},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := mustDevice(t, tc.start, false).AdjustVolume(tc.delta)
			if got.Volume().Level() != tc.want {
				t.Errorf("%d %+d = %d, se esperaba %d", tc.start, tc.delta, got.Volume().Level(), tc.want)
			}
		})
	}
}

// Volumen y silencio son controles independientes.
func TestVolumenYSilencioSonIndependientes(t *testing.T) {
	muted := mustDevice(t, 50, true)

	louder := muted.SetVolume(audio.ClampVolume(80))
	if !louder.Muted() {
		t.Error("subir el volumen no debe quitar el silencio")
	}
	if louder.Volume().Level() != 80 {
		t.Errorf("volumen = %d", louder.Volume().Level())
	}

	unmuted := louder.SetMuted(false)
	if unmuted.Volume().Level() != 80 {
		t.Error("quitar el silencio no debe tocar el volumen")
	}
}

func TestToggleMuted(t *testing.T) {
	device := mustDevice(t, 50, false)

	once := device.ToggleMuted()
	if !once.Muted() {
		t.Error("no silenció")
	}
	if device.Muted() {
		t.Error("mutó el original en vez de devolver copia")
	}
	if twice := once.ToggleMuted(); twice.Muted() {
		t.Error("dos veces debe volver al estado inicial")
	}
}

func TestRestore(t *testing.T) {
	id, _ := audio.NewID("x")
	name, _ := audio.NewName("X")

	t.Run("sin identificador", func(t *testing.T) {
		if _, err := audio.Restore(audio.ID{}, name, audio.Input, audio.Volume{}, false, false); !errors.Is(err, audio.ErrInvalidID) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("sin nombre", func(t *testing.T) {
		if _, err := audio.Restore(id, audio.Name{}, audio.Input, audio.Volume{}, false, false); !errors.Is(err, audio.ErrEmptyName) {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestNewID(t *testing.T) {
	if _, err := audio.NewID("   "); !errors.Is(err, audio.ErrInvalidID) {
		t.Errorf("id vacío: %v", err)
	}
	got, err := audio.NewID("  alsa_output.x  ")
	if err != nil || got.String() != "alsa_output.x" {
		t.Errorf("NewID = %q, %v", got, err)
	}
}

func TestNewName(t *testing.T) {
	if _, err := audio.NewName(""); !errors.Is(err, audio.ErrEmptyName) {
		t.Errorf("nombre vacío: %v", err)
	}
	if _, err := audio.NewName(strings.Repeat("a", audio.MaxNameLength+1)); !errors.Is(err, audio.ErrNameTooLong) {
		t.Errorf("nombre largo: %v", err)
	}
}

func TestDirectionString(t *testing.T) {
	if audio.Output.String() != "output" || audio.Input.String() != "input" {
		t.Errorf("Direction.String = %q/%q", audio.Output, audio.Input)
	}
}
