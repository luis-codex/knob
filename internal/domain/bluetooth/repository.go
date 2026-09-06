package bluetooth

import "context"

// Repository is the persistence port: the domain declares it and
// infrastructure implements it.
type Repository interface {
	// Save inserts or updates the device.
	Save(ctx context.Context, d Device) error
	// Delete removes by address. Returns ErrNotFound if it does not exist.
	Delete(ctx context.Context, address Address) error
	// FindByAddress returns ErrNotFound if it does not exist.
	FindByAddress(ctx context.Context, address Address) (Device, error)
	// List returns the devices in discovery order.
	List(ctx context.Context) ([]Device, error)
}

// AdapterRepository holds the radio's state. It is a separate port because the
// adapter is another aggregate: it does not share a lifecycle with the
// devices.
type AdapterRepository interface {
	// Get returns the adapter. It never fails for not existing: if nothing is
	// stored it returns one that is off.
	Get(ctx context.Context) (Adapter, error)
	Save(ctx context.Context, a Adapter) error
}
