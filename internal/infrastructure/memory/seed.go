package memory

import (
	"context"
	"errors"

	"knob/internal/domain/bluetooth"
)

// fixture is one example device for the fake backend.
type fixture struct {
	address string
	name    string
	kind    bluetooth.Kind
	// connected implies paired; battery is reported only while connected.
	paired    bool
	connected bool
	battery   int
}

// fixtures is what the fake Bluetooth backend starts with: a small, stable set
// so the screen has something to show without hardware.
var fixtures = []fixture{
	{address: "AA:BB:CC:DD:EE:FF", name: "WH-1000XM4", kind: bluetooth.KindHeadphones, paired: true, connected: true, battery: 82},
	{address: "11:22:33:44:55:66", name: "MX Master 3S", kind: bluetooth.KindMouse, paired: true, connected: true, battery: 45},
	{address: "77:88:99:AA:BB:CC", name: "Keyboard K380", kind: bluetooth.KindKeyboard, paired: true},
}

// SeedDevices loads the example devices into repo. It is idempotent: an
// address already stored is left untouched. A malformed fixture is skipped, not
// returned as an error, so one bad entry cannot stop a dev run.
func SeedDevices(ctx context.Context, repo bluetooth.Repository) error {
	for _, f := range fixtures {
		device, err := f.build()
		if err != nil {
			continue
		}

		switch _, err := repo.FindByAddress(ctx, device.Address()); {
		case err == nil:
			continue
		case !errors.Is(err, bluetooth.ErrNotFound):
			return err
		}

		if err := repo.Save(ctx, device); err != nil {
			return err
		}
	}
	return nil
}

// build turns a fixture into a valid aggregate through the domain constructors
// and transitions, the same path a real backend would take.
func (f fixture) build() (bluetooth.Device, error) {
	address, err := bluetooth.NewAddress(f.address)
	if err != nil {
		return bluetooth.Device{}, err
	}
	name, err := bluetooth.NewName(f.name)
	if err != nil {
		return bluetooth.Device{}, err
	}

	device, err := bluetooth.Discover(address, name, f.kind)
	if err != nil {
		return bluetooth.Device{}, err
	}

	if f.paired || f.connected {
		if device, err = device.Pair(); err != nil {
			return bluetooth.Device{}, err
		}
	}
	if f.connected {
		if device, err = device.Connect(); err != nil {
			return bluetooth.Device{}, err
		}
	}
	if f.connected && f.battery > 0 {
		battery, err := bluetooth.NewBattery(f.battery)
		if err != nil {
			return bluetooth.Device{}, err
		}
		if device, err = device.ReportBattery(battery); err != nil {
			return bluetooth.Device{}, err
		}
	}

	return device, nil
}
