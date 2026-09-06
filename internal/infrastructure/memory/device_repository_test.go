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

func TestDeviceListConservaOrden(t *testing.T) {
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
			t.Fatalf("posición %d = %s, se esperaba %s", i, got[i].Address(), want)
		}
	}
}

// Guardar la misma dirección actualiza en vez de duplicar, incluso si venía
// escrita con otro formato.
func TestDeviceSaveEsUpsert(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewDeviceRepository()

	original := newDevice(t, "aa:bb:cc:dd:ee:ff", "antes")
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
		t.Fatalf("se duplicó: %d dispositivos", len(all))
	}
	if all[0].State() != bluetooth.StatePaired {
		t.Errorf("estado = %v, se esperaba paired", all[0].State())
	}
}

func TestDeviceSaveRechazaCero(t *testing.T) {
	err := memory.NewDeviceRepository().Save(context.Background(), bluetooth.Device{})
	if !errors.Is(err, bluetooth.ErrInvalidAddress) {
		t.Fatalf("error = %v, se esperaba ErrInvalidAddress", err)
	}
}

func TestDeviceFindByAddressYDelete(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewDeviceRepository()
	device := newDevice(t, "AA:BB:CC:DD:EE:FF", "buscar")

	if err := repo.Save(ctx, device); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// En minúsculas: la normalización debe encontrarlo igual.
	lower, _ := bluetooth.NewAddress("aa:bb:cc:dd:ee:ff")
	if _, err := repo.FindByAddress(ctx, lower); err != nil {
		t.Fatalf("FindByAddress con otro formato: %v", err)
	}

	if err := repo.Delete(ctx, device.Address()); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.FindByAddress(ctx, device.Address()); !errors.Is(err, bluetooth.ErrNotFound) {
		t.Fatalf("tras borrar: %v", err)
	}
	if err := repo.Delete(ctx, device.Address()); !errors.Is(err, bluetooth.ErrNotFound) {
		t.Fatalf("borrar dos veces: %v", err)
	}
}

func TestDeviceListDevuelveCopia(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewDeviceRepository()
	if err := repo.Save(ctx, newDevice(t, "AA:BB:CC:DD:EE:FF", "intacto")); err != nil {
		t.Fatalf("Save: %v", err)
	}

	first, _ := repo.List(ctx)
	first[0] = bluetooth.Device{}

	second, _ := repo.List(ctx)
	if second[0].Name().String() != "intacto" {
		t.Errorf("mutar la copia afectó al almacén: %q", second[0].Name())
	}
}

func TestDeviceRespetaContextoCancelado(t *testing.T) {
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
