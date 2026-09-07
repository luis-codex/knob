package sound

import (
	"context"

	"knob/internal/domain/audio"
)

// ToggleMuted flips the device's mute.
type ToggleMuted struct{ deps Deps }

func NewToggleMuted(deps Deps) ToggleMuted { return ToggleMuted{deps: deps} }

type ToggleMutedCommand struct{ ID string }

type ToggleMutedResponse struct{ Device audio.Device }

func (u ToggleMuted) Execute(ctx context.Context, cmd ToggleMutedCommand) (ToggleMutedResponse, error) {
	device, err := u.deps.apply(ctx, cmd.ID, audio.Device.ToggleMuted)
	if err != nil {
		return ToggleMutedResponse{}, err
	}
	return ToggleMutedResponse{Device: device}, nil
}
