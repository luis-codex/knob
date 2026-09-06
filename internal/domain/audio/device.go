// Package audio holds the sound-device aggregate -- outputs and microphones --
// its value objects and the persistence port. It depends on no other layer.
package audio

import (
	"strings"
	"unicode/utf8"
)

const (
	// MaxVolume allows going past 100%: PipeWire and PulseAudio amplify above
	// the nominal level, and refusing that would make it impossible to reflect
	// the real state of the system.
	MaxVolume = 150
	// NominalVolume is 100%, the maximum without amplification.
	NominalVolume = 100
	// MaxNameLength is the name limit, in characters.
	MaxNameLength = 80
)

// Direction tells which way the sound goes.
type Direction int

const (
	// Output is speakers and headphones.
	Output Direction = iota
	// Input is microphones.
	Input
)

func (d Direction) String() string {
	if d == Input {
		return "input"
	}
	return "output"
}

// ID identifies a device. Its format is opaque: it comes from the sound
// server and must not be interpreted.
type ID struct {
	value string
}

func NewID(s string) (ID, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return ID{}, ErrInvalidID
	}
	return ID{value: s}, nil
}

func (i ID) String() string { return i.value }

func (i ID) IsZero() bool { return i.value == "" }

func (i ID) Equals(other ID) bool { return i.value == other.value }

// Name is the human-readable description of the device.
type Name struct {
	value string
}

func NewName(s string) (Name, error) {
	s = strings.TrimSpace(s)

	switch {
	case s == "":
		return Name{}, ErrEmptyName
	case utf8.RuneCountInString(s) > MaxNameLength:
		return Name{}, ErrNameTooLong
	}
	return Name{value: s}, nil
}

func (n Name) String() string { return n.value }

func (n Name) IsZero() bool { return n.value == "" }

// Volume is a level from 0 to MaxVolume. Its zero value is 0%, which is a
// valid level: muting by lowering the volume is not the same as muting with
// mute, but both are legitimate states.
type Volume struct {
	level int
}

func NewVolume(level int) (Volume, error) {
	if level < 0 || level > MaxVolume {
		return Volume{}, ErrInvalidVolume
	}
	return Volume{level: level}, nil
}

// ClampVolume clamps instead of failing. It is for rebuilding system state:
// if the sound server reports 200%, refusing to read it would be worse than
// reflecting it as high as the domain allows.
func ClampVolume(level int) Volume {
	return Volume{level: min(max(level, 0), MaxVolume)}
}

func (v Volume) Level() int { return v.level }

// Amplified reports whether it goes past the nominal level.
func (v Volume) Amplified() bool { return v.level > NominalVolume }

// Device is the aggregate: a sound output or input.
type Device struct {
	id        ID
	name      Name
	direction Direction
	volume    Volume
	muted     bool
	isDefault bool
}

// Restore rebuilds a device exactly as the system reports it. There is no
// creation constructor: sound devices are not created from the app, they are
// discovered.
func Restore(id ID, name Name, direction Direction, volume Volume, muted, isDefault bool) (Device, error) {
	switch {
	case id.IsZero():
		return Device{}, ErrInvalidID
	case name.IsZero():
		return Device{}, ErrEmptyName
	}

	return Device{
		id:        id,
		name:      name,
		direction: direction,
		volume:    volume,
		muted:     muted,
		isDefault: isDefault,
	}, nil
}

func (d Device) ID() ID { return d.id }

func (d Device) Name() Name { return d.name }

func (d Device) Direction() Direction { return d.direction }

func (d Device) Volume() Volume { return d.volume }

func (d Device) Muted() bool { return d.muted }

// IsDefault reports whether this is the device the system uses by default for
// its direction.
func (d Device) IsDefault() bool { return d.isDefault }

func (d Device) IsZero() bool { return d.id.IsZero() }

// SetVolume adjusts the level. It does not clear mute: they are two
// independent controls, just like on any mixer.
func (d Device) SetVolume(v Volume) Device {
	d.volume = v
	return d
}

// AdjustVolume adds delta to the current level, clamping at the ends. It is
// what a volume-up or volume-down key does: hitting the limit is not an error.
func (d Device) AdjustVolume(delta int) Device {
	return d.SetVolume(ClampVolume(d.volume.Level() + delta))
}

// SetMuted is idempotent, like any switch.
func (d Device) SetMuted(muted bool) Device {
	d.muted = muted
	return d
}

func (d Device) ToggleMuted() Device {
	return d.SetMuted(!d.muted)
}
