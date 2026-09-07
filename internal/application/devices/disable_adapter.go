package devices

import (
	"context"

	"knob/internal/domain/bluetooth"
)

// DisableAdapter turns the radio off and drops the active connections.
//
// Devices are disconnected before the adapter is saved: if something failed
// halfway, the saved state would still say "on", which is more faithful than
// the opposite.
type DisableAdapter struct{ deps Deps }

func NewDisableAdapter(deps Deps) DisableAdapter { return DisableAdapter{deps: deps} }

type DisableAdapterCommand struct{}

type DisableAdapterResponse struct{ Adapter bluetooth.Adapter }

func (u DisableAdapter) Execute(ctx context.Context, _ DisableAdapterCommand) (DisableAdapterResponse, error) {
	adapter, err := u.deps.Adapters.Get(ctx)
	if err != nil {
		return DisableAdapterResponse{}, err
	}

	if err := u.deps.disconnectAll(ctx); err != nil {
		return DisableAdapterResponse{}, err
	}

	next := adapter.Disable()
	if err := u.deps.Adapters.Save(ctx, next); err != nil {
		return DisableAdapterResponse{}, err
	}
	return DisableAdapterResponse{Adapter: next}, nil
}
