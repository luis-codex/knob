// Package devices holds the Bluetooth-device use cases: one type per use case,
// each with a single Execute method, all sharing the injected ports through
// Deps.
//
// The rule that you cannot pair or connect while the radio is off lives here,
// not in the domain: it crosses two aggregates, and neither can know the other
// without coupling them.
//
// It is named devices (plural) rather than bluetooth so it does not clash with
// the domain package.
package devices

import (
	"context"
	"errors"

	"knob/internal/domain/bluetooth"
)

// Deps are the ports every use case in this package needs. The composition
// root fills it once and NewUseCases hands it to each use case.
type Deps struct {
	Repo     bluetooth.Repository
	Adapters bluetooth.AdapterRepository
	Scanner  bluetooth.Scanner
}

// transition is an aggregate operation that returns the resulting device. It
// matches the signature of Pair, Connect and friends, passed as method
// expressions.
type transition func(bluetooth.Device) (bluetooth.Device, error)

// requireAdapter fails if the radio is off.
func (d Deps) requireAdapter(ctx context.Context) error {
	adapter, err := d.Adapters.Get(ctx)
	if err != nil {
		return err
	}
	if !adapter.Enabled() {
		return bluetooth.ErrAdapterDisabled
	}
	return nil
}

// apply loads the device, applies the transition and saves the result. Every
// state operation follows this same path.
func (d Deps) apply(ctx context.Context, address string, change transition) (bluetooth.Device, error) {
	addr, err := bluetooth.NewAddress(address)
	if err != nil {
		return bluetooth.Device{}, err
	}

	current, err := d.Repo.FindByAddress(ctx, addr)
	if err != nil {
		return bluetooth.Device{}, err
	}

	next, err := change(current)
	if err != nil {
		return bluetooth.Device{}, err
	}

	if err := d.Repo.Save(ctx, next); err != nil {
		return bluetooth.Device{}, err
	}
	return next, nil
}

// knownAddresses is the set of addresses already stored.
func (d Deps) knownAddresses(ctx context.Context) (map[string]struct{}, error) {
	all, err := d.Repo.List(ctx)
	if err != nil {
		return nil, err
	}

	known := make(map[string]struct{}, len(all))
	for _, dev := range all {
		known[dev.Address().String()] = struct{}{}
	}
	return known, nil
}

// isKnown tells "does not exist" apart from "the repository failed".
func (d Deps) isKnown(ctx context.Context, address bluetooth.Address) (bool, error) {
	switch _, err := d.Repo.FindByAddress(ctx, address); {
	case err == nil:
		return true, nil
	case errors.Is(err, bluetooth.ErrNotFound):
		return false, nil
	default:
		return false, err
	}
}

// disconnectAll drops every active connection, saving each change.
func (d Deps) disconnectAll(ctx context.Context) error {
	all, err := d.Repo.List(ctx)
	if err != nil {
		return err
	}

	for _, dev := range all {
		if dev.State() != bluetooth.StateConnected {
			continue
		}

		next, err := dev.Disconnect()
		if err != nil {
			return err
		}
		if err := d.Repo.Save(ctx, next); err != nil {
			return err
		}
	}
	return nil
}
