package devices

import (
	"context"
	"strings"

	"knob/internal/domain/bluetooth"
)

// SearchDevices filters the known devices by name, case-insensitively. An empty
// query returns them all.
type SearchDevices struct{ deps Deps }

func NewSearchDevices(deps Deps) SearchDevices { return SearchDevices{deps: deps} }

type SearchDevicesCommand struct{ Query string }

type SearchDevicesResponse struct{ Devices []bluetooth.Device }

func (u SearchDevices) Execute(ctx context.Context, cmd SearchDevicesCommand) (SearchDevicesResponse, error) {
	all, err := u.deps.Repo.List(ctx)
	if err != nil {
		return SearchDevicesResponse{}, err
	}

	needle := strings.ToLower(strings.TrimSpace(cmd.Query))
	if needle == "" {
		return SearchDevicesResponse{Devices: all}, nil
	}

	out := make([]bluetooth.Device, 0, len(all))
	for _, d := range all {
		if strings.Contains(strings.ToLower(d.Name().String()), needle) {
			out = append(out, d)
		}
	}
	return SearchDevicesResponse{Devices: out}, nil
}
