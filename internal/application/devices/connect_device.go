package devices

import (
	"context"

	"knob/internal/domain/bluetooth"
)

// ConnectDevice connects a paired device. The radio must be on.
type ConnectDevice struct{ deps Deps }

func NewConnectDevice(deps Deps) ConnectDevice { return ConnectDevice{deps: deps} }

type ConnectDeviceCommand struct{ Address string }

type ConnectDeviceResponse struct{ Device bluetooth.Device }

func (u ConnectDevice) Execute(ctx context.Context, cmd ConnectDeviceCommand) (ConnectDeviceResponse, error) {
	if err := u.deps.requireAdapter(ctx); err != nil {
		return ConnectDeviceResponse{}, err
	}

	device, err := u.deps.apply(ctx, cmd.Address, bluetooth.Device.Connect)
	if err != nil {
		return ConnectDeviceResponse{}, err
	}
	return ConnectDeviceResponse{Device: device}, nil
}
