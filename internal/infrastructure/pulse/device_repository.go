package pulse

import (
	"context"
	"encoding/json"
	"strconv"

	"settings-cli/internal/domain/audio"
	"settings-cli/internal/domain/errs"
)

// Repository expone los dispositivos de sonido del servidor.
type Repository struct{}

var _ audio.Repository = (*Repository)(nil)

func NewRepository() *Repository { return &Repository{} }

func (r *Repository) List(ctx context.Context, direction audio.Direction) ([]audio.Device, error) {
	defaultName, err := r.defaultName(ctx, direction)
	if err != nil {
		return nil, err
	}

	raw, err := run(ctx, "list", kindOf(direction)+"s")
	if err != nil {
		return nil, err
	}

	decoded, err := decodeDevices(raw)
	if err != nil {
		return nil, err
	}

	devices := make([]audio.Device, 0, len(decoded))
	for _, d := range decoded {
		if device, ok := toDevice(d, direction, defaultName); ok {
			devices = append(devices, device)
		}
	}
	return devices, nil
}

// defaultName lee el predeterminado de la dirección desde `pactl info`.
func (r *Repository) defaultName(ctx context.Context, direction audio.Direction) (string, error) {
	raw, err := run(ctx, "info")
	if err != nil {
		return "", err
	}

	var info jsonInfo
	if err := json.Unmarshal(raw, &info); err != nil {
		return "", errs.Wrap(errs.KindConflict, "respuesta ilegible del servidor de sonido", err)
	}

	if direction == audio.Input {
		return info.DefaultSource, nil
	}
	return info.DefaultSink, nil
}

// FindByID busca en ambas direcciones: el identificador no dice de cuál es.
func (r *Repository) FindByID(ctx context.Context, id audio.ID) (audio.Device, error) {
	for _, direction := range []audio.Direction{audio.Output, audio.Input} {
		devices, err := r.List(ctx, direction)
		if err != nil {
			return audio.Device{}, err
		}

		for _, device := range devices {
			if device.ID().Equals(id) {
				return device, nil
			}
		}
	}
	return audio.Device{}, audio.ErrNotFound
}

// Save aplica volumen y silencio. Ambos comandos son idempotentes, así que no
// hace falta comparar antes con el estado actual.
func (r *Repository) Save(ctx context.Context, d audio.Device) error {
	if d.IsZero() {
		return audio.ErrInvalidID
	}

	kind := kindOf(d.Direction())
	id := d.ID().String()

	if _, err := run(ctx, "set-"+kind+"-volume", id, strconv.Itoa(d.Volume().Level())+"%"); err != nil {
		return err
	}

	muted := "0"
	if d.Muted() {
		muted = "1"
	}
	_, err := run(ctx, "set-"+kind+"-mute", id, muted)
	return err
}

// SetDefault necesita saber la dirección, que no está en el identificador: se
// resuelve leyendo el dispositivo primero.
func (r *Repository) SetDefault(ctx context.Context, id audio.ID) error {
	device, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}

	_, err = run(ctx, "set-default-"+kindOf(device.Direction()), id.String())
	return err
}
