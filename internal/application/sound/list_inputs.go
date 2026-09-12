package sound

import (
	"context"

	"knob/internal/domain/audio"
)

// ListInputs returns the input devices: microphones.
type ListInputs struct{ deps Deps }

func NewListInputs(deps Deps) ListInputs { return ListInputs{deps: deps} }

type ListInputsCommand struct{}

type ListInputsResponse struct{ Devices []audio.Device }

func (u ListInputs) Execute(ctx context.Context, _ ListInputsCommand) (ListInputsResponse, error) {
	devices, err := u.deps.Repo.List(ctx, audio.Input)
	if err != nil {
		return ListInputsResponse{}, err
	}
	return ListInputsResponse{Devices: devices}, nil
}
