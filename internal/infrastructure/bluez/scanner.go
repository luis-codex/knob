package bluez

import (
	"context"
	"strconv"
	"time"

	"knob/internal/domain/bluetooth"
)

// Scanner asks BlueZ to look for devices for a while.
type Scanner struct {
	window time.Duration
	repo   *Repository
}

var _ bluetooth.Scanner = (*Scanner)(nil)

// NewScanner takes how long the search lasts.
func NewScanner(window time.Duration) *Scanner {
	return &Scanner{window: window, repo: NewRepository()}
}

// Scan starts discovery and returns what BlueZ knows when it finishes.
//
// bluetoothctl does not report what it found in that pass, so everything
// visible is returned; keeping only the new ones is the use case's job.
func (s *Scanner) Scan(ctx context.Context) ([]bluetooth.Device, error) {
	seconds := strconv.Itoa(int(s.window.Seconds()))
	if _, err := run(ctx, "--timeout", seconds, "scan", "on"); err != nil {
		return nil, err
	}
	return s.repo.List(ctx)
}
