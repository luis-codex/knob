package pulse

import (
	"context"
	"encoding/json"
	"strconv"

	"knob/internal/domain/audio"
	"knob/internal/domain/errs"
)

// Repository exposes the server's sound devices.
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

// defaultName reads the direction's default from `pactl info`.
func (r *Repository) defaultName(ctx context.Context, direction audio.Direction) (string, error) {
	raw, err := run(ctx, "info")
	if err != nil {
		return "", err
	}

	var info jsonInfo
	if err := json.Unmarshal(raw, &info); err != nil {
		return "", errs.Wrap(errs.KindConflict, "unreadable response from the sound server", err)
	}

	if direction == audio.Input {
		return info.DefaultSource, nil
	}
	return info.DefaultSink, nil
}

// FindByID searches both directions: the identifier does not say which one it
// belongs to.
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

// Save applies volume and mute. Both commands are idempotent, so there is no
// need to compare against the current state first.
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

// SetDefault needs to know the direction, which is not in the identifier: it
// is resolved by reading the device first.
func (r *Repository) SetDefault(ctx context.Context, id audio.ID) error {
	device, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}

	_, err = run(ctx, "set-default-"+kindOf(device.Direction()), id.String())
	return err
}
