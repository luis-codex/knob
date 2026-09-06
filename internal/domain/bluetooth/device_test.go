package bluetooth_test

import (
	"errors"
	"strings"
	"testing"

	"knob/internal/domain/bluetooth"
	"knob/internal/domain/errs"
)

func mustDevice(t *testing.T) bluetooth.Device {
	t.Helper()
	addr, err := bluetooth.NewAddress("AA:BB:CC:DD:EE:FF")
	if err != nil {
		t.Fatalf("NewAddress: %v", err)
	}
	name, err := bluetooth.NewName("WH-1000XM4")
	if err != nil {
		t.Fatalf("NewName: %v", err)
	}
	device, err := bluetooth.Discover(addr, name, bluetooth.KindHeadphones)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	return device
}

// paired and connected return a device already in that state.
func paired(t *testing.T) bluetooth.Device {
	t.Helper()
	d, err := mustDevice(t).Pair()
	if err != nil {
		t.Fatalf("Pair: %v", err)
	}
	return d
}

func connected(t *testing.T) bluetooth.Device {
	t.Helper()
	d, err := paired(t).Connect()
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	return d
}

func TestNewAddress(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		err   error
	}{
		{name: "normalizes to uppercase", input: "aa:bb:cc:dd:ee:ff", want: "AA:BB:CC:DD:EE:FF"},
		{name: "accepts hyphens", input: "AA-BB-CC-DD-EE-FF", want: "AA:BB:CC:DD:EE:FF"},
		{name: "trims spaces", input: "  AA:BB:CC:DD:EE:FF  ", want: "AA:BB:CC:DD:EE:FF"},
		{name: "empty", input: "", err: bluetooth.ErrInvalidAddress},
		{name: "too few octets", input: "AA:BB:CC:DD:EE", err: bluetooth.ErrInvalidAddress},
		{name: "too many octets", input: "AA:BB:CC:DD:EE:FF:00", err: bluetooth.ErrInvalidAddress},
		{name: "short octet", input: "AA:BB:CC:DD:EE:F", err: bluetooth.ErrInvalidAddress},
		{name: "non-hex character", input: "AA:BB:CC:DD:EE:GG", err: bluetooth.ErrInvalidAddress},
		{name: "no separators", input: "AABBCCDDEEFF", err: bluetooth.ErrInvalidAddress},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := bluetooth.NewAddress(tc.input)

			if tc.err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("error = %v, want %v", err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

// The same address written two ways must be the same Address.
func TestAddressEqualsAfterNormalizing(t *testing.T) {
	a, _ := bluetooth.NewAddress("aa-bb-cc-dd-ee-ff")
	b, _ := bluetooth.NewAddress("AA:BB:CC:DD:EE:FF")

	if !a.Equals(b) {
		t.Errorf("%q and %q should be equal", a, b)
	}
}

func TestNewName(t *testing.T) {
	if _, err := bluetooth.NewName("  "); !errors.Is(err, bluetooth.ErrEmptyName) {
		t.Errorf("empty name: %v", err)
	}
	if _, err := bluetooth.NewName(strings.Repeat("a", bluetooth.MaxNameLength+1)); !errors.Is(err, bluetooth.ErrNameTooLong) {
		t.Errorf("long name: %v", err)
	}
	got, err := bluetooth.NewName("  Keyboard K380 ")
	if err != nil || got.String() != "Keyboard K380" {
		t.Errorf("NewName = %q, %v", got, err)
	}
}

func TestBattery(t *testing.T) {
	t.Run("unknown is the zero value", func(t *testing.T) {
		var zero bluetooth.Battery
		if zero.Known() {
			t.Error("the zero value must be unknown, not 0%")
		}
		if bluetooth.UnknownBattery().Known() {
			t.Error("UnknownBattery must be unknown")
		}
	})

	t.Run("valid range", func(t *testing.T) {
		for _, level := range []int{0, 50, 100} {
			b, err := bluetooth.NewBattery(level)
			if err != nil {
				t.Fatalf("NewBattery(%d): %v", level, err)
			}
			if !b.Known() || b.Level() != level {
				t.Errorf("NewBattery(%d) = %d/%v", level, b.Level(), b.Known())
			}
		}
	})

	t.Run("out of range", func(t *testing.T) {
		for _, level := range []int{-1, 101} {
			if _, err := bluetooth.NewBattery(level); !errors.Is(err, bluetooth.ErrInvalidBattery) {
				t.Errorf("NewBattery(%d) = %v", level, err)
			}
		}
	})
}

func TestDiscoverStartsUnpaired(t *testing.T) {
	d := mustDevice(t)

	if d.State() != bluetooth.StateDiscovered {
		t.Errorf("state = %v, want discovered", d.State())
	}
	if d.Battery().Known() {
		t.Error("a just-seen device reports no battery")
	}
}

// The full lifecycle walk.
func TestLifecycle(t *testing.T) {
	d := mustDevice(t)

	d, err := d.Pair()
	if err != nil || d.State() != bluetooth.StatePaired {
		t.Fatalf("Pair = %v, %v", d.State(), err)
	}

	d, err = d.Connect()
	if err != nil || d.State() != bluetooth.StateConnected {
		t.Fatalf("Connect = %v, %v", d.State(), err)
	}

	d, err = d.Disconnect()
	if err != nil || d.State() != bluetooth.StatePaired {
		t.Fatalf("Disconnect = %v, %v", d.State(), err)
	}

	d, err = d.Unpair()
	if err != nil || d.State() != bluetooth.StateDiscovered {
		t.Fatalf("Unpair = %v, %v", d.State(), err)
	}
}

// Impossible transitions must be conflicts, not invalid data.
func TestInvalidTransitions(t *testing.T) {
	tests := []struct {
		name string
		do   func(*testing.T) (bluetooth.Device, error)
		err  error
	}{
		{"connect without pairing", func(t *testing.T) (bluetooth.Device, error) { return mustDevice(t).Connect() }, bluetooth.ErrNotPaired},
		{"disconnect without connecting", func(t *testing.T) (bluetooth.Device, error) { return paired(t).Disconnect() }, bluetooth.ErrNotConnected},
		{"pair twice", func(t *testing.T) (bluetooth.Device, error) { return paired(t).Pair() }, bluetooth.ErrAlreadyPaired},
		{"pair while connected", func(t *testing.T) (bluetooth.Device, error) { return connected(t).Pair() }, bluetooth.ErrAlreadyPaired},
		{"connect twice", func(t *testing.T) (bluetooth.Device, error) { return connected(t).Connect() }, bluetooth.ErrAlreadyConnected},
		{"unpair without pairing", func(t *testing.T) (bluetooth.Device, error) { return mustDevice(t).Unpair() }, bluetooth.ErrNotPaired},
		{"battery while disconnected", func(t *testing.T) (bluetooth.Device, error) {
			b, _ := bluetooth.NewBattery(50)
			return paired(t).ReportBattery(b)
		}, bluetooth.ErrNotConnected},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.do(t)

			if !errors.Is(err, tc.err) {
				t.Fatalf("error = %v, want %v", err, tc.err)
			}
			if !errs.IsConflict(err) {
				t.Error("an impossible transition must be a conflict, not invalid data")
			}
			if !got.IsZero() {
				t.Error("a failed transition must not return a device")
			}
		})
	}
}

// Unpairing disconnects along the way and forgets the battery.
func TestUnpairFromConnected(t *testing.T) {
	b, _ := bluetooth.NewBattery(80)
	d, err := connected(t).ReportBattery(b)
	if err != nil {
		t.Fatalf("ReportBattery: %v", err)
	}

	d, err = d.Unpair()
	if err != nil {
		t.Fatalf("Unpair: %v", err)
	}
	if d.State() != bluetooth.StateDiscovered {
		t.Errorf("state = %v", d.State())
	}
	if d.Battery().Known() {
		t.Error("forgetting the device makes the battery unknown again")
	}
}

// Transitions return copies: the original does not change.
func TestTransitionsDoNotMutateOriginal(t *testing.T) {
	original := mustDevice(t)

	if _, err := original.Pair(); err != nil {
		t.Fatalf("Pair: %v", err)
	}
	if original.State() != bluetooth.StateDiscovered {
		t.Errorf("the original mutated to %v", original.State())
	}
}

func TestParseKind(t *testing.T) {
	for _, kind := range []bluetooth.Kind{
		bluetooth.KindUnknown, bluetooth.KindHeadphones, bluetooth.KindSpeaker,
		bluetooth.KindMouse, bluetooth.KindKeyboard, bluetooth.KindPhone,
	} {
		got, err := bluetooth.ParseKind(kind.String())
		if err != nil {
			t.Fatalf("ParseKind(%q): %v", kind, err)
		}
		if got != kind {
			t.Errorf("round trip of %v gave %v", kind, got)
		}
	}

	if _, err := bluetooth.ParseKind("toaster"); !errors.Is(err, bluetooth.ErrUnknownKind) {
		t.Errorf("unknown kind: %v", err)
	}
}

func TestRestoreKeepsState(t *testing.T) {
	addr, _ := bluetooth.NewAddress("11:22:33:44:55:66")
	name, _ := bluetooth.NewName("MX Master")
	battery, _ := bluetooth.NewBattery(45)

	d, err := bluetooth.Restore(addr, name, bluetooth.KindMouse, bluetooth.StateConnected, battery)
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}

	if d.State() != bluetooth.StateConnected || d.Battery().Level() != 45 {
		t.Errorf("Restore returned %v / %d%%", d.State(), d.Battery().Level())
	}

	if _, err := bluetooth.Restore(bluetooth.Address{}, name, bluetooth.KindMouse, bluetooth.StatePaired, battery); !errors.Is(err, bluetooth.ErrInvalidAddress) {
		t.Errorf("no address: %v", err)
	}
}
