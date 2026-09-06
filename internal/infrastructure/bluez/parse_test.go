package bluez

import (
	"errors"
	"testing"

	"knob/internal/domain/bluetooth"
)

// Real output of `bluetoothctl info` for a connected headset.
const connectedHeadset = `Device AA:BB:CC:DD:EE:FF (public)
	Name: WH-1000XM4
	Alias: Living room headphones
	Class: 0x00240404 (2360324)
	Icon: audio-headset
	Paired: yes
	Bonded: yes
	Trusted: yes
	Blocked: no
	Connected: yes
	LegacyPairing: no
	UUID: Vendor specific           (00000000-deca-fade-deca-deafdecacaff)
	UUID: Audio Sink                (0000110b-0000-1000-8000-00805f9b34fb)
	Battery Percentage: 0x52 (82)
`

const pairedKeyboard = `Device 77:88:99:AA:BB:CC (public)
	Name: Keyboard K380
	Alias: Keyboard K380
	Icon: input-keyboard
	Paired: yes
	Trusted: yes
	Blocked: no
	Connected: no
	LegacyPairing: no
`

const unpairedDevice = `Device DE:AD:BE:EF:00:11 (random)
	Alias: DE-AD-BE-EF-00-11
	Paired: no
	Trusted: no
	Blocked: no
	Connected: no
`

func mustAddress(t *testing.T, s string) bluetooth.Address {
	t.Helper()
	address, err := bluetooth.NewAddress(s)
	if err != nil {
		t.Fatalf("NewAddress(%q): %v", s, err)
	}
	return address
}

func TestDeviceFromInfo(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		out         string
		wantName    string
		wantKind    bluetooth.Kind
		wantState   bluetooth.State
		wantBattery int // -1 = unknown
	}{
		{
			name: "connected with battery", address: "AA:BB:CC:DD:EE:FF", out: connectedHeadset,
			wantName: "Living room headphones", wantKind: bluetooth.KindHeadphones,
			wantState: bluetooth.StateConnected, wantBattery: 82,
		},
		{
			name: "paired but not connected", address: "77:88:99:AA:BB:CC", out: pairedKeyboard,
			wantName: "Keyboard K380", wantKind: bluetooth.KindKeyboard,
			wantState: bluetooth.StatePaired, wantBattery: -1,
		},
		{
			name: "visible but not paired", address: "DE:AD:BE:EF:00:11", out: unpairedDevice,
			wantName: "DE-AD-BE-EF-00-11", wantKind: bluetooth.KindUnknown,
			wantState: bluetooth.StateDiscovered, wantBattery: -1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := deviceFromInfo(mustAddress(t, tc.address), tc.out)
			if err != nil {
				t.Fatalf("deviceFromInfo: %v", err)
			}

			if got.Name().String() != tc.wantName {
				t.Errorf("name = %q, want %q", got.Name(), tc.wantName)
			}
			if got.Kind() != tc.wantKind {
				t.Errorf("kind = %v, want %v", got.Kind(), tc.wantKind)
			}
			if got.State() != tc.wantState {
				t.Errorf("state = %v, want %v", got.State(), tc.wantState)
			}

			battery := got.Battery()
			if tc.wantBattery < 0 {
				if battery.Known() {
					t.Errorf("battery = %d%%, want unknown", battery.Level())
				}
				return
			}
			if !battery.Known() || battery.Level() != tc.wantBattery {
				t.Errorf("battery = %d/%v, want %d%%", battery.Level(), battery.Known(), tc.wantBattery)
			}
		})
	}
}

// Alias beats Name: it is the name the user has set.
func TestDeviceFromInfoPrefersAlias(t *testing.T) {
	got, err := deviceFromInfo(mustAddress(t, "AA:BB:CC:DD:EE:FF"), connectedHeadset)
	if err != nil {
		t.Fatalf("deviceFromInfo: %v", err)
	}
	if got.Name().String() == "WH-1000XM4" {
		t.Error("kept Name when Alias was available")
	}
}

func TestDeviceFromInfoNotFound(t *testing.T) {
	out := "Device AA:BB:CC:DD:EE:FF not available\n"
	if _, err := deviceFromInfo(mustAddress(t, "AA:BB:CC:DD:EE:FF"), out); !errors.Is(err, bluetooth.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestParseBattery(t *testing.T) {
	tests := []struct {
		raw   string
		level int // -1 = unknown
	}{
		{"0x52 (82)", 82},
		{"0x64 (100)", 100},
		{"0x00 (0)", 0},
		{"", -1},
		{"0x52", -1},
		{"(abc)", -1},
		{"0xff (255)", -1}, // outside the domain range
	}

	for _, tc := range tests {
		t.Run(tc.raw, func(t *testing.T) {
			got := parseBattery(tc.raw)

			if tc.level < 0 {
				if got.Known() {
					t.Errorf("parseBattery(%q) = %d%%, want unknown", tc.raw, got.Level())
				}
				return
			}
			if !got.Known() || got.Level() != tc.level {
				t.Errorf("parseBattery(%q) = %d/%v", tc.raw, got.Level(), got.Known())
			}
		})
	}
}

func TestParseDeviceLine(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"Device AA:BB:CC:DD:EE:FF WH-1000XM4", "AA:BB:CC:DD:EE:FF"},
		{"Device aa:bb:cc:dd:ee:ff nameless with spaces", "AA:BB:CC:DD:EE:FF"},
		{"Controller 50:EE:32:81:D5:42 arch-pc [default]", ""},
		{"", ""},
		{"Device not-a-mac X", ""},
	}

	for _, tc := range tests {
		t.Run(tc.line, func(t *testing.T) {
			got, ok := parseDeviceLine(tc.line)

			if tc.want == "" {
				if ok {
					t.Errorf("accepted %q as a device", tc.line)
				}
				return
			}
			if !ok || got.String() != tc.want {
				t.Errorf("parseDeviceLine(%q) = %q/%v", tc.line, got, ok)
			}
		})
	}
}

// Repeated keys (UUID) must not overwrite the first value.
func TestParseFieldsIgnoresRepeats(t *testing.T) {
	fields := parseFields(connectedHeadset)

	if fields["Icon"] != "audio-headset" {
		t.Errorf("Icon = %q", fields["Icon"])
	}
	if !yes(fields, "Connected") || !yes(fields, "Paired") {
		t.Error("did not read Connected/Paired")
	}
	if yes(fields, "Blocked") {
		t.Error("Blocked: no was read as yes")
	}
	if fields["UUID"] != "Vendor specific           (00000000-deca-fade-deca-deafdecacaff)" {
		t.Errorf("a repeated key overwrote the first: %q", fields["UUID"])
	}
}
