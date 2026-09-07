package devices

import (
	"context"

	"knob/internal/domain/bluetooth"
)

// DisconnectDevice drops the connection, leaving the device paired. It works
// with the radio off.
type DisconnectDevice struct{ deps Deps }

func NewDisconnectDevice(deps Deps) DisconnectDevice { return DisconnectDevice{deps: deps} }

type DisconnectDeviceCommand struct{ Address string }

type DisconnectDeviceResponse struct{ Device bluetooth.Device }

func (u DisconnectDevice) Execute(ctx context.Context, cmd DisconnectDeviceCommand) (DisconnectDeviceResponse, error) {
	device, err := u.deps.apply(ctx, cmd.Address, bluetooth.Device.Disconnect)
	if err != nil {
		return DisconnectDeviceResponse{}, err
	}
	return DisconnectDeviceResponse{Device: device}, nil
}
