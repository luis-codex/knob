package memory_test

import (
	"context"
	"errors"
	"testing"

	"settings-cli/internal/domain/bluetooth"
	"settings-cli/internal/infrastructure/memory"
)

func newDevice(t *testing.T, address, name string) bluetooth.Device {
	t.Helper()
	addr, err := bluetooth.NewAddress(address)
	if err != nil {
		t.Fatalf("NewAddress(%q): %v", address, err)
	}
	deviceName, err := bluetooth.NewName(name)
	if err != nil {
		t.Fatalf("NewName(%q): %v", name, err)
	}
	device, err := bluetooth.Discover(addr, deviceName, bluetooth.KindMouse)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	return device
}

func TestDeviceListKeepsOrder(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewDeviceRepository()

	addresses := []string{"AA:BB:CC:DD:EE:01", "AA:BB:CC:DD:EE:02", "AA:BB:CC:DD:EE:03"}
	for i, addr := range addresses {
		if err := repo.Save(ctx, newDevice(t, addr, string(rune('A'+i)))); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	got, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, want := range addresses {
		if got[i].Address().String() != want {
			t.Fatalf("position %d = %s, want %s", i, got[i].Address(), want)
		}
	}
}

// Saving the same address updates instead of duplicating, even when it came
// written in another format.
func TestDeviceSaveIsUpsert(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewDeviceRepository()

	original := newDevice(t, "aa:bb:cc:dd:ee:ff", "before")
	if err := repo.Save(ctx, original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	paired, err := original.Pair()
	if err != nil {
		t.Fatalf("Pair: %v", err)
	}
	if err := repo.Save(ctx, paired); err != nil {
		t.Fatalf("Save: %v", err)
	}

	all, _ := repo.List(ctx)
	if len(all) != 1 {
		t.Fatalf("duplicated: %d devices", len(all))
	}
	if all[0].State() != bluetooth.StatePaired {
		t.Errorf("state = %v, want paired", all[0].State())
	}
}

func TestDeviceSaveRejectsZero(t *testing.T) {
	err := memory.NewDeviceRepository().Save(context.Background(), bluetooth.Device{})
	if !errors.Is(err, bluetooth.ErrInvalidAddress) {
		t.Fatalf("error = %v, want ErrInvalidAddress", err)
	}
}

func TestDeviceFindByAddressAndDelete(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewDeviceRepository()
	device := newDevice(t, "AA:BB:CC:DD:EE:FF", "lookup")

	if err := repo.Save(ctx, device); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Lowercase: normalization must find it just the same.
	lower, _ := bluetooth.NewAddress("aa:bb:cc:dd:ee:ff")
	if _, err := repo.FindByAddress(ctx, lower); err != nil {
		t.Fatalf("FindByAddress with another format: %v", err)
	}

	if err := repo.Delete(ctx, device.Address()); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.FindByAddress(ctx, device.Address()); !errors.Is(err, bluetooth.ErrNotFound) {
		t.Fatalf("after delete: %v", err)
	}
	if err := repo.Delete(ctx, device.Address()); !errors.Is(err, bluetooth.ErrNotFound) {
		t.Fatalf("delete twice: %v", err)
	}
}

func TestDeviceListReturnsCopy(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewDeviceRepository()
	if err := repo.Save(ctx, newDevice(t, "AA:BB:CC:DD:EE:FF", "untouched")); err != nil {
		t.Fatalf("Save: %v", err)
	}

	first, _ := repo.List(ctx)
	first[0] = bluetooth.Device{}

	second, _ := repo.List(ctx)
	if second[0].Name().String() != "untouched" {
		t.Errorf("mutating the copy affected the store: %q", second[0].Name())
	}
}

func TestDeviceRespectsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := memory.NewDeviceRepository()
	if _, err := repo.List(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("List = %v", err)
	}
	if err := repo.Save(ctx, newDevice(t, "AA:BB:CC:DD:EE:FF", "x")); !errors.Is(err, context.Canceled) {
		t.Errorf("Save = %v", err)
	}
}
