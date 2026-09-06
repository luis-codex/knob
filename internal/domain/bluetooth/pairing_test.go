package bluetooth_test

import (
	"errors"
	"testing"

	"settings-cli/internal/domain/bluetooth"
)

func TestPasskey(t *testing.T) {
	tests := []struct {
		name string
		code int
		want string
		err  error
	}{
		{name: "seis dígitos", code: 123456, want: "123456"},
		{name: "rellena con ceros", code: 42, want: "000042"},
		{name: "cero", code: 0, want: "000000"},
		{name: "máximo", code: 999999, want: "999999"},
		{name: "negativo", code: -1, err: bluetooth.ErrInvalidPasskey},
		{name: "demasiado grande", code: 1000000, err: bluetooth.ErrInvalidPasskey},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := bluetooth.NewPasskey(tc.code)

			if tc.err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("error = %v, se esperaba %v", err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if !got.Present() || got.String() != tc.want {
				t.Errorf("String() = %q, se esperaba %q", got, tc.want)
			}
		})
	}
}

func TestNoPasskey(t *testing.T) {
	p := bluetooth.NoPasskey()
	if p.Present() || p.String() != "" {
		t.Errorf("NoPasskey = %q/%v", p, p.Present())
	}

	// El valor cero debe comportarse igual: sin código, no "000000".
	var zero bluetooth.Passkey
	if zero.Present() {
		t.Error("el valor cero de Passkey debe ser 'sin código'")
	}
}

func TestNewPairingRequest(t *testing.T) {
	address, err := bluetooth.NewAddress("AA:BB:CC:DD:EE:FF")
	if err != nil {
		t.Fatalf("NewAddress: %v", err)
	}
	name, err := bluetooth.NewName("iPhone de Ana")
	if err != nil {
		t.Fatalf("NewName: %v", err)
	}
	passkey, err := bluetooth.NewPasskey(482913)
	if err != nil {
		t.Fatalf("NewPasskey: %v", err)
	}

	t.Run("válida", func(t *testing.T) {
		got, err := bluetooth.NewPairingRequest(address, name, passkey)
		if err != nil {
			t.Fatalf("NewPairingRequest: %v", err)
		}
		if got.Address().String() != "AA:BB:CC:DD:EE:FF" || got.Passkey().String() != "482913" {
			t.Errorf("solicitud = %s / %s", got.Address(), got.Passkey())
		}
	})

	t.Run("sin dirección", func(t *testing.T) {
		if _, err := bluetooth.NewPairingRequest(bluetooth.Address{}, name, passkey); !errors.Is(err, bluetooth.ErrInvalidAddress) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("sin nombre", func(t *testing.T) {
		if _, err := bluetooth.NewPairingRequest(address, bluetooth.Name{}, passkey); !errors.Is(err, bluetooth.ErrEmptyName) {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestAdapterVisibilidad(t *testing.T) {
	t.Run("apagado no puede ser visible", func(t *testing.T) {
		off := bluetooth.NewAdapter(false).SetVisible(true)

		if off.Discoverable() || off.Pairable() || off.Visible() {
			t.Error("una radio apagada no puede anunciarse ni aceptar solicitudes")
		}
	})

	t.Run("encendido sí", func(t *testing.T) {
		on := bluetooth.NewAdapter(true).SetVisible(true)

		if !on.Discoverable() || !on.Pairable() || !on.Visible() {
			t.Errorf("visible = %v/%v", on.Discoverable(), on.Pairable())
		}
	})

	t.Run("apagar quita la visibilidad", func(t *testing.T) {
		off := bluetooth.NewAdapter(true).SetVisible(true).Disable()

		if off.Discoverable() || off.Pairable() {
			t.Error("apagar debe dejar el equipo invisible")
		}

		// Volver a encender no la recupera sola: es una decisión aparte.
		if on := off.Enable(); on.Visible() {
			t.Error("encender no debe restaurar la visibilidad")
		}
	})

	t.Run("Visible exige ambas", func(t *testing.T) {
		half := bluetooth.NewAdapter(true).SetDiscoverable(true)

		if half.Visible() {
			t.Error("anunciarse sin aceptar emparejamientos no es ser visible")
		}
	})
}
