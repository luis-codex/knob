package sound

import (
	"context"

	"knob/internal/domain/audio"
)

// ToggleStreamMuted flips a stream's mute.
type ToggleStreamMuted struct{ deps Deps }

func NewToggleStreamMuted(deps Deps) ToggleStreamMuted { return ToggleStreamMuted{deps: deps} }

type ToggleStreamMutedCommand struct{ Index int }

type ToggleStreamMutedResponse struct{ Stream audio.Stream }

func (u ToggleStreamMuted) Execute(ctx context.Context, cmd ToggleStreamMutedCommand) (ToggleStreamMutedResponse, error) {
	stream, err := u.deps.applyStream(ctx, cmd.Index, audio.Stream.ToggleMuted)
	if err != nil {
		return ToggleStreamMutedResponse{}, err
	}
	return ToggleStreamMutedResponse{Stream: stream}, nil
}
