// Package audio contiene el agregado de dispositivos de sonido —salidas y
// micrófonos—, sus objetos de valor y el puerto de persistencia. No depende de
// ninguna otra capa.
package audio

import (
	"strings"
	"unicode/utf8"
)

const (
	// MaxVolume permite pasar del 100 %: PipeWire y PulseAudio amplifican por
	// encima del nivel nominal, y negarlo haría imposible reflejar el estado
	// real del sistema.
	MaxVolume = 150
	// NominalVolume es el 100 %, el máximo sin amplificar.
	NominalVolume = 100
	// MaxNameLength es el límite del nombre, en caracteres.
	MaxNameLength = 80
)

// Direction distingue por dónde va el sonido.
type Direction int

const (
	// Output son altavoces y auriculares.
	Output Direction = iota
	// Input son micrófonos.
	Input
)

func (d Direction) String() string {
	if d == Input {
		return "input"
	}
	return "output"
}

// ID identifica un dispositivo. Su formato es opaco: viene del servidor de
// sonido y no debe interpretarse.
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

// Name es la descripción legible del dispositivo.
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

// Volume es un nivel de 0 a MaxVolume. Su valor cero es 0 %, que es un nivel
// válido: silenciar bajando el volumen no es lo mismo que silenciar con mute,
// pero ambos son estados legítimos.
type Volume struct {
	level int
}

func NewVolume(level int) (Volume, error) {
	if level < 0 || level > MaxVolume {
		return Volume{}, ErrInvalidVolume
	}
	return Volume{level: level}, nil
}

// ClampVolume acota en vez de fallar. Es para reconstruir estado del sistema:
// si el servidor de sonido informa de un 200 %, negarse a leerlo sería peor
// que reflejarlo tan alto como el dominio permite.
func ClampVolume(level int) Volume {
	return Volume{level: min(max(level, 0), MaxVolume)}
}

func (v Volume) Level() int { return v.level }

// Amplified indica si pasa del nivel nominal.
func (v Volume) Amplified() bool { return v.level > NominalVolume }

// Device es el agregado: una salida o una entrada de sonido.
type Device struct {
	id        ID
	name      Name
	direction Direction
	volume    Volume
	muted     bool
	isDefault bool
}

// Restore reconstruye un dispositivo tal y como lo informa el sistema. No hay
// constructor de alta: los dispositivos de sonido no se crean desde la app,
// se descubren.
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

// IsDefault indica si es el dispositivo que usa el sistema por defecto para
// su dirección.
func (d Device) IsDefault() bool { return d.isDefault }

func (d Device) IsZero() bool { return d.id.IsZero() }

// SetVolume ajusta el nivel. No quita el silencio: son dos controles
// independientes, igual que en cualquier mezclador.
func (d Device) SetVolume(v Volume) Device {
	d.volume = v
	return d
}

// AdjustVolume suma delta al nivel actual, acotando en los extremos. Es lo que
// hace una tecla de subir o bajar volumen: llegar al tope no es un error.
func (d Device) AdjustVolume(delta int) Device {
	return d.SetVolume(ClampVolume(d.volume.Level() + delta))
}

// SetMuted es idempotente, como cualquier interruptor.
func (d Device) SetMuted(muted bool) Device {
	d.muted = muted
	return d
}

func (d Device) ToggleMuted() Device {
	return d.SetMuted(!d.muted)
}
