package bluetooth

import (
	"context"
	"fmt"
)

// passkeyDigits is the length of the code both ends display.
const passkeyDigits = 6

// Passkey is the code both devices must see identically. There may be none:
// "Just Works" pairing shows no code.
type Passkey struct {
	value   int
	present bool
}

// NewPasskey accepts 6-digit codes.
func NewPasskey(code int) (Passkey, error) {
	if code < 0 || code > 999999 {
		return Passkey{}, ErrInvalidPasskey
	}
	return Passkey{value: code, present: true}, nil
}

// NoPasskey is pairing without code confirmation.
func NoPasskey() Passkey { return Passkey{} }

func (p Passkey) Present() bool { return p.present }

// String left-pads with zeros: a code is 6 digits even when it starts with
// zero.
func (p Passkey) String() string {
	if !p.present {
		return ""
	}
	return fmt.Sprintf("%0*d", passkeyDigits, p.value)
}

// PairingRequest is an incoming pairing request: another device asks to pair
// with this machine.
//
// It is not a persistent entity but a one-off fact, so it has no repository:
// it arrives, is answered and is over.
type PairingRequest struct {
	address Address
	name    Name
	passkey Passkey
}

func NewPairingRequest(address Address, name Name, passkey Passkey) (PairingRequest, error) {
	switch {
	case address.IsZero():
		return PairingRequest{}, ErrInvalidAddress
	case name.IsZero():
		return PairingRequest{}, ErrEmptyName
	}
	return PairingRequest{address: address, name: name, passkey: passkey}, nil
}

func (r PairingRequest) Address() Address { return r.address }

func (r PairingRequest) Name() Name { return r.name }

func (r PairingRequest) Passkey() Passkey { return r.passkey }

func (r PairingRequest) IsZero() bool { return r.address.IsZero() }

// PairingAgent handles incoming pairing requests.
//
// It is the system's first input port: the rest are queried, this one pushes.
// That is why it hands back a channel instead of returning a value.
type PairingAgent interface {
	// Requests delivers requests as they arrive. The channel closes when the
	// context ends.
	Requests(ctx context.Context) (<-chan PairingRequest, error)
	// Accept confirms the requested pairing.
	Accept(ctx context.Context, address Address) error
	// Reject turns it down.
	Reject(ctx context.Context, address Address) error
}
