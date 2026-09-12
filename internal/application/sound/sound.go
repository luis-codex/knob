// Package sound holds the audio use cases (outputs, microphones and streams):
// one type per use case, each with a single Execute method, all sharing the
// injected ports through Deps.
//
// It is named sound rather than audio so it does not clash with the domain
// package.
package sound

import (
	"context"

	"knob/internal/domain/audio"
)

// Deps are the ports every use case in this package needs. The composition
// root fills it once and NewUseCases hands it to each use case.
type Deps struct {
	Repo    audio.Repository
	Streams audio.StreamRepository
	Watcher audio.Watcher
}

// change is an aggregate operation. It returns no error: adjusting volume or
// mute is always possible; what can fail is the requested level, and that is
// validated beforehand.
type change func(audio.Device) audio.Device

// streamChange is the same for streams.
type streamChange func(audio.Stream) audio.Stream

// apply loads the device, applies the change and saves it. Every device
// operation follows this same path.
func (d Deps) apply(ctx context.Context, id string, c change) (audio.Device, error) {
	deviceID, err := audio.NewID(id)
	if err != nil {
		return audio.Device{}, err
	}

	current, err := d.Repo.FindByID(ctx, deviceID)
	if err != nil {
		return audio.Device{}, err
	}

	next := c(current)
	if err := d.Repo.Save(ctx, next); err != nil {
		return audio.Device{}, err
	}
	return next, nil
}

// applyStream loads the stream, applies the change and saves it.
func (d Deps) applyStream(ctx context.Context, index int, c streamChange) (audio.Stream, error) {
	id, err := audio.NewStreamID(index)
	if err != nil {
		return audio.Stream{}, err
	}

	current, err := d.Streams.FindByID(ctx, id)
	if err != nil {
		return audio.Stream{}, err
	}

	next := c(current)
	if err := d.Streams.Save(ctx, next); err != nil {
		return audio.Stream{}, err
	}
	return next, nil
}
