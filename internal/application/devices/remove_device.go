package devices

import (
	"context"

	"knob/internal/domain/bluetooth"
)

// RemoveDevice forgets the device entirely. Unlike UnpairDevice, it drops out
// of the list.
type RemoveDevice struct{ deps Deps }

func NewRemoveDevice(deps Deps) RemoveDevice { return RemoveDevice{deps: deps} }

type RemoveDeviceCommand struct{ Address string }

type RemoveDeviceResponse struct{}

func (u RemoveDevice) Execute(ctx context.Context, cmd RemoveDeviceCommand) (RemoveDeviceResponse, error) {
	addr, err := bluetooth.NewAddress(cmd.Address)
	if err != nil {
		return RemoveDeviceResponse{}, err
	}
	if err := u.deps.Repo.Delete(ctx, addr); err != nil {
		return RemoveDeviceResponse{}, err
	}
	return RemoveDeviceResponse{}, nil
}
