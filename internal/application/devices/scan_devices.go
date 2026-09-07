package devices

import (
	"context"

	"knob/internal/domain/bluetooth"
)

// ScanDevices looks for nearby devices and records the ones that were not
// known, returning only the new ones so a known device's state and name are
// never overwritten.
type ScanDevices struct{ deps Deps }

func NewScanDevices(deps Deps) ScanDevices { return ScanDevices{deps: deps} }

type ScanDevicesCommand struct{}

type ScanDevicesResponse struct{ Devices []bluetooth.Device }

func (u ScanDevices) Execute(ctx context.Context, _ ScanDevicesCommand) (ScanDevicesResponse, error) {
	if err := u.deps.requireAdapter(ctx); err != nil {
		return ScanDevicesResponse{}, err
	}

	// What is already known is checked BEFORE scanning. With a real backend
	// the scan itself makes the repository come to know what it finds, so
	// asking it afterwards would always say "already known" and there would
	// never be anything new.
	known, err := u.deps.knownAddresses(ctx)
	if err != nil {
		return ScanDevicesResponse{}, err
	}

	found, err := u.deps.Scanner.Scan(ctx)
	if err != nil {
		return ScanDevicesResponse{}, err
	}

	var added []bluetooth.Device
	for _, device := range found {
		if _, ok := known[device.Address().String()]; ok {
			continue
		}

		if err := u.deps.Repo.Save(ctx, device); err != nil {
			return ScanDevicesResponse{}, err
		}
		added = append(added, device)
	}
	return ScanDevicesResponse{Devices: added}, nil
}
