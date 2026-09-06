package bluez

import (
	"context"

	"knob/internal/domain/bluetooth"
)

// AdapterRepository reads and changes the state of the real radio.
type AdapterRepository struct{}

var _ bluetooth.AdapterRepository = (*AdapterRepository)(nil)

func NewAdapterRepository() *AdapterRepository { return &AdapterRepository{} }

// Get reads "Powered: yes" from `bluetoothctl show`.
func (r *AdapterRepository) Get(ctx context.Context) (bluetooth.Adapter, error) {
	out, err := run(ctx, "show")
	if err != nil {
		return bluetooth.Adapter{}, err
	}
	return bluetooth.NewAdapter(yes(parseFields(out), "Powered")), nil
}

// Save turns the adapter on or off. It only acts if there is something to
// change: a redundant `power on` returns an error in BlueZ.
func (r *AdapterRepository) Save(ctx context.Context, a bluetooth.Adapter) error {
	current, err := r.Get(ctx)
	if err != nil {
		return err
	}
	if current.Enabled() == a.Enabled() {
		return nil
	}

	state := "off"
	if a.Enabled() {
		state = "on"
	}

	_, err = run(ctx, "power", state)
	return err
}
