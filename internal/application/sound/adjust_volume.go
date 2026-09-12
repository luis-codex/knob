package sound

import (
	"context"

	"knob/internal/domain/audio"
)

// AdjustVolume adds delta to the device's current level, clamping at the ends:
// it is what a volume-up or volume-down key does.
type AdjustVolume struct{ deps Deps }

func NewAdjustVolume(deps Deps) AdjustVolume { return AdjustVolume{deps: deps} }

type AdjustVolumeCommand struct {
	ID    string
	Delta int
}

type AdjustVolumeResponse struct{ Device audio.Device }

func (u AdjustVolume) Execute(ctx context.Context, cmd AdjustVolumeCommand) (AdjustVolumeResponse, error) {
	device, err := u.deps.apply(ctx, cmd.ID, func(d audio.Device) audio.Device {
		return d.AdjustVolume(cmd.Delta)
	})
	if err != nil {
		return AdjustVolumeResponse{}, err
	}
	return AdjustVolumeResponse{Device: device}, nil
}
