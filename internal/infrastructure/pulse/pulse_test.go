package pulse

import (
	"testing"

	"settings-cli/internal/domain/audio"
)

// Real output of `pactl -f json list sinks`, trimmed.
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

// On a source, an empty monitor_source means a real input.
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

func TestOutputs(t *testing.T) {
	got := devicesFrom(t, sinksJSON, audio.Output, "alsa_output.pci-0000_01_00.1.hdmi-stereo")

	if len(got) != 2 {
		t.Fatalf("read %d outputs, want 2", len(got))
	}

	t.Run("volume", func(t *testing.T) {
		// base_volume is always 100%: reading it by mistake would set everything to 100%.
		if got[0].Volume().Level() != 45 || got[1].Volume().Level() != 50 {
			t.Errorf("volumes = %d%%/%d%%", got[0].Volume().Level(), got[1].Volume().Level())
		}
	})

	t.Run("default", func(t *testing.T) {
		if !got[0].IsDefault() || got[1].IsDefault() {
			t.Errorf("defaults = %v/%v", got[0].IsDefault(), got[1].IsDefault())
		}
	})

	t.Run("mute", func(t *testing.T) {
		if got[0].Muted() || !got[1].Muted() {
			t.Errorf("mutes = %v/%v", got[0].Muted(), got[1].Muted())
		}
	})

	t.Run("readable name", func(t *testing.T) {
		if got[1].Name().String() != "Built-in Audio Analog Stereo" {
			t.Errorf("name = %q", got[1].Name())
		}
	})
}

// Monitors are the copy of an output, not real microphones.
func TestMicrophonesDropMonitors(t *testing.T) {
	got := devicesFrom(t, sourcesJSON, audio.Input, "alsa_input.usb-Maono_DGM20_USB_Microphone_20230101-00.analog-stereo")

	if len(got) != 1 {
		t.Fatalf("read %d microphones, want 1", len(got))
	}
	if got[0].Name().String() != "DGM20 USB Microphone Analog Stereo" {
		t.Errorf("name = %q", got[0].Name())
	}
	if !got[0].IsDefault() || got[0].Direction() != audio.Input {
		t.Errorf("microphone = %v/%v", got[0].IsDefault(), got[0].Direction())
	}
}

// A sink with a filled monitor_source is still a valid output: that key only
// drops things when reading inputs.
func TestMonitorSourceDoesNotDropOutputs(t *testing.T) {
	if got := devicesFrom(t, sinksJSON, audio.Output, ""); len(got) != 2 {
		t.Fatalf("outputs dropped for having a monitor: %d remain", len(got))
	}
}

// Go's map iteration is random: keeping any one channel would give a different
// level on each read.
func TestVolumeOfTakesTheMaxAndIsDeterministic(t *testing.T) {
	channels := map[string]jsonChannel{
		"front-left":  {ValuePercent: "30%"},
		"front-right": {ValuePercent: "80%"},
		"lfe":         {ValuePercent: "55%"},
	}

	for i := 0; i < 50; i++ {
		got, ok := volumeOf(channels)
		if !ok || got.Level() != 80 {
			t.Fatalf("volumeOf = %d/%v on iteration %d", got.Level(), ok, i)
		}
	}
}

func TestVolumeOfWithNoChannels(t *testing.T) {
	if _, ok := volumeOf(nil); ok {
		t.Error("with no channels there must be no volume")
	}
	if _, ok := volumeOf(map[string]jsonChannel{"x": {ValuePercent: "weird"}}); ok {
		t.Error("an unreadable percentage must not yield a volume")
	}
}

func TestVolumeAboveTheMaxIsClamped(t *testing.T) {
	raw := `[{"name":"weird","description":"Cranks up","mute":false,"volume":{"mono":{"value_percent":"200%"}}}]`

	got := devicesFrom(t, raw, audio.Output, "")
	if len(got) != 1 {
		t.Fatalf("the device was dropped instead of clamped")
	}
	if got[0].Volume().Level() != audio.MaxVolume {
		t.Errorf("volume = %d%%, want %d%%", got[0].Volume().Level(), audio.MaxVolume)
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
		{"weird%", 0, false},
	}

	for _, tc := range tests {
		got, ok := parsePercent(tc.in)
		if ok != tc.found || got != tc.want {
			t.Errorf("parsePercent(%q) = %d/%v, want %d/%v", tc.in, got, ok, tc.want, tc.found)
		}
	}
}

func TestDecodeDevicesRejectsGarbage(t *testing.T) {
	if _, err := decodeDevices([]byte("this is not json")); err == nil {
		t.Fatal("accepted an unreadable response")
	}
}
