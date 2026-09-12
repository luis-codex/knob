package sound

import (
	"context"

	"knob/internal/domain/audio"
)

// ListOutputs returns the output devices: speakers and headphones.
type ListOutputs struct{ deps Deps }

func NewListOutputs(deps Deps) ListOutputs { return ListOutputs{deps: deps} }

type ListOutputsCommand struct{}

type ListOutputsResponse struct{ Devices []audio.Device }

func (u ListOutputs) Execute(ctx context.Context, _ ListOutputsCommand) (ListOutputsResponse, error) {
	devices, err := u.deps.Repo.List(ctx, audio.Output)
	if err != nil {
		return ListOutputsResponse{}, err
	}
	return ListOutputsResponse{Devices: devices}, nil
}
