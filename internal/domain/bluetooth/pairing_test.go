package bluetooth_test

import (
	"errors"
	"testing"

	"settings-cli/internal/domain/bluetooth"
)

func TestPasskey(t *testing.T) {
	tests := []struct {
		name string
		code int
		want string
		err  error
	}{
		{name: "six digits", code: 123456, want: "123456"},
		{name: "left-pads with zeros", code: 42, want: "000042"},
		{name: "zero", code: 0, want: "000000"},
		{name: "max", code: 999999, want: "999999"},
		{name: "negative", code: -1, err: bluetooth.ErrInvalidPasskey},
		{name: "too large", code: 1000000, err: bluetooth.ErrInvalidPasskey},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := bluetooth.NewPasskey(tc.code)

			if tc.err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("error = %v, want %v", err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Present() || got.String() != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNoPasskey(t *testing.T) {
	p := bluetooth.NoPasskey()
	if p.Present() || p.String() != "" {
		t.Errorf("NoPasskey = %q/%v", p, p.Present())
	}

	// The zero value must behave the same: no code, not "000000".
	var zero bluetooth.Passkey
	if zero.Present() {
		t.Error("the zero value of Passkey must be 'no code'")
	}
}

func TestNewPairingRequest(t *testing.T) {
	address, err := bluetooth.NewAddress("AA:BB:CC:DD:EE:FF")
	if err != nil {
		t.Fatalf("NewAddress: %v", err)
	}
	name, err := bluetooth.NewName("Ana's iPhone")
	if err != nil {
		t.Fatalf("NewName: %v", err)
	}
	passkey, err := bluetooth.NewPasskey(482913)
	if err != nil {
		t.Fatalf("NewPasskey: %v", err)
	}

	t.Run("valid", func(t *testing.T) {
		got, err := bluetooth.NewPairingRequest(address, name, passkey)
		if err != nil {
			t.Fatalf("NewPairingRequest: %v", err)
		}
		if got.Address().String() != "AA:BB:CC:DD:EE:FF" || got.Passkey().String() != "482913" {
			t.Errorf("request = %s / %s", got.Address(), got.Passkey())
		}
	})

	t.Run("no address", func(t *testing.T) {
		if _, err := bluetooth.NewPairingRequest(bluetooth.Address{}, name, passkey); !errors.Is(err, bluetooth.ErrInvalidAddress) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("no name", func(t *testing.T) {
		if _, err := bluetooth.NewPairingRequest(address, bluetooth.Name{}, passkey); !errors.Is(err, bluetooth.ErrEmptyName) {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestAdapterVisibility(t *testing.T) {
	t.Run("off cannot be visible", func(t *testing.T) {
		off := bluetooth.NewAdapter(false).SetVisible(true)

		if off.Discoverable() || off.Pairable() || off.Visible() {
			t.Error("a radio that is off cannot advertise itself or accept requests")
		}
	})

	t.Run("on can", func(t *testing.T) {
		on := bluetooth.NewAdapter(true).SetVisible(true)

		if !on.Discoverable() || !on.Pairable() || !on.Visible() {
			t.Errorf("visible = %v/%v", on.Discoverable(), on.Pairable())
		}
	})

	t.Run("turning off removes visibility", func(t *testing.T) {
		off := bluetooth.NewAdapter(true).SetVisible(true).Disable()

		if off.Discoverable() || off.Pairable() {
			t.Error("turning off must leave the machine invisible")
		}

		// Turning it back on does not restore it by itself: that is a
		// separate decision.
		if on := off.Enable(); on.Visible() {
			t.Error("turning on must not restore visibility")
		}
	})

	t.Run("Visible requires both", func(t *testing.T) {
		half := bluetooth.NewAdapter(true).SetDiscoverable(true)

		if half.Visible() {
			t.Error("advertising without accepting pairings is not being visible")
		}
	})
}
