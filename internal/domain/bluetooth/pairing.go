package bluetooth

import (
	"context"
	"fmt"
)

// passkeyDigits es la longitud del código que muestran ambos extremos.
const passkeyDigits = 6

// Passkey es el código que los dos dispositivos deben ver igual. Puede no
// haberlo: el emparejamiento "Just Works" no muestra ninguno.
type Passkey struct {
	value   int
	present bool
}

// NewPasskey acepta códigos de 6 dígitos.
func NewPasskey(code int) (Passkey, error) {
	if code < 0 || code > 999999 {
		return Passkey{}, ErrInvalidPasskey
	}
	return Passkey{value: code, present: true}, nil
}

// NoPasskey es el emparejamiento sin confirmación de código.
func NoPasskey() Passkey { return Passkey{} }

func (p Passkey) Present() bool { return p.present }

// String rellena con ceros a la izquierda: un código es de 6 dígitos aunque
// empiece por cero.
func (p Passkey) String() string {
	if !p.present {
		return ""
	}
	return fmt.Sprintf("%0*d", passkeyDigits, p.value)
}

// PairingRequest es una solicitud de emparejamiento entrante: otro
// dispositivo pide emparejarse con este equipo.
//
// No es una entidad persistente sino un hecho puntual, así que no tiene
// repositorio: llega, se responde y se acaba.
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

// PairingAgent atiende las solicitudes de emparejamiento entrantes.
//
// Es el primer puerto de entrada del sistema: el resto se consultan, este
// empuja. Por eso entrega un canal en vez de devolver un valor.
type PairingAgent interface {
	// Requests entrega las solicitudes conforme llegan. El canal se cierra
	// cuando el contexto termina.
	Requests(ctx context.Context) (<-chan PairingRequest, error)
	// Accept confirma el emparejamiento solicitado.
	Accept(ctx context.Context, address Address) error
	// Reject lo rechaza.
	Reject(ctx context.Context, address Address) error
}
