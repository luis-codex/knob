package bluez

import (
	"context"

	"settings-cli/internal/domain/bluetooth"
)

// AdapterRepository lee y cambia el estado de la radio real.
type AdapterRepository struct{}

var _ bluetooth.AdapterRepository = (*AdapterRepository)(nil)

func NewAdapterRepository() *AdapterRepository { return &AdapterRepository{} }

// Get lee "Powered: yes" de `bluetoothctl show`.
func (r *AdapterRepository) Get(ctx context.Context) (bluetooth.Adapter, error) {
	out, err := run(ctx, "show")
	if err != nil {
		return bluetooth.Adapter{}, err
	}
	return bluetooth.NewAdapter(yes(parseFields(out), "Powered")), nil
}

// Save enciende o apaga el adaptador. Solo actúa si hay algo que cambiar: un
// `power on` redundante devuelve error en BlueZ.
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
