package sound

import (
	"context"

	"knob/internal/domain/audio"
)

// ListStreams returns the applications' active audio streams.
type ListStreams struct{ deps Deps }

func NewListStreams(deps Deps) ListStreams { return ListStreams{deps: deps} }

type ListStreamsCommand struct{}

type ListStreamsResponse struct{ Streams []audio.Stream }

func (u ListStreams) Execute(ctx context.Context, _ ListStreamsCommand) (ListStreamsResponse, error) {
	streams, err := u.deps.Streams.List(ctx)
	if err != nil {
		return ListStreamsResponse{}, err
	}
	return ListStreamsResponse{Streams: streams}, nil
}
