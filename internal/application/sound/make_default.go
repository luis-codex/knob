package sound

import (
	"context"

	"knob/internal/domain/audio"
)

// MakeDefault marks the device as the default for its direction.
//
// It does not go through apply: changing the default affects the other
// devices, so the sound server resolves it and then the device is re-read.
type MakeDefault struct{ deps Deps }

func NewMakeDefault(deps Deps) MakeDefault { return MakeDefault{deps: deps} }

type MakeDefaultCommand struct{ ID string }

type MakeDefaultResponse struct{ Device audio.Device }

func (u MakeDefault) Execute(ctx context.Context, cmd MakeDefaultCommand) (MakeDefaultResponse, error) {
	deviceID, err := audio.NewID(cmd.ID)
	if err != nil {
		return MakeDefaultResponse{}, err
	}

	if _, err := u.deps.Repo.FindByID(ctx, deviceID); err != nil {
		return MakeDefaultResponse{}, err
	}

	if err := u.deps.Repo.SetDefault(ctx, deviceID); err != nil {
		return MakeDefaultResponse{}, err
	}

	device, err := u.deps.Repo.FindByID(ctx, deviceID)
	if err != nil {
		return MakeDefaultResponse{}, err
	}
	return MakeDefaultResponse{Device: device}, nil
}
