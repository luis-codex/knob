package memory

import (
	"context"
	"sync"

	"settings-cli/internal/domain/bluetooth"
)

// DeviceRepository implementa bluetooth.Repository.
//
// Guarda en slice y no en mapa porque el puerto exige orden de descubrimiento.
type DeviceRepository struct {
	mu      sync.RWMutex
	devices []bluetooth.Device
}

var _ bluetooth.Repository = (*DeviceRepository)(nil)

func NewDeviceRepository() *DeviceRepository {
	return &DeviceRepository{}
}

// Save inserta o actualiza según exista ya la dirección.
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

// List devuelve una copia: sin ella, quien la reciba podría reordenar o
// sobrescribir el almacén.
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

// indexOf devuelve la posición de address, o -1. Se llama con el candado
// tomado.
func (r *DeviceRepository) indexOf(address bluetooth.Address) int {
	for i, d := range r.devices {
		if d.Address().Equals(address) {
			return i
		}
	}
	return -1
}
