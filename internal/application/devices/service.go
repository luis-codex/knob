// Package devices holds the Bluetooth-device use cases.
//
// It is the domain's boundary: it takes primitives from the interface, turns
// them into value objects and orchestrates the repository.
//
// It is named devices (plural) rather than bluetooth so it does not clash with
// the domain package.
package devices

import (
	"context"
	"errors"
	"strings"

	"knob/internal/domain/bluetooth"
)

// transition is an aggregate operation that returns the resulting device. It
// matches the signature of Pair, Connect and friends, which are passed as
// method expressions.
type transition func(bluetooth.Device) (bluetooth.Device, error)

// Service groups the use cases over devices and the adapter.
//
// The rule that you cannot pair or connect while the radio is off lives here,
// not in the domain: it crosses two aggregates, and neither can know the other
// without coupling them.
type Service struct {
	repo     bluetooth.Repository
	adapters bluetooth.AdapterRepository
	scanner  bluetooth.Scanner
}

// NewService takes the ports, not concrete implementations.
func NewService(repo bluetooth.Repository, adapters bluetooth.AdapterRepository, scanner bluetooth.Scanner) *Service {
	return &Service{repo: repo, adapters: adapters, scanner: scanner}
}

// Scan looks for nearby devices and records the ones that were not known.
// It returns only the new ones: the already-known devices are left untouched
// so their state and name are not overwritten.
func (s *Service) Scan(ctx context.Context) ([]bluetooth.Device, error) {
	if err := s.requireAdapter(ctx); err != nil {
		return nil, err
	}

	// What is already known is checked BEFORE scanning. With a real backend
	// the scan itself makes the repository come to know what it finds, so
	// asking it afterwards would always say "already known" and there would
	// never be anything new.
	known, err := s.knownAddresses(ctx)
	if err != nil {
		return nil, err
	}

	found, err := s.scanner.Scan(ctx)
	if err != nil {
		return nil, err
	}

	var added []bluetooth.Device
	for _, device := range found {
		if _, ok := known[device.Address().String()]; ok {
			continue
		}

		if err := s.repo.Save(ctx, device); err != nil {
			return nil, err
		}
		added = append(added, device)
	}
	return added, nil
}

func (s *Service) knownAddresses(ctx context.Context) (map[string]struct{}, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	known := make(map[string]struct{}, len(all))
	for _, d := range all {
		known[d.Address().String()] = struct{}{}
	}
	return known, nil
}

// isKnown tells "does not exist" apart from "the repository failed".
func (s *Service) isKnown(ctx context.Context, address bluetooth.Address) (bool, error) {
	switch _, err := s.repo.FindByAddress(ctx, address); {
	case err == nil:
		return true, nil
	case errors.Is(err, bluetooth.ErrNotFound):
		return false, nil
	default:
		return false, err
	}
}

// Adapter returns the radio's state.
func (s *Service) Adapter(ctx context.Context) (bluetooth.Adapter, error) {
	return s.adapters.Get(ctx)
}

// EnableAdapter turns the radio on. It reconnects nothing: reconnecting is the
// user's decision, not an effect of turning on.
func (s *Service) EnableAdapter(ctx context.Context) (bluetooth.Adapter, error) {
	adapter, err := s.adapters.Get(ctx)
	if err != nil {
		return bluetooth.Adapter{}, err
	}

	next := adapter.Enable()
	if err := s.adapters.Save(ctx, next); err != nil {
		return bluetooth.Adapter{}, err
	}
	return next, nil
}

// DisableAdapter turns the radio off and drops the active connections.
//
// Devices are disconnected before the adapter is saved: if something failed
// halfway, the saved state would still say "on", which is more faithful than
// the opposite.
func (s *Service) DisableAdapter(ctx context.Context) (bluetooth.Adapter, error) {
	adapter, err := s.adapters.Get(ctx)
	if err != nil {
		return bluetooth.Adapter{}, err
	}

	if err := s.disconnectAll(ctx); err != nil {
		return bluetooth.Adapter{}, err
	}

	next := adapter.Disable()
	if err := s.adapters.Save(ctx, next); err != nil {
		return bluetooth.Adapter{}, err
	}
	return next, nil
}

func (s *Service) disconnectAll(ctx context.Context) error {
	all, err := s.repo.List(ctx)
	if err != nil {
		return err
	}

	for _, d := range all {
		if d.State() != bluetooth.StateConnected {
			continue
		}

		next, err := d.Disconnect()
		if err != nil {
			return err
		}
		if err := s.repo.Save(ctx, next); err != nil {
			return err
		}
	}
	return nil
}

// requireAdapter fails if the radio is off.
func (s *Service) requireAdapter(ctx context.Context) error {
	adapter, err := s.adapters.Get(ctx)
	if err != nil {
		return err
	}
	if !adapter.Enabled() {
		return bluetooth.ErrAdapterDisabled
	}
	return nil
}

// apply loads the device, applies the transition to it and saves the result.
// Every state operation follows this same path.
func (s *Service) apply(ctx context.Context, address string, change transition) (bluetooth.Device, error) {
	addr, err := bluetooth.NewAddress(address)
	if err != nil {
		return bluetooth.Device{}, err
	}

	current, err := s.repo.FindByAddress(ctx, addr)
	if err != nil {
		return bluetooth.Device{}, err
	}

	next, err := change(current)
	if err != nil {
		return bluetooth.Device{}, err
	}

	if err := s.repo.Save(ctx, next); err != nil {
		return bluetooth.Device{}, err
	}
	return next, nil
}

// Discover records a new device, not yet paired.
func (s *Service) Discover(ctx context.Context, address, name, kind string) (bluetooth.Device, error) {
	addr, err := bluetooth.NewAddress(address)
	if err != nil {
		return bluetooth.Device{}, err
	}

	deviceName, err := bluetooth.NewName(name)
	if err != nil {
		return bluetooth.Device{}, err
	}

	deviceKind, err := bluetooth.ParseKind(kind)
	if err != nil {
		return bluetooth.Device{}, err
	}

	// Recording the same address twice is a conflict, not a create: a silent
	// create would lose the state of the already-known device.
	known, err := s.isKnown(ctx, addr)
	if err != nil {
		return bluetooth.Device{}, err
	}
	if known {
		return bluetooth.Device{}, bluetooth.ErrAlreadyKnown
	}

	device, err := bluetooth.Discover(addr, deviceName, deviceKind)
	if err != nil {
		return bluetooth.Device{}, err
	}

	if err := s.repo.Save(ctx, device); err != nil {
		return bluetooth.Device{}, err
	}
	return device, nil
}

// Pair and Connect require the radio to be on. The other operations are
// management and work the same with Bluetooth off, as in any settings panel.
func (s *Service) Pair(ctx context.Context, address string) (bluetooth.Device, error) {
	if err := s.requireAdapter(ctx); err != nil {
		return bluetooth.Device{}, err
	}
	return s.apply(ctx, address, bluetooth.Device.Pair)
}

func (s *Service) Unpair(ctx context.Context, address string) (bluetooth.Device, error) {
	return s.apply(ctx, address, bluetooth.Device.Unpair)
}

func (s *Service) Connect(ctx context.Context, address string) (bluetooth.Device, error) {
	if err := s.requireAdapter(ctx); err != nil {
		return bluetooth.Device{}, err
	}
	return s.apply(ctx, address, bluetooth.Device.Connect)
}

func (s *Service) Disconnect(ctx context.Context, address string) (bluetooth.Device, error) {
	return s.apply(ctx, address, bluetooth.Device.Disconnect)
}

// Rename changes the device's visible name.
func (s *Service) Rename(ctx context.Context, address, name string) (bluetooth.Device, error) {
	newName, err := bluetooth.NewName(name)
	if err != nil {
		return bluetooth.Device{}, err
	}

	return s.apply(ctx, address, func(d bluetooth.Device) (bluetooth.Device, error) {
		return d.Rename(newName)
	})
}

// ReportBattery records the charge level reported by the device.
func (s *Service) ReportBattery(ctx context.Context, address string, level int) (bluetooth.Device, error) {
	battery, err := bluetooth.NewBattery(level)
	if err != nil {
		return bluetooth.Device{}, err
	}

	return s.apply(ctx, address, func(d bluetooth.Device) (bluetooth.Device, error) {
		return d.ReportBattery(battery)
	})
}

// Remove forgets the device entirely. Unlike Unpair, it drops out of the list.
func (s *Service) Remove(ctx context.Context, address string) error {
	addr, err := bluetooth.NewAddress(address)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, addr)
}

// List returns every known device.
func (s *Service) List(ctx context.Context) ([]bluetooth.Device, error) {
	return s.repo.List(ctx)
}

// Search filters by name, case-insensitively. With an empty query it is the
// same as List.
func (s *Service) Search(ctx context.Context, query string) ([]bluetooth.Device, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	needle := strings.ToLower(strings.TrimSpace(query))
	if needle == "" {
		return all, nil
	}

	out := make([]bluetooth.Device, 0, len(all))
	for _, d := range all {
		if strings.Contains(strings.ToLower(d.Name().String()), needle) {
			out = append(out, d)
		}
	}
	return out, nil
}
