// Package bluetooth holds the Bluetooth-device aggregate, its value objects
// and the persistence port. It depends on no other layer.
package bluetooth

import (
	"strings"
	"unicode/utf8"
)

const (
	// MaxNameLength is the name limit, in characters.
	MaxNameLength = 64
	// addressOctets is the number of hex pairs in a MAC.
	addressOctets = 6
)

// State is the device's situation. The valid transitions are
// Discovered -> Paired -> Connected and their reverses.
type State int

const (
	// StateDiscovered: visible but not paired.
	StateDiscovered State = iota
	// StatePaired: paired with no active connection.
	StatePaired
	// StateConnected: paired and connected.
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

// IsPaired includes StateConnected: connecting requires having paired first.
func (s State) IsPaired() bool { return s == StatePaired || s == StateConnected }

// Kind is the device type. String returns a stable identifier, not a display
// label: translating it is the interface's job.
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

// ParseKind rebuilds a Kind from its identifier.
func ParseKind(s string) (Kind, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	for kind, name := range kindNames {
		if name == s {
			return kind, nil
		}
	}
	return KindUnknown, ErrUnknownKind
}

// Address is a MAC address. It identifies the device.
type Address struct {
	value string
}

// NewAddress validates and normalizes a MAC. It accepts ':' and '-' as
// separators and always returns the ':' form in uppercase, so the same
// address written two ways is the same Address.
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

// Name is a device's validated name.
type Name struct {
	value string
}

// NewName validates and trims a name.
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

// Battery is a charge level that may be unknown. The zero value is "unknown",
// not "0%", which is what makes not reporting it safe.
type Battery struct {
	level int
	known bool
}

// NewBattery accepts levels from 0 to 100.
func NewBattery(level int) (Battery, error) {
	if level < 0 || level > 100 {
		return Battery{}, ErrInvalidBattery
	}
	return Battery{level: level, known: true}, nil
}

// UnknownBattery is the level of a device that does not report it.
func UnknownBattery() Battery { return Battery{} }

// Level is only meaningful when Known is true.
func (b Battery) Level() int { return b.level }

func (b Battery) Known() bool { return b.known }

// Device is the aggregate. Its fields are unexported: only Discover and
// Restore build valid devices, and the state changes only through transitions.
type Device struct {
	address Address
	name    Name
	kind    Kind
	state   State
	battery Battery
}

// Discover records a just-seen device, not yet paired.
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

// Restore rebuilds an existing device from storage.
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

// Pair pairs a discovered device.
func (d Device) Pair() (Device, error) {
	if d.state.IsPaired() {
		return Device{}, ErrAlreadyPaired
	}
	d.state = StatePaired
	return d, nil
}

// Unpair forgets the device. It disconnects along the way if needed: forcing
// a disconnect first would only push the step onto the caller.
func (d Device) Unpair() (Device, error) {
	if !d.state.IsPaired() {
		return Device{}, ErrNotPaired
	}
	d.state = StateDiscovered
	d.battery = UnknownBattery()
	return d, nil
}

// Connect requires the device to be paired.
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

// Disconnect leaves the device paired.
func (d Device) Disconnect() (Device, error) {
	if d.state != StateConnected {
		return Device{}, ErrNotConnected
	}
	d.state = StatePaired
	return d, nil
}

// Rename changes the visible name without touching the state.
func (d Device) Rename(name Name) (Device, error) {
	if name.IsZero() {
		return Device{}, ErrEmptyName
	}
	d.name = name
	return d, nil
}

// ReportBattery updates the charge. It only makes sense while connected: a
// device that is not connected reports nothing.
func (d Device) ReportBattery(battery Battery) (Device, error) {
	if d.state != StateConnected {
		return Device{}, ErrNotConnected
	}
	d.battery = battery
	return d, nil
}
