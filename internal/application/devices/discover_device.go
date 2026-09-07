package devices

import (
	"context"

	"knob/internal/domain/bluetooth"
)

// DiscoverDevice records a new device, not yet paired. Recording the same
// address twice is a conflict, not a create: a silent create would lose the
// state of the already-known device.
type DiscoverDevice struct{ deps Deps }

func NewDiscoverDevice(deps Deps) DiscoverDevice { return DiscoverDevice{deps: deps} }

type DiscoverDeviceCommand struct {
	Address string
	Name    string
	Kind    string
}

type DiscoverDeviceResponse struct{ Device bluetooth.Device }

func (u DiscoverDevice) Execute(ctx context.Context, cmd DiscoverDeviceCommand) (DiscoverDeviceResponse, error) {
	addr, err := bluetooth.NewAddress(cmd.Address)
	if err != nil {
		return DiscoverDeviceResponse{}, err
	}

	deviceName, err := bluetooth.NewName(cmd.Name)
	if err != nil {
		return DiscoverDeviceResponse{}, err
	}

	deviceKind, err := bluetooth.ParseKind(cmd.Kind)
	if err != nil {
		return DiscoverDeviceResponse{}, err
	}

	known, err := u.deps.isKnown(ctx, addr)
	if err != nil {
		return DiscoverDeviceResponse{}, err
	}
	if known {
		return DiscoverDeviceResponse{}, bluetooth.ErrAlreadyKnown
	}

	device, err := bluetooth.Discover(addr, deviceName, deviceKind)
	if err != nil {
		return DiscoverDeviceResponse{}, err
	}

	if err := u.deps.Repo.Save(ctx, device); err != nil {
		return DiscoverDeviceResponse{}, err
	}
	return DiscoverDeviceResponse{Device: device}, nil
}
