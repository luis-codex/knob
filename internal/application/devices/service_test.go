package devices_test

import (
	"context"
	"errors"
	"testing"

	"settings-cli/internal/application/devices"
	"settings-cli/internal/domain/bluetooth"
	"settings-cli/internal/domain/errs"
	"settings-cli/internal/infrastructure/memory"
	"settings-cli/internal/infrastructure/simulated"
)

const (
	addrA = "AA:BB:CC:DD:EE:FF"
	addrB = "11:22:33:44:55:66"
)

// newService wires the service with the real repositories: the in-memory ones
// are the production implementation, so no mock is needed. The radio starts on
// unless said otherwise.
func newService(t *testing.T) (*devices.Service, context.Context) {
	t.Helper()
	return newServiceWithAdapter(t, true)
}

func newServiceWithAdapter(t *testing.T, enabled bool) (*devices.Service, context.Context) {
	t.Helper()
	svc := devices.NewService(memory.NewDeviceRepository(), memory.NewAdapterRepository(enabled), simulated.NewScanner(0))
	return svc, context.Background()
}

func discovered(t *testing.T) (*devices.Service, context.Context) {
	t.Helper()
	svc, ctx := newService(t)
	if _, err := svc.Discover(ctx, addrA, "WH-1000XM4", "headphones"); err != nil {
		t.Fatalf("Discover: %v", err)
	}
	return svc, ctx
}

func TestDiscover(t *testing.T) {
	t.Run("records without pairing", func(t *testing.T) {
		svc, ctx := discovered(t)

		all, err := svc.List(ctx)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(all) != 1 || all[0].State() != bluetooth.StateDiscovered {
			t.Fatalf("List returned %d devices in state %v", len(all), all[0].State())
		}
	})

	t.Run("the same address twice is a conflict", func(t *testing.T) {
		svc, ctx := discovered(t)

		// Lowercase and with hyphens: it must be recognized as the same.
		_, err := svc.Discover(ctx, "aa-bb-cc-dd-ee-ff", "Another name", "mouse")
		if !errors.Is(err, bluetooth.ErrAlreadyKnown) {
			t.Fatalf("error = %v, want ErrAlreadyKnown", err)
		}
		if !errs.IsConflict(err) {
			t.Error("must arrive classified as a conflict")
		}

		all, _ := svc.List(ctx)
		if len(all) != 1 {
			t.Errorf("a duplicate create must add nothing: %d devices", len(all))
		}
		if all[0].Name().String() != "WH-1000XM4" {
			t.Errorf("the duplicate create overwrote the name: %q", all[0].Name())
		}
	})

	t.Run("rejects invalid data without persisting", func(t *testing.T) {
		tests := []struct {
			name            string
			addr, dev, kind string
			err             error
		}{
			{"invalid address", "not-a-mac", "X", "mouse", bluetooth.ErrInvalidAddress},
			{"empty name", addrA, "   ", "mouse", bluetooth.ErrEmptyName},
			{"unknown kind", addrA, "X", "toaster", bluetooth.ErrUnknownKind},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				svc, ctx := newService(t)

				if _, err := svc.Discover(ctx, tc.addr, tc.dev, tc.kind); !errors.Is(err, tc.err) {
					t.Fatalf("error = %v, want %v", err, tc.err)
				}
				if all, _ := svc.List(ctx); len(all) != 0 {
					t.Errorf("must persist nothing: %d devices", len(all))
				}
			})
		}
	})
}

// The full walk, checking that every step is saved.
func TestLifecyclePersists(t *testing.T) {
	svc, ctx := discovered(t)

	steps := []struct {
		name string
		do   func() (bluetooth.Device, error)
		want bluetooth.State
	}{
		{"pair", func() (bluetooth.Device, error) { return svc.Pair(ctx, addrA) }, bluetooth.StatePaired},
		{"connect", func() (bluetooth.Device, error) { return svc.Connect(ctx, addrA) }, bluetooth.StateConnected},
		{"disconnect", func() (bluetooth.Device, error) { return svc.Disconnect(ctx, addrA) }, bluetooth.StatePaired},
		{"unpair", func() (bluetooth.Device, error) { return svc.Unpair(ctx, addrA) }, bluetooth.StateDiscovered},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			if _, err := step.do(); err != nil {
				t.Fatalf("%s: %v", step.name, err)
			}

			all, _ := svc.List(ctx)
			if all[0].State() != step.want {
				t.Errorf("after %s the saved state is %v, want %v", step.name, all[0].State(), step.want)
			}
		})
	}
}

// An impossible transition must not modify what is saved.
func TestInvalidTransitionDoesNotPersist(t *testing.T) {
	svc, ctx := discovered(t)

	_, err := svc.Connect(ctx, addrA)
	if !errors.Is(err, bluetooth.ErrNotPaired) {
		t.Fatalf("error = %v, want ErrNotPaired", err)
	}
	if !errs.IsConflict(err) {
		t.Error("must arrive classified as a conflict")
	}

	all, _ := svc.List(ctx)
	if all[0].State() != bluetooth.StateDiscovered {
		t.Errorf("the state changed to %v", all[0].State())
	}
}

func TestNonexistentDevice(t *testing.T) {
	svc, ctx := newService(t)

	ops := map[string]func() error{
		"pair":       func() error { _, err := svc.Pair(ctx, addrA); return err },
		"connect":    func() error { _, err := svc.Connect(ctx, addrA); return err },
		"disconnect": func() error { _, err := svc.Disconnect(ctx, addrA); return err },
		"unpair":     func() error { _, err := svc.Unpair(ctx, addrA); return err },
		"rename":     func() error { _, err := svc.Rename(ctx, addrA, "X"); return err },
		"remove":     func() error { return svc.Remove(ctx, addrA) },
	}

	for name, op := range ops {
		t.Run(name, func(t *testing.T) {
			if err := op(); !errors.Is(err, bluetooth.ErrNotFound) {
				t.Fatalf("error = %v, want ErrNotFound", err)
			}
		})
	}
}

func TestRename(t *testing.T) {
	svc, ctx := discovered(t)

	if _, err := svc.Rename(ctx, addrA, "  Living room headphones "); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	all, _ := svc.List(ctx)
	if all[0].Name().String() != "Living room headphones" {
		t.Errorf("name = %q", all[0].Name())
	}

	if _, err := svc.Rename(ctx, addrA, ""); !errors.Is(err, bluetooth.ErrEmptyName) {
		t.Fatalf("empty name: %v", err)
	}
	all, _ = svc.List(ctx)
	if all[0].Name().String() != "Living room headphones" {
		t.Errorf("an invalid rename changed the name: %q", all[0].Name())
	}
}

func TestReportBattery(t *testing.T) {
	svc, ctx := discovered(t)

	// Without connecting, the device reports nothing.
	if _, err := svc.ReportBattery(ctx, addrA, 80); !errors.Is(err, bluetooth.ErrNotConnected) {
		t.Fatalf("not connected: %v", err)
	}

	if _, err := svc.Pair(ctx, addrA); err != nil {
		t.Fatalf("Pair: %v", err)
	}
	if _, err := svc.Connect(ctx, addrA); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	if _, err := svc.ReportBattery(ctx, addrA, 80); err != nil {
		t.Fatalf("ReportBattery: %v", err)
	}
	all, _ := svc.List(ctx)
	if !all[0].Battery().Known() || all[0].Battery().Level() != 80 {
		t.Errorf("battery = %d/%v", all[0].Battery().Level(), all[0].Battery().Known())
	}

	if _, err := svc.ReportBattery(ctx, addrA, 101); !errors.Is(err, bluetooth.ErrInvalidBattery) {
		t.Errorf("out-of-range level: %v", err)
	}
}

func TestRemove(t *testing.T) {
	svc, ctx := discovered(t)

	if err := svc.Remove(ctx, addrA); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if all, _ := svc.List(ctx); len(all) != 0 {
		t.Errorf("%d devices remain", len(all))
	}
	if err := svc.Remove(ctx, addrA); !errors.Is(err, bluetooth.ErrNotFound) {
		t.Errorf("delete twice: %v", err)
	}
}

func TestSearch(t *testing.T) {
	svc, ctx := newService(t)
	if _, err := svc.Discover(ctx, addrA, "WH-1000XM4", "headphones"); err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if _, err := svc.Discover(ctx, addrB, "MX Master 3S", "mouse"); err != nil {
		t.Fatalf("Discover: %v", err)
	}

	tests := []struct {
		name  string
		query string
		want  int
	}{
		{"empty returns all", "", 2},
		{"only spaces returns all", "  ", 2},
		{"substring", "master", 1},
		{"case-insensitive", "wh-1000", 1},
		{"no matches", "zzz", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := svc.Search(ctx, tc.query)
			if err != nil {
				t.Fatalf("Search: %v", err)
			}
			if len(got) != tc.want {
				t.Errorf("Search(%q) returned %d, want %d", tc.query, len(got), tc.want)
			}
		})
	}
}

func TestAdapter(t *testing.T) {
	t.Run("turning off disconnects what is connected", func(t *testing.T) {
		svc, ctx := discovered(t)
		if _, err := svc.Pair(ctx, addrA); err != nil {
			t.Fatalf("Pair: %v", err)
		}
		if _, err := svc.Connect(ctx, addrA); err != nil {
			t.Fatalf("Connect: %v", err)
		}

		adapter, err := svc.DisableAdapter(ctx)
		if err != nil {
			t.Fatalf("DisableAdapter: %v", err)
		}
		if adapter.Enabled() {
			t.Error("the adapter is still on")
		}

		all, _ := svc.List(ctx)
		if all[0].State() != bluetooth.StatePaired {
			t.Errorf("state = %v, want paired after turning off", all[0].State())
		}
	})

	t.Run("turning on does not reconnect", func(t *testing.T) {
		svc, ctx := discovered(t)
		if _, err := svc.Pair(ctx, addrA); err != nil {
			t.Fatalf("Pair: %v", err)
		}
		if _, err := svc.Connect(ctx, addrA); err != nil {
			t.Fatalf("Connect: %v", err)
		}
		if _, err := svc.DisableAdapter(ctx); err != nil {
			t.Fatalf("DisableAdapter: %v", err)
		}

		if _, err := svc.EnableAdapter(ctx); err != nil {
			t.Fatalf("EnableAdapter: %v", err)
		}

		all, _ := svc.List(ctx)
		if all[0].State() != bluetooth.StatePaired {
			t.Errorf("state = %v: turning on must not reconnect", all[0].State())
		}
	})

	t.Run("is idempotent", func(t *testing.T) {
		svc, ctx := newService(t)

		for i := 0; i < 2; i++ {
			a, err := svc.EnableAdapter(ctx)
			if err != nil || !a.Enabled() {
				t.Fatalf("EnableAdapter #%d = %v, %v", i, a.Enabled(), err)
			}
		}
		for i := 0; i < 2; i++ {
			a, err := svc.DisableAdapter(ctx)
			if err != nil || a.Enabled() {
				t.Fatalf("DisableAdapter #%d = %v, %v", i, a.Enabled(), err)
			}
		}
	})
}

// With the radio off you cannot pair or connect, but managing already-known
// devices keeps working.
func TestAdapterOffBlocksConnections(t *testing.T) {
	blocked := map[string]func(*devices.Service, context.Context) error{
		"pair":    func(s *devices.Service, ctx context.Context) error { _, err := s.Pair(ctx, addrA); return err },
		"connect": func(s *devices.Service, ctx context.Context) error { _, err := s.Connect(ctx, addrA); return err },
	}

	for name, op := range blocked {
		t.Run(name+" blocked", func(t *testing.T) {
			svc, ctx := newServiceWithAdapter(t, false)
			if _, err := svc.Discover(ctx, addrA, "X", "mouse"); err != nil {
				t.Fatalf("Discover: %v", err)
			}

			err := op(svc, ctx)
			if !errors.Is(err, bluetooth.ErrAdapterDisabled) {
				t.Fatalf("error = %v, want ErrAdapterDisabled", err)
			}
			if !errs.IsConflict(err) {
				t.Error("must arrive classified as a conflict")
			}

			all, _ := svc.List(ctx)
			if all[0].State() != bluetooth.StateDiscovered {
				t.Errorf("the state changed to %v", all[0].State())
			}
		})
	}

	t.Run("management is still allowed", func(t *testing.T) {
		svc, ctx := newServiceWithAdapter(t, false)
		if _, err := svc.Discover(ctx, addrA, "X", "mouse"); err != nil {
			t.Fatalf("Discover: %v", err)
		}

		if _, err := svc.Rename(ctx, addrA, "Renamed"); err != nil {
			t.Errorf("Rename with the radio off: %v", err)
		}
		if err := svc.Remove(ctx, addrA); err != nil {
			t.Errorf("Remove with the radio off: %v", err)
		}
	})
}

func TestScan(t *testing.T) {
	t.Run("records the unknown ones", func(t *testing.T) {
		svc, ctx := newService(t)

		added, err := svc.Scan(ctx)
		if err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if len(added) == 0 {
			t.Fatal("the first scan found nothing")
		}

		all, _ := svc.List(ctx)
		if len(all) != len(added) {
			t.Errorf("saved %d of %d found", len(all), len(added))
		}
		for _, d := range all {
			if d.State() != bluetooth.StateDiscovered {
				t.Errorf("%s arrived in state %v", d.Name(), d.State())
			}
		}
	})

	t.Run("is idempotent: no duplicates, no overwriting the known", func(t *testing.T) {
		svc, ctx := newService(t)

		first, err := svc.Scan(ctx)
		if err != nil {
			t.Fatalf("Scan: %v", err)
		}

		// Pair one and rename another before scanning again.
		target := first[0].Address().String()
		if _, err := svc.Pair(ctx, target); err != nil {
			t.Fatalf("Pair: %v", err)
		}
		if _, err := svc.Rename(ctx, target, "Mine"); err != nil {
			t.Fatalf("Rename: %v", err)
		}

		second, err := svc.Scan(ctx)
		if err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if len(second) != 0 {
			t.Errorf("the second scan added %d devices", len(second))
		}

		all, _ := svc.List(ctx)
		if len(all) != len(first) {
			t.Errorf("duplicated: %d, want %d", len(all), len(first))
		}

		found, err := svc.Search(ctx, "Mine")
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(found) != 1 || found[0].State() != bluetooth.StatePaired {
			t.Error("the scan overwrote the name or state of the known device")
		}
	})

	t.Run("does not scan with the radio off", func(t *testing.T) {
		svc, ctx := newServiceWithAdapter(t, false)

		if _, err := svc.Scan(ctx); !errors.Is(err, bluetooth.ErrAdapterDisabled) {
			t.Fatalf("error = %v, want ErrAdapterDisabled", err)
		}
		if all, _ := svc.List(ctx); len(all) != 0 {
			t.Errorf("recorded %d devices with the radio off", len(all))
		}
	})

	t.Run("respects a cancelled context", func(t *testing.T) {
		svc, _ := newService(t)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if _, err := svc.Scan(ctx); !errors.Is(err, context.Canceled) {
			t.Errorf("error = %v, want context.Canceled", err)
		}
	})
}

// registeringScanner mimics BlueZ: on scan, the repository itself comes to know
// what was found. This is the case that broke the new-device count.
type registeringScanner struct {
	repo    bluetooth.Repository
	devices []bluetooth.Device
}

func (s *registeringScanner) Scan(ctx context.Context) ([]bluetooth.Device, error) {
	for _, d := range s.devices {
		if err := s.repo.Save(ctx, d); err != nil {
			return nil, err
		}
	}
	return s.devices, nil
}

func TestScanCountsNewWithSelfRegisteringBackend(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewDeviceRepository()

	nearby := make([]bluetooth.Device, 0, 2)
	for _, spec := range []struct{ address, name string }{
		{addrA, "WH-1000XM4"},
		{addrB, "MX Master 3S"},
	} {
		address, err := bluetooth.NewAddress(spec.address)
		if err != nil {
			t.Fatalf("NewAddress: %v", err)
		}
		name, err := bluetooth.NewName(spec.name)
		if err != nil {
			t.Fatalf("NewName: %v", err)
		}
		device, err := bluetooth.Discover(address, name, bluetooth.KindUnknown)
		if err != nil {
			t.Fatalf("Discover: %v", err)
		}
		nearby = append(nearby, device)
	}

	svc := devices.NewService(repo, memory.NewAdapterRepository(true), &registeringScanner{repo: repo, devices: nearby})

	added, err := svc.Scan(ctx)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(added) != len(nearby) {
		t.Fatalf("reported %d new devices, want %d", len(added), len(nearby))
	}

	// The second pass no longer finds anything new.
	again, err := svc.Scan(ctx)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(again) != 0 {
		t.Errorf("the second scan reported %d new devices", len(again))
	}
}
