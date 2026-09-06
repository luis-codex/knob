package bluez

import (
	"bufio"
	"context"
	"strings"

	"knob/internal/domain/bluetooth"
	"knob/internal/domain/errs"
)

// Repository exposes the devices BlueZ knows about.
type Repository struct{}

var _ bluetooth.Repository = (*Repository)(nil)

func NewRepository() *Repository { return &Repository{} }

// List enumerates the known devices and queries the detail of each one.
// That is N+1 invocations, acceptable for a list of Bluetooth devices.
func (r *Repository) List(ctx context.Context) ([]bluetooth.Device, error) {
	out, err := run(ctx, "devices")
	if err != nil {
		return nil, err
	}

	var devices []bluetooth.Device
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		address, ok := parseDeviceLine(scanner.Text())
		if !ok {
			continue
		}

		device, err := r.FindByAddress(ctx, address)
		if err != nil {
			// A device that disappears between the two calls must not bring
			// down the whole listing.
			continue
		}
		devices = append(devices, device)
	}
	return devices, nil
}

// parseDeviceLine reads "Device AA:BB:CC:DD:EE:FF Name".
func parseDeviceLine(line string) (bluetooth.Address, bool) {
	parts := strings.Fields(strings.TrimSpace(line))
	if len(parts) < 2 || parts[0] != "Device" {
		return bluetooth.Address{}, false
	}

	address, err := bluetooth.NewAddress(parts[1])
	if err != nil {
		return bluetooth.Address{}, false
	}
	return address, true
}

func (r *Repository) FindByAddress(ctx context.Context, address bluetooth.Address) (bluetooth.Device, error) {
	out, err := run(ctx, "info", address.String())
	if err != nil {
		return bluetooth.Device{}, err
	}
	return deviceFromInfo(address, out)
}

// deviceFromInfo translates the output of `bluetoothctl info` to the domain.
// It is separate from the invocation so it can be tested without hardware.
func deviceFromInfo(address bluetooth.Address, out string) (bluetooth.Device, error) {
	if strings.Contains(out, "not available") {
		return bluetooth.Device{}, bluetooth.ErrNotFound
	}

	fields := parseFields(out)

	// Alias is the name the user sees and can change; Name is the one the
	// device advertises. Alias is preferred, with Name as a fallback and the
	// address as a last resort: a device with no name must not bring down the
	// listing.
	name, err := bluetooth.NewName(firstNonEmpty(fields["Alias"], fields["Name"], address.String()))
	if err != nil {
		return bluetooth.Device{}, err
	}

	return bluetooth.Restore(
		address,
		name,
		parseKind(fields["Icon"]),
		parseState(fields),
		parseBattery(fields["Battery Percentage"]),
	)
}

func parseState(fields map[string]string) bluetooth.State {
	switch {
	case yes(fields, "Connected"):
		return bluetooth.StateConnected
	case yes(fields, "Paired"):
		return bluetooth.StatePaired
	default:
		return bluetooth.StateDiscovered
	}
}

// Save reconciles: it compares the desired state with the real one and emits
// whatever commands are needed.
//
// This is the deep difference from a store. BlueZ does not save what you pass
// it; you have to *make* reality match.
func (r *Repository) Save(ctx context.Context, d bluetooth.Device) error {
	current, err := r.FindByAddress(ctx, d.Address())
	if err != nil {
		return err
	}
	if current.State() == d.State() {
		return r.saveAlias(ctx, d, current)
	}

	address := d.Address().String()

	switch d.State() {
	case bluetooth.StateConnected:
		if !current.State().IsPaired() {
			if _, err := run(ctx, "pair", address); err != nil {
				return err
			}
		}
		_, err = run(ctx, "connect", address)

	case bluetooth.StatePaired:
		if current.State() == bluetooth.StateConnected {
			_, err = run(ctx, "disconnect", address)
			break
		}
		_, err = run(ctx, "pair", address)

	case bluetooth.StateDiscovered:
		// Forget: BlueZ disconnects on its own when the pairing is removed.
		_, err = run(ctx, "remove", address)
	}

	if err != nil {
		return err
	}
	return r.saveAlias(ctx, d, current)
}

// saveAlias is not supported by bluetoothctl, which does not expose the Alias
// property. Changing the name would require talking to D-Bus directly.
func (r *Repository) saveAlias(_ context.Context, d, current bluetooth.Device) error {
	if d.Name().String() == current.Name().String() {
		return nil
	}
	return errs.Conflict("renaming devices is not supported with BlueZ")
}

func (r *Repository) Delete(ctx context.Context, address bluetooth.Address) error {
	out, err := run(ctx, "remove", address.String())
	if err != nil {
		return err
	}
	if strings.Contains(out, "not available") {
		return bluetooth.ErrNotFound
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
