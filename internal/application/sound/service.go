// Package sound holds the audio use cases: outputs and microphones.
//
// It is the domain's boundary: it takes primitives from the interface, turns
// them into value objects and orchestrates the repository.
//
// It is named sound rather than audio so it does not clash with the domain
// package.
package sound

import (
	"context"

	"knob/internal/domain/audio"
)

// change is an aggregate operation. It returns no error: adjusting volume or
// mute is always possible; what can fail is the requested level, and that is
// validated beforehand.
type change func(audio.Device) audio.Device

// streamChange is the same for streams.
type streamChange func(audio.Stream) audio.Stream

// Service groups the use cases over sound devices.
type Service struct {
	repo    audio.Repository
	streams audio.StreamRepository
	watcher audio.Watcher
}

// NewService takes the ports, not concrete implementations.
func NewService(repo audio.Repository, streams audio.StreamRepository, watcher audio.Watcher) *Service {
	return &Service{repo: repo, streams: streams, watcher: watcher}
}

// Streams are the applications' audio streams.
func (s *Service) Streams(ctx context.Context) ([]audio.Stream, error) {
	return s.streams.List(ctx)
}

// applyStream loads the stream, applies the change and saves it.
func (s *Service) applyStream(ctx context.Context, index int, c streamChange) (audio.Stream, error) {
	id, err := audio.NewStreamID(index)
	if err != nil {
		return audio.Stream{}, err
	}

	current, err := s.streams.FindByID(ctx, id)
	if err != nil {
		return audio.Stream{}, err
	}

	next := c(current)
	if err := s.streams.Save(ctx, next); err != nil {
		return audio.Stream{}, err
	}
	return next, nil
}

// AdjustStreamVolume adds delta to the stream's level, clamping at the ends.
func (s *Service) AdjustStreamVolume(ctx context.Context, index, delta int) (audio.Stream, error) {
	return s.applyStream(ctx, index, func(st audio.Stream) audio.Stream {
		return st.AdjustVolume(delta)
	})
}

func (s *Service) ToggleStreamMuted(ctx context.Context, index int) (audio.Stream, error) {
	return s.applyStream(ctx, index, audio.Stream.ToggleMuted)
}

// Changes reports changes made outside the application.
func (s *Service) Changes(ctx context.Context) (<-chan struct{}, error) {
	return s.watcher.Changes(ctx)
}

// Outputs are the outputs: speakers and headphones.
func (s *Service) Outputs(ctx context.Context) ([]audio.Device, error) {
	return s.repo.List(ctx, audio.Output)
}

// Inputs are the microphones.
func (s *Service) Inputs(ctx context.Context) ([]audio.Device, error) {
	return s.repo.List(ctx, audio.Input)
}

// apply loads the device, applies the change and saves it. Every operation
// follows this same path.
func (s *Service) apply(ctx context.Context, id string, c change) (audio.Device, error) {
	deviceID, err := audio.NewID(id)
	if err != nil {
		return audio.Device{}, err
	}

	current, err := s.repo.FindByID(ctx, deviceID)
	if err != nil {
		return audio.Device{}, err
	}

	next := c(current)
	if err := s.repo.Save(ctx, next); err != nil {
		return audio.Device{}, err
	}
	return next, nil
}

// AdjustVolume adds delta to the current level. Unlike SetVolume it clamps at
// the ends: it is what a volume-up or volume-down key does.
func (s *Service) AdjustVolume(ctx context.Context, id string, delta int) (audio.Device, error) {
	return s.apply(ctx, id, func(d audio.Device) audio.Device {
		return d.AdjustVolume(delta)
	})
}

func (s *Service) ToggleMuted(ctx context.Context, id string) (audio.Device, error) {
	return s.apply(ctx, id, audio.Device.ToggleMuted)
}

// MakeDefault marks the device as the default for its direction.
//
// It does not go through apply: changing the default affects the other
// devices, so the sound server resolves it and then it is re-read.
func (s *Service) MakeDefault(ctx context.Context, id string) (audio.Device, error) {
	deviceID, err := audio.NewID(id)
	if err != nil {
		return audio.Device{}, err
	}

	if _, err := s.repo.FindByID(ctx, deviceID); err != nil {
		return audio.Device{}, err
	}

	if err := s.repo.SetDefault(ctx, deviceID); err != nil {
		return audio.Device{}, err
	}
	return s.repo.FindByID(ctx, deviceID)
}
