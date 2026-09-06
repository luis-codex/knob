package memory

import (
	"context"
	"sync"

	"knob/internal/domain/bluetooth"
)

// DeviceRepository implements bluetooth.Repository.
//
// It stores a slice rather than a map because the port requires discovery
// order.
type DeviceRepository struct {
	mu      sync.RWMutex
	devices []bluetooth.Device
}

var _ bluetooth.Repository = (*DeviceRepository)(nil)

func NewDeviceRepository() *DeviceRepository {
	return &DeviceRepository{}
}

// Save inserts or updates depending on whether the address already exists.
func (r *DeviceRepository) Save(ctx context.Context, d bluetooth.Device) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if d.IsZero() {
		return bluetooth.ErrInvalidAddress
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if i := r.indexOf(d.Address()); i >= 0 {
		r.devices[i] = d
		return nil
	}

	r.devices = append(r.devices, d)
	return nil
}

func (r *DeviceRepository) Delete(ctx context.Context, address bluetooth.Address) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	i := r.indexOf(address)
	if i < 0 {
		return bluetooth.ErrNotFound
	}

	r.devices = append(r.devices[:i], r.devices[i+1:]...)
	return nil
}

func (r *DeviceRepository) FindByAddress(ctx context.Context, address bluetooth.Address) (bluetooth.Device, error) {
	if err := ctx.Err(); err != nil {
		return bluetooth.Device{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	i := r.indexOf(address)
	if i < 0 {
		return bluetooth.Device{}, bluetooth.ErrNotFound
	}
	return r.devices[i], nil
}

// List returns a copy: without it, the receiver could reorder or overwrite the
// store.
func (r *DeviceRepository) List(ctx context.Context) ([]bluetooth.Device, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]bluetooth.Device, len(r.devices))
	copy(out, r.devices)
	return out, nil
}

// indexOf returns the position of address, or -1. It is called with the lock
// held.
func (r *DeviceRepository) indexOf(address bluetooth.Address) int {
	for i, d := range r.devices {
		if d.Address().Equals(address) {
			return i
		}
	}
	return -1
}
