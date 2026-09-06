// Package bluetooth contiene el agregado de dispositivos Bluetooth, sus
// objetos de valor y el puerto de persistencia. No depende de ninguna otra
// capa.
package bluetooth

import (
	"strings"
	"unicode/utf8"
)

const (
	// MaxNameLength es el límite del nombre, en caracteres.
	MaxNameLength = 64
	// addressOctets son los pares hexadecimales de una MAC.
	addressOctets = 6
)

// State es la situación del dispositivo. Las transiciones válidas son
// Discovered → Paired → Connected y sus inversas.
type State int

const (
	// StateDiscovered: visible pero no emparejado.
	StateDiscovered State = iota
	// StatePaired: emparejado y sin conexión activa.
	StatePaired
	// StateConnected: emparejado y conectado.
	StateConnected
)

func (s State) String() string {
	switch s {
	case StatePaired:
		return "paired"
	case StateConnected:
		return "connected"
	default:
		return "discovered"
	}
}

// IsPaired incluye StateConnected: conectarse exige haber emparejado antes.
func (s State) IsPaired() bool { return s == StatePaired || s == StateConnected }

// Kind es el tipo de dispositivo. String devuelve un identificador estable,
// no una etiqueta de pantalla: traducirlo es cosa de la interfaz.
type Kind int

const (
	KindUnknown Kind = iota
	KindHeadphones
	KindSpeaker
	KindMouse
	KindKeyboard
	KindPhone
)

var kindNames = map[Kind]string{
	KindUnknown:    "unknown",
	KindHeadphones: "headphones",
	KindSpeaker:    "speaker",
	KindMouse:      "mouse",
	KindKeyboard:   "keyboard",
	KindPhone:      "phone",
}

func (k Kind) String() string {
	if name, ok := kindNames[k]; ok {
		return name
	}
	return kindNames[KindUnknown]
}

// ParseKind reconstruye un Kind desde su identificador.
func ParseKind(s string) (Kind, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	for kind, name := range kindNames {
		if name == s {
			return kind, nil
		}
	}
	return KindUnknown, ErrUnknownKind
}

// Address es una dirección MAC. Identifica al dispositivo.
type Address struct {
	value string
}

// NewAddress valida y normaliza una MAC. Acepta ':' y '-' como separadores y
// devuelve siempre el formato con ':' en mayúsculas, para que la misma
// dirección escrita de dos maneras sea el mismo Address.
func NewAddress(s string) (Address, error) {
	parts := strings.Split(strings.ReplaceAll(strings.TrimSpace(s), "-", ":"), ":")
	if len(parts) != addressOctets {
		return Address{}, ErrInvalidAddress
	}

	for i, octet := range parts {
		if len(octet) != 2 || !isHex(octet) {
			return Address{}, ErrInvalidAddress
		}
		parts[i] = strings.ToUpper(octet)
	}

	return Address{value: strings.Join(parts, ":")}, nil
}

func isHex(s string) bool {
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9', r >= 'a' && r <= 'f', r >= 'A' && r <= 'F':
		default:
			return false
		}
	}
	return true
}

func (a Address) String() string { return a.value }

func (a Address) IsZero() bool { return a.value == "" }

func (a Address) Equals(other Address) bool { return a.value == other.value }

// Name es el nombre validado de un dispositivo.
type Name struct {
	value string
}

// NewName valida y recorta un nombre.
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

// Battery es un nivel de carga que puede ser desconocido. El valor cero es
// "desconocido" y no "0 %", que es lo que hace seguro no informarlo.
type Battery struct {
	level int
	known bool
}

// NewBattery acepta niveles de 0 a 100.
func NewBattery(level int) (Battery, error) {
	if level < 0 || level > 100 {
		return Battery{}, ErrInvalidBattery
	}
	return Battery{level: level, known: true}, nil
}

// UnknownBattery es el nivel de un dispositivo que no lo informa.
func UnknownBattery() Battery { return Battery{} }

// Level solo es significativo si Known es true.
func (b Battery) Level() int { return b.level }

func (b Battery) Known() bool { return b.known }

// Device es el agregado. Sus campos no se exportan: solo Discover y Restore
// construyen dispositivos válidos, y el estado solo cambia por transiciones.
type Device struct {
	address Address
	name    Name
	kind    Kind
	state   State
	battery Battery
}

// Discover registra un dispositivo recién visto, sin emparejar.
func Discover(address Address, name Name, kind Kind) (Device, error) {
	switch {
	case address.IsZero():
		return Device{}, ErrInvalidAddress
	case name.IsZero():
		return Device{}, ErrEmptyName
	}

	return Device{
		address: address,
		name:    name,
		kind:    kind,
		state:   StateDiscovered,
		battery: UnknownBattery(),
	}, nil
}

// Restore reconstruye un dispositivo existente desde el almacenamiento.
func Restore(address Address, name Name, kind Kind, state State, battery Battery) (Device, error) {
	device, err := Discover(address, name, kind)
	if err != nil {
		return Device{}, err
	}

	device.state, device.battery = state, battery
	return device, nil
}

func (d Device) Address() Address { return d.address }

func (d Device) Name() Name { return d.name }

func (d Device) Kind() Kind { return d.kind }

func (d Device) State() State { return d.state }

func (d Device) Battery() Battery { return d.battery }

func (d Device) IsZero() bool { return d.address.IsZero() }

// Pair empareja un dispositivo descubierto.
func (d Device) Pair() (Device, error) {
	if d.state.IsPaired() {
		return Device{}, ErrAlreadyPaired
	}
	d.state = StatePaired
	return d, nil
}

// Unpair olvida el dispositivo. Desconecta de paso si hacía falta: obligar a
// desconectar antes solo trasladaría el paso al llamante.
func (d Device) Unpair() (Device, error) {
	if !d.state.IsPaired() {
		return Device{}, ErrNotPaired
	}
	d.state = StateDiscovered
	d.battery = UnknownBattery()
	return d, nil
}

// Connect exige que el dispositivo esté emparejado.
func (d Device) Connect() (Device, error) {
	switch d.state {
	case StateConnected:
		return Device{}, ErrAlreadyConnected
	case StateDiscovered:
		return Device{}, ErrNotPaired
	}

	d.state = StateConnected
	return d, nil
}

// Disconnect deja el dispositivo emparejado.
func (d Device) Disconnect() (Device, error) {
	if d.state != StateConnected {
		return Device{}, ErrNotConnected
	}
	d.state = StatePaired
	return d, nil
}

// Rename cambia el nombre visible sin tocar el estado.
func (d Device) Rename(name Name) (Device, error) {
	if name.IsZero() {
		return Device{}, ErrEmptyName
	}
	d.name = name
	return d, nil
}

// ReportBattery actualiza la carga. Solo tiene sentido conectado: un
// dispositivo que no lo está no informa de nada.
func (d Device) ReportBattery(battery Battery) (Device, error) {
	if d.state != StateConnected {
		return Device{}, ErrNotConnected
	}
	d.battery = battery
	return d, nil
}
