// Package sound contiene los casos de uso de audio: salidas y micrófonos.
//
// Es la frontera del dominio: recibe primitivas de la interfaz, las convierte
// en objetos de valor y orquesta el repositorio.
//
// Se llama sound y no audio para no chocar con el paquete de dominio.
package sound

import (
	"context"

	"settings-cli/internal/domain/audio"
)

// change es una operación del agregado. No devuelve error: ajustar volumen o
// silencio siempre es posible, lo que puede fallar es el nivel pedido, y eso
// se valida antes.
type change func(audio.Device) audio.Device

// streamChange es lo mismo para los flujos.
type streamChange func(audio.Stream) audio.Stream

// Service agrupa los casos de uso sobre dispositivos de sonido.
type Service struct {
	repo    audio.Repository
	streams audio.StreamRepository
	watcher audio.Watcher
}

// NewService recibe los puertos, no implementaciones concretas.
func NewService(repo audio.Repository, streams audio.StreamRepository, watcher audio.Watcher) *Service {
	return &Service{repo: repo, streams: streams, watcher: watcher}
}

// Streams son los flujos de audio de las aplicaciones.
func (s *Service) Streams(ctx context.Context) ([]audio.Stream, error) {
	return s.streams.List(ctx)
}

// applyStream carga el flujo, le aplica el cambio y lo guarda.
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

// AdjustStreamVolume suma delta al nivel del flujo, acotando en los extremos.
func (s *Service) AdjustStreamVolume(ctx context.Context, index, delta int) (audio.Stream, error) {
	return s.applyStream(ctx, index, func(st audio.Stream) audio.Stream {
		return st.AdjustVolume(delta)
	})
}

func (s *Service) ToggleStreamMuted(ctx context.Context, index int) (audio.Stream, error) {
	return s.applyStream(ctx, index, audio.Stream.ToggleMuted)
}

// Changes avisa de los cambios hechos fuera de la aplicación.
func (s *Service) Changes(ctx context.Context) (<-chan struct{}, error) {
	return s.watcher.Changes(ctx)
}

// Outputs son las salidas: altavoces y auriculares.
func (s *Service) Outputs(ctx context.Context) ([]audio.Device, error) {
	return s.repo.List(ctx, audio.Output)
}

// Inputs son los micrófonos.
func (s *Service) Inputs(ctx context.Context) ([]audio.Device, error) {
	return s.repo.List(ctx, audio.Input)
}

// apply carga el dispositivo, le aplica el cambio y lo guarda. Todas las
// operaciones siguen este mismo camino.
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

// AdjustVolume suma delta al nivel actual. A diferencia de SetVolume acota en
// los extremos: es lo que hace una tecla de subir o bajar volumen.
func (s *Service) AdjustVolume(ctx context.Context, id string, delta int) (audio.Device, error) {
	return s.apply(ctx, id, func(d audio.Device) audio.Device {
		return d.AdjustVolume(delta)
	})
}

func (s *Service) ToggleMuted(ctx context.Context, id string) (audio.Device, error) {
	return s.apply(ctx, id, audio.Device.ToggleMuted)
}

// MakeDefault marca el dispositivo como predeterminado para su dirección.
//
// No pasa por apply: cambiar el predeterminado afecta a los demás
// dispositivos, así que lo resuelve el servidor de sonido y luego se relee.
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
