package bluez

import (
	"context"
	"strconv"
	"time"

	"settings-cli/internal/domain/bluetooth"
)

// Scanner pide a BlueZ que busque dispositivos durante un tiempo.
type Scanner struct {
	window time.Duration
	repo   *Repository
}

var _ bluetooth.Scanner = (*Scanner)(nil)

// NewScanner recibe cuánto dura la búsqueda.
func NewScanner(window time.Duration) *Scanner {
	return &Scanner{window: window, repo: NewRepository()}
}

// Scan lanza el descubrimiento y devuelve lo que BlueZ conoce al terminar.
//
// bluetoothctl no informa de qué encontró en esa pasada, así que se devuelve
// todo lo visible; quedarse solo con lo nuevo es cosa del caso de uso.
func (s *Scanner) Scan(ctx context.Context) ([]bluetooth.Device, error) {
	seconds := strconv.Itoa(int(s.window.Seconds()))
	if _, err := run(ctx, "--timeout", seconds, "scan", "on"); err != nil {
		return nil, err
	}
	return s.repo.List(ctx)
}
