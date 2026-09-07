package devices

import (
	"context"

	"knob/internal/domain/bluetooth"
)

// ListDevices returns every known device, in discovery order.
type ListDevices struct{ deps Deps }

func NewListDevices(deps Deps) ListDevices { return ListDevices{deps: deps} }

type ListDevicesCommand struct{}

type ListDevicesResponse struct{ Devices []bluetooth.Device }

func (u ListDevices) Execute(ctx context.Context, _ ListDevicesCommand) (ListDevicesResponse, error) {
	all, err := u.deps.Repo.List(ctx)
	if err != nil {
		return ListDevicesResponse{}, err
	}
	return ListDevicesResponse{Devices: all}, nil
}
