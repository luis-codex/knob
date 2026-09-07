package devices

import (
	"context"

	"knob/internal/domain/bluetooth"
)

// EnableAdapter turns the radio on. It reconnects nothing: reconnecting is the
// user's decision, not an effect of turning on.
type EnableAdapter struct{ deps Deps }

func NewEnableAdapter(deps Deps) EnableAdapter { return EnableAdapter{deps: deps} }

type EnableAdapterCommand struct{}

type EnableAdapterResponse struct{ Adapter bluetooth.Adapter }

func (u EnableAdapter) Execute(ctx context.Context, _ EnableAdapterCommand) (EnableAdapterResponse, error) {
	adapter, err := u.deps.Adapters.Get(ctx)
	if err != nil {
		return EnableAdapterResponse{}, err
	}

	next := adapter.Enable()
	if err := u.deps.Adapters.Save(ctx, next); err != nil {
		return EnableAdapterResponse{}, err
	}
	return EnableAdapterResponse{Adapter: next}, nil
}
