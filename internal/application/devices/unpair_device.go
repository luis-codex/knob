package devices

import (
	"context"

	"knob/internal/domain/bluetooth"
)

// UnpairDevice forgets the pairing, leaving the device merely discovered. It
// works with the radio off, like any settings panel.
type UnpairDevice struct{ deps Deps }

func NewUnpairDevice(deps Deps) UnpairDevice { return UnpairDevice{deps: deps} }

type UnpairDeviceCommand struct{ Address string }

type UnpairDeviceResponse struct{ Device bluetooth.Device }

func (u UnpairDevice) Execute(ctx context.Context, cmd UnpairDeviceCommand) (UnpairDeviceResponse, error) {
	device, err := u.deps.apply(ctx, cmd.Address, bluetooth.Device.Unpair)
	if err != nil {
		return UnpairDeviceResponse{}, err
	}
	return UnpairDeviceResponse{Device: device}, nil
}
