// Package simulated implements the hardware ports with no hardware behind
// them. It is for developing the interface while the real adapter does not
// exist.
package simulated

import (
	"context"
	"time"

	"settings-cli/internal/domain/bluetooth"
)

// Scanner always returns the same set of devices. Deduplicating against what
// is already known is the use case's job, not the scanner's: a real scanner
// does not know what you have paired either.
type Scanner struct {
	pool  []bluetooth.Device
	delay time.Duration
}

var _ bluetooth.Scanner = (*Scanner)(nil)

// NewScanner takes the delay that simulates the search. With 0 it responds
// instantly, which is what the tests want.
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

// defaultPool is the "nearby" devices. They are built with the domain
// constructors, so a bad value here never reaches the application: it simply
// does not appear.
func defaultPool() []bluetooth.Device {
	seed := []struct {
		address, name string
		kind          bluetooth.Kind
	}{
		{"AA:BB:CC:DD:EE:FF", "WH-1000XM4", bluetooth.KindHeadphones},
		{"11:22:33:44:55:66", "MX Master 3S", bluetooth.KindMouse},
		{"77:88:99:AA:BB:CC", "Keyboard K380", bluetooth.KindKeyboard},
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
