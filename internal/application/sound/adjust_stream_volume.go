package sound

import (
	"context"

	"knob/internal/domain/audio"
)

// AdjustStreamVolume adds delta to the stream's level, clamping at the ends.
type AdjustStreamVolume struct{ deps Deps }

func NewAdjustStreamVolume(deps Deps) AdjustStreamVolume { return AdjustStreamVolume{deps: deps} }

type AdjustStreamVolumeCommand struct {
	Index int
	Delta int
}

type AdjustStreamVolumeResponse struct{ Stream audio.Stream }

func (u AdjustStreamVolume) Execute(ctx context.Context, cmd AdjustStreamVolumeCommand) (AdjustStreamVolumeResponse, error) {
	stream, err := u.deps.applyStream(ctx, cmd.Index, func(st audio.Stream) audio.Stream {
		return st.AdjustVolume(cmd.Delta)
	})
	if err != nil {
		return AdjustStreamVolumeResponse{}, err
	}
	return AdjustStreamVolumeResponse{Stream: stream}, nil
}
