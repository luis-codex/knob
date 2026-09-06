// Package simulated implementa los puertos de hardware sin hardware detrás.
// Sirve para desarrollar la interfaz mientras no exista el adaptador real.
package simulated

import (
	"context"
	"time"

	"settings-cli/internal/domain/bluetooth"
)

// Scanner devuelve siempre el mismo conjunto de dispositivos. Deduplicar
// contra lo ya conocido es cosa del caso de uso, no del escáner: un escáner
// real tampoco sabe qué tienes emparejado.
type Scanner struct {
	pool  []bluetooth.Device
	delay time.Duration
}

var _ bluetooth.Scanner = (*Scanner)(nil)

// NewScanner recibe el retardo que simula la búsqueda. Con 0 responde al
// instante, que es lo que quieren los tests.
func NewScanner(delay time.Duration) *Scanner {
	return &Scanner{pool: defaultPool(), delay: delay}
}

func (s *Scanner) Scan(ctx context.Context) ([]bluetooth.Device, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(s.delay):
	}

	out := make([]bluetooth.Device, len(s.pool))
	copy(out, s.pool)
	return out, nil
}

// defaultPool son los dispositivos "cercanos". Se construyen con los
// constructores del dominio, así que un dato mal escrito aquí no llega a la
// aplicación: simplemente no aparece.
func defaultPool() []bluetooth.Device {
	seed := []struct {
		address, name string
		kind          bluetooth.Kind
	}{
		{"AA:BB:CC:DD:EE:FF", "WH-1000XM4", bluetooth.KindHeadphones},
		{"11:22:33:44:55:66", "MX Master 3S", bluetooth.KindMouse},
		{"77:88:99:AA:BB:CC", "Teclado K380", bluetooth.KindKeyboard},
		{"DE:AD:BE:EF:00:11", "Pixel Buds", bluetooth.KindHeadphones},
		{"01:23:45:67:89:AB", "JBL Flip 6", bluetooth.KindSpeaker},
		{"CA:FE:BA:BE:12:34", "Galaxy S24", bluetooth.KindPhone},
	}

	out := make([]bluetooth.Device, 0, len(seed))
	for _, s := range seed {
		address, err := bluetooth.NewAddress(s.address)
		if err != nil {
			continue
		}
		name, err := bluetooth.NewName(s.name)
		if err != nil {
			continue
		}
		device, err := bluetooth.Discover(address, name, s.kind)
		if err != nil {
			continue
		}
		out = append(out, device)
	}
	return out
}
