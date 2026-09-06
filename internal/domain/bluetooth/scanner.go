package bluetooth

import "context"

// Scanner discovers nearby devices.
//
// It is an output port like Repository, but toward the hardware rather than
// storage: it returns what is around, without knowing what is already known.
type Scanner interface {
	// Scan returns the visible devices, all in StateDiscovered.
	Scan(ctx context.Context) ([]Device, error)
}
