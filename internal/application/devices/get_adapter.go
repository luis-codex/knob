package devices

import (
	"context"

	"knob/internal/domain/bluetooth"
)

// GetAdapter returns the radio's state.
type GetAdapter struct{ deps Deps }

func NewGetAdapter(deps Deps) GetAdapter { return GetAdapter{deps: deps} }

type GetAdapterCommand struct{}

type GetAdapterResponse struct{ Adapter bluetooth.Adapter }

func (u GetAdapter) Execute(ctx context.Context, _ GetAdapterCommand) (GetAdapterResponse, error) {
	adapter, err := u.deps.Adapters.Get(ctx)
	if err != nil {
		return GetAdapterResponse{}, err
	}
	return GetAdapterResponse{Adapter: adapter}, nil
}
