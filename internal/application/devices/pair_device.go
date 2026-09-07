package devices

import (
	"context"

	"knob/internal/domain/bluetooth"
)

// PairDevice pairs a discovered device. The radio must be on.
type PairDevice struct{ deps Deps }

func NewPairDevice(deps Deps) PairDevice { return PairDevice{deps: deps} }

type PairDeviceCommand struct{ Address string }

type PairDeviceResponse struct{ Device bluetooth.Device }

func (u PairDevice) Execute(ctx context.Context, cmd PairDeviceCommand) (PairDeviceResponse, error) {
	if err := u.deps.requireAdapter(ctx); err != nil {
		return PairDeviceResponse{}, err
	}

	device, err := u.deps.apply(ctx, cmd.Address, bluetooth.Device.Pair)
	if err != nil {
		return PairDeviceResponse{}, err
	}
	return PairDeviceResponse{Device: device}, nil
}
