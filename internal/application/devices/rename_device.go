package devices

import (
	"context"

	"knob/internal/domain/bluetooth"
)

// RenameDevice changes the device's visible name.
type RenameDevice struct{ deps Deps }

func NewRenameDevice(deps Deps) RenameDevice { return RenameDevice{deps: deps} }

type RenameDeviceCommand struct {
	Address string
	Name    string
}

type RenameDeviceResponse struct{ Device bluetooth.Device }

func (u RenameDevice) Execute(ctx context.Context, cmd RenameDeviceCommand) (RenameDeviceResponse, error) {
	newName, err := bluetooth.NewName(cmd.Name)
	if err != nil {
		return RenameDeviceResponse{}, err
	}

	device, err := u.deps.apply(ctx, cmd.Address, func(d bluetooth.Device) (bluetooth.Device, error) {
		return d.Rename(newName)
	})
	if err != nil {
		return RenameDeviceResponse{}, err
	}
	return RenameDeviceResponse{Device: device}, nil
}
