package memory_test

import (
	"context"
	"testing"

	"knob/internal/domain/bluetooth"
	"knob/internal/infrastructure/memory"
)

func TestSeedDevicesPopulatesTheFakeBackend(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewDeviceRepository()

	if err := memory.SeedDevices(ctx, repo); err != nil {
		t.Fatalf("SeedDevices: %v", err)
	}

	all, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(all) == 0 {
		t.Fatal("seeded nothing")
	}

	// At least one fixture must come out connected with a battery reading:
	// that is what makes the screen worth looking at on a dev run.
	var connectedWithBattery bool
	for _, d := range all {
		if d.State() == bluetooth.StateConnected && d.Battery().Known() {
			connectedWithBattery = true
		}
	}
	if !connectedWithBattery {
		t.Error("no seeded device is connected with a battery level")
	}
}

func TestSeedDevicesIsIdempotent(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewDeviceRepository()

	if err := memory.SeedDevices(ctx, repo); err != nil {
		t.Fatalf("first SeedDevices: %v", err)
	}
	first, _ := repo.List(ctx)

	if err := memory.SeedDevices(ctx, repo); err != nil {
		t.Fatalf("second SeedDevices: %v", err)
	}
	second, _ := repo.List(ctx)

	if len(second) != len(first) {
		t.Fatalf("second run changed the count: %d -> %d", len(first), len(second))
	}
}
