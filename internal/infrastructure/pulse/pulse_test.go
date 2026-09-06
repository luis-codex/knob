package pulse

import (
	"testing"

	"settings-cli/internal/domain/audio"
)

// Salida real de `pactl -f json list sinks`, recortada.
const sinksJSON = `[
  {
    "index": 55,
    "name": "alsa_output.pci-0000_01_00.1.hdmi-stereo",
    "description": "GB206 High Definition Audio Controller Digital Stereo (HDMI)",
    "mute": false,
    "monitor_source": "alsa_output.pci-0000_01_00.1.hdmi-stereo.monitor",
    "volume": {
      "front-left":  {"value": 29489, "value_percent": "45%", "db": "-20.81 dB"},
      "front-right": {"value": 29489, "value_percent": "45%", "db": "-20.81 dB"}
    },
    "base_volume": {"value": 65536, "value_percent": "100%", "db": "0.00 dB"}
  },
  {
    "index": 58,
    "name": "alsa_output.pci-0000_00_1f.3.analog-stereo",
    "description": "Built-in Audio Analog Stereo",
    "mute": true,
    "monitor_source": "alsa_output.pci-0000_00_1f.3.analog-stereo.monitor",
    "volume": {
      "front-left":  {"value": 32766, "value_percent": "50%", "db": "-18.06 dB"},
      "front-right": {"value": 32766, "value_percent": "50%", "db": "-18.06 dB"}
    }
  }
]`

// En una fuente, monitor_source vacío significa entrada real.
const sourcesJSON = `[
  {
    "name": "alsa_output.pci-0000_01_00.1.hdmi-stereo.monitor",
    "description": "Monitor of GB206 Digital Stereo (HDMI)",
    "mute": false,
    "monitor_source": "alsa_output.pci-0000_01_00.1.hdmi-stereo",
    "volume": {"front-left": {"value_percent": "100%"}}
  },
  {
    "name": "alsa_input.usb-Maono_DGM20_USB_Microphone_20230101-00.analog-stereo",
    "description": "DGM20 USB Microphone Analog Stereo",
    "mute": false,
    "monitor_source": "",
    "volume": {"front-left": {"value_percent": "100%"}, "front-right": {"value_percent": "100%"}}
  }
]`

func devicesFrom(t *testing.T, raw string, direction audio.Direction, defaultName string) []audio.Device {
	t.Helper()

	decoded, err := decodeDevices([]byte(raw))
	if err != nil {
		t.Fatalf("decodeDevices: %v", err)
	}

	devices := make([]audio.Device, 0, len(decoded))
	for _, d := range decoded {
		if device, ok := toDevice(d, direction, defaultName); ok {
			devices = append(devices, device)
		}
	}
	return devices
}

func TestSalidas(t *testing.T) {
	got := devicesFrom(t, sinksJSON, audio.Output, "alsa_output.pci-0000_01_00.1.hdmi-stereo")

	if len(got) != 2 {
		t.Fatalf("se leyeron %d salidas, se esperaban 2", len(got))
	}

	t.Run("volumen", func(t *testing.T) {
		// base_volume es siempre 100%: leerlo por error pondría todo al 100%.
		if got[0].Volume().Level() != 45 || got[1].Volume().Level() != 50 {
			t.Errorf("volúmenes = %d%%/%d%%", got[0].Volume().Level(), got[1].Volume().Level())
		}
	})

	t.Run("predeterminado", func(t *testing.T) {
		if !got[0].IsDefault() || got[1].IsDefault() {
			t.Errorf("predeterminados = %v/%v", got[0].IsDefault(), got[1].IsDefault())
		}
	})

	t.Run("silencio", func(t *testing.T) {
		if got[0].Muted() || !got[1].Muted() {
			t.Errorf("silencios = %v/%v", got[0].Muted(), got[1].Muted())
		}
	})

	t.Run("nombre legible", func(t *testing.T) {
		if got[1].Name().String() != "Built-in Audio Analog Stereo" {
			t.Errorf("nombre = %q", got[1].Name())
		}
	})
}

// Los monitores son la copia de una salida, no micrófonos reales.
func TestMicrofonosDescartanMonitores(t *testing.T) {
	got := devicesFrom(t, sourcesJSON, audio.Input, "alsa_input.usb-Maono_DGM20_USB_Microphone_20230101-00.analog-stereo")

	if len(got) != 1 {
		t.Fatalf("se leyeron %d micrófonos, se esperaba 1", len(got))
	}
	if got[0].Name().String() != "DGM20 USB Microphone Analog Stereo" {
		t.Errorf("nombre = %q", got[0].Name())
	}
	if !got[0].IsDefault() || got[0].Direction() != audio.Input {
		t.Errorf("micrófono = %v/%v", got[0].IsDefault(), got[0].Direction())
	}
}

// Un sink con monitor_source relleno sigue siendo una salida válida: esa clave
// solo descarta cuando se leen entradas.
func TestMonitorSourceNoDescartaSalidas(t *testing.T) {
	if got := devicesFrom(t, sinksJSON, audio.Output, ""); len(got) != 2 {
		t.Fatalf("se descartaron salidas por tener monitor: quedaron %d", len(got))
	}
}

// El recorrido de un mapa en Go es aleatorio: quedarse con un canal
// cualquiera daría un nivel distinto en cada lectura.
func TestVolumeOfTomaElMaximoYEsDeterminista(t *testing.T) {
	channels := map[string]jsonChannel{
		"front-left":  {ValuePercent: "30%"},
		"front-right": {ValuePercent: "80%"},
		"lfe":         {ValuePercent: "55%"},
	}

	for i := 0; i < 50; i++ {
		got, ok := volumeOf(channels)
		if !ok || got.Level() != 80 {
			t.Fatalf("volumeOf = %d/%v en la iteración %d", got.Level(), ok, i)
		}
	}
}

func TestVolumeOfSinCanales(t *testing.T) {
	if _, ok := volumeOf(nil); ok {
		t.Error("sin canales no debe haber volumen")
	}
	if _, ok := volumeOf(map[string]jsonChannel{"x": {ValuePercent: "raro"}}); ok {
		t.Error("un porcentaje ilegible no debe dar volumen")
	}
}

func TestVolumenPorEncimaDelMaximoSeAcota(t *testing.T) {
	raw := `[{"name":"raro","description":"Sube mucho","mute":false,"volume":{"mono":{"value_percent":"200%"}}}]`

	got := devicesFrom(t, raw, audio.Output, "")
	if len(got) != 1 {
		t.Fatalf("se descartó el dispositivo en vez de acotarlo")
	}
	if got[0].Volume().Level() != audio.MaxVolume {
		t.Errorf("volumen = %d%%, se esperaba %d%%", got[0].Volume().Level(), audio.MaxVolume)
	}
}

func TestParsePercent(t *testing.T) {
	tests := []struct {
		in    string
		want  int
		found bool
	}{
		{"45%", 45, true},
		{"100%", 100, true},
		{" 0% ", 0, true},
		{"45", 0, false},
		{"", 0, false},
		{"raro%", 0, false},
	}

	for _, tc := range tests {
		got, ok := parsePercent(tc.in)
		if ok != tc.found || got != tc.want {
			t.Errorf("parsePercent(%q) = %d/%v, se esperaba %d/%v", tc.in, got, ok, tc.want, tc.found)
		}
	}
}

func TestDecodeDevicesRechazaBasura(t *testing.T) {
	if _, err := decodeDevices([]byte("esto no es json")); err == nil {
		t.Fatal("aceptó una respuesta ilegible")
	}
}
