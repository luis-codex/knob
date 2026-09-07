package devices_test

import (
	"context"
	"errors"
	"testing"

	"knob/internal/application/devices"
	"knob/internal/domain/bluetooth"
	"knob/internal/domain/errs"
	"knob/internal/infrastructure/memory"
	"knob/internal/infrastructure/simulated"
)

const (
	addrA = "AA:BB:CC:DD:EE:FF"
	addrB = "11:22:33:44:55:66"
)

// newUseCases wires the use cases with the real repositories: the in-memory
// ones are the production implementation, so no mock is needed. The radio
// starts on unless said otherwise.
func newUseCases(t *testing.T) (devices.UseCases, context.Context) {
	t.Helper()
	return newUseCasesWithAdapter(t, true)
}

func newUseCasesWithAdapter(t *testing.T, enabled bool) (devices.UseCases, context.Context) {
	t.Helper()
	uc := devices.NewUseCases(devices.Deps{
		Repo:     memory.NewDeviceRepository(),
		Adapters: memory.NewAdapterRepository(enabled),
		Scanner:  simulated.NewScanner(0),
	})
	return uc, context.Background()
}

func discovered(t *testing.T) (devices.UseCases, context.Context) {
	t.Helper()
	uc, ctx := newUseCases(t)
	if _, err := uc.DiscoverDevice.Execute(ctx, devices.DiscoverDeviceCommand{
		Address: addrA, Name: "WH-1000XM4", Kind: "headphones",
	}); err != nil {
		t.Fatalf("Discover: %v", err)
	}
	return uc, ctx
}

// listDevices reads the whole store back, the way most assertions here check
// that a step was persisted.
func listDevices(t *testing.T, ctx context.Context, uc devices.UseCases) []bluetooth.Device {
	t.Helper()
	res, err := uc.ListDevices.Execute(ctx, devices.ListDevicesCommand{})
	if err != nil {
		t.Fatalf("ListDevices: %v", err)
	}
	return res.Devices
}

func TestDiscover(t *testing.T) {
	t.Run("records without pairing", func(t *testing.T) {
		uc, ctx := discovered(t)

		all := listDevices(t, ctx, uc)
		if len(all) != 1 || all[0].State() != bluetooth.StateDiscovered {
			t.Fatalf("List returned %d devices in state %v", len(all), all[0].State())
		}
	})

	t.Run("the same address twice is a conflict", func(t *testing.T) {
		uc, ctx := discovered(t)

		// Lowercase and with hyphens: it must be recognized as the same.
		_, err := uc.DiscoverDevice.Execute(ctx, devices.DiscoverDeviceCommand{
			Address: "aa-bb-cc-dd-ee-ff", Name: "Another name", Kind: "mouse",
		})
		if !errors.Is(err, bluetooth.ErrAlreadyKnown) {
			t.Fatalf("error = %v, want ErrAlreadyKnown", err)
		}
		if !errs.IsConflict(err) {
			t.Error("must arrive classified as a conflict")
		}

		all := listDevices(t, ctx, uc)
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
				uc, ctx := newUseCases(t)

				_, err := uc.DiscoverDevice.Execute(ctx, devices.DiscoverDeviceCommand{
					Address: tc.addr, Name: tc.dev, Kind: tc.kind,
				})
				if !errors.Is(err, tc.err) {
					t.Fatalf("error = %v, want %v", err, tc.err)
				}
				if all := listDevices(t, ctx, uc); len(all) != 0 {
					t.Errorf("must persist nothing: %d devices", len(all))
				}
			})
		}
	})
}

// The full walk, checking that every step is saved.
func TestLifecyclePersists(t *testing.T) {
	uc, ctx := discovered(t)

	steps := []struct {
		name string
		do   func() error
		want bluetooth.State
	}{
		{"pair", func() error {
			_, err := uc.PairDevice.Execute(ctx, devices.PairDeviceCommand{Address: addrA})
			return err
		}, bluetooth.StatePaired},
		{"connect", func() error {
			_, err := uc.ConnectDevice.Execute(ctx, devices.ConnectDeviceCommand{Address: addrA})
			return err
		}, bluetooth.StateConnected},
		{"disconnect", func() error {
			_, err := uc.DisconnectDevice.Execute(ctx, devices.DisconnectDeviceCommand{Address: addrA})
			return err
		}, bluetooth.StatePaired},
		{"unpair", func() error {
			_, err := uc.UnpairDevice.Execute(ctx, devices.UnpairDeviceCommand{Address: addrA})
			return err
		}, bluetooth.StateDiscovered},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			if err := step.do(); err != nil {
				t.Fatalf("%s: %v", step.name, err)
			}

			all := listDevices(t, ctx, uc)
			if all[0].State() != step.want {
				t.Errorf("after %s the saved state is %v, want %v", step.name, all[0].State(), step.want)
			}
		})
	}
}

// An impossible transition must not modify what is saved.
func TestInvalidTransitionDoesNotPersist(t *testing.T) {
	uc, ctx := discovered(t)

	_, err := uc.ConnectDevice.Execute(ctx, devices.ConnectDeviceCommand{Address: addrA})
	if !errors.Is(err, bluetooth.ErrNotPaired) {
		t.Fatalf("error = %v, want ErrNotPaired", err)
	}
	if !errs.IsConflict(err) {
		t.Error("must arrive classified as a conflict")
	}

	all := listDevices(t, ctx, uc)
	if all[0].State() != bluetooth.StateDiscovered {
		t.Errorf("the state changed to %v", all[0].State())
	}
}

func TestNonexistentDevice(t *testing.T) {
	uc, ctx := newUseCases(t)

	ops := map[string]func() error{
		"pair": func() error {
			_, err := uc.PairDevice.Execute(ctx, devices.PairDeviceCommand{Address: addrA})
			return err
		},
		"connect": func() error {
			_, err := uc.ConnectDevice.Execute(ctx, devices.ConnectDeviceCommand{Address: addrA})
			return err
		},
		"disconnect": func() error {
			_, err := uc.DisconnectDevice.Execute(ctx, devices.DisconnectDeviceCommand{Address: addrA})
			return err
		},
		"unpair": func() error {
			_, err := uc.UnpairDevice.Execute(ctx, devices.UnpairDeviceCommand{Address: addrA})
			return err
		},
		"rename": func() error {
			_, err := uc.RenameDevice.Execute(ctx, devices.RenameDeviceCommand{Address: addrA, Name: "X"})
			return err
		},
		"remove": func() error {
			_, err := uc.RemoveDevice.Execute(ctx, devices.RemoveDeviceCommand{Address: addrA})
			return err
		},
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
	uc, ctx := discovered(t)

	if _, err := uc.RenameDevice.Execute(ctx, devices.RenameDeviceCommand{
		Address: addrA, Name: "  Living room headphones ",
	}); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	all := listDevices(t, ctx, uc)
	if all[0].Name().String() != "Living room headphones" {
		t.Errorf("name = %q", all[0].Name())
	}

	if _, err := uc.RenameDevice.Execute(ctx, devices.RenameDeviceCommand{
		Address: addrA, Name: "",
	}); !errors.Is(err, bluetooth.ErrEmptyName) {
		t.Fatalf("empty name: %v", err)
	}
	all = listDevices(t, ctx, uc)
	if all[0].Name().String() != "Living room headphones" {
		t.Errorf("an invalid rename changed the name: %q", all[0].Name())
	}
}

func TestReportBattery(t *testing.T) {
	uc, ctx := discovered(t)

	// Without connecting, the device reports nothing.
	if _, err := uc.ReportBattery.Execute(ctx, devices.ReportBatteryCommand{Address: addrA, Level: 80}); !errors.Is(err, bluetooth.ErrNotConnected) {
		t.Fatalf("not connected: %v", err)
	}

	if _, err := uc.PairDevice.Execute(ctx, devices.PairDeviceCommand{Address: addrA}); err != nil {
		t.Fatalf("Pair: %v", err)
	}
	if _, err := uc.ConnectDevice.Execute(ctx, devices.ConnectDeviceCommand{Address: addrA}); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	if _, err := uc.ReportBattery.Execute(ctx, devices.ReportBatteryCommand{Address: addrA, Level: 80}); err != nil {
		t.Fatalf("ReportBattery: %v", err)
	}
	all := listDevices(t, ctx, uc)
	if !all[0].Battery().Known() || all[0].Battery().Level() != 80 {
		t.Errorf("battery = %d/%v", all[0].Battery().Level(), all[0].Battery().Known())
	}

	if _, err := uc.ReportBattery.Execute(ctx, devices.ReportBatteryCommand{Address: addrA, Level: 101}); !errors.Is(err, bluetooth.ErrInvalidBattery) {
		t.Errorf("out-of-range level: %v", err)
	}
}

func TestRemove(t *testing.T) {
	uc, ctx := discovered(t)

	if _, err := uc.RemoveDevice.Execute(ctx, devices.RemoveDeviceCommand{Address: addrA}); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if all := listDevices(t, ctx, uc); len(all) != 0 {
		t.Errorf("%d devices remain", len(all))
	}
	if _, err := uc.RemoveDevice.Execute(ctx, devices.RemoveDeviceCommand{Address: addrA}); !errors.Is(err, bluetooth.ErrNotFound) {
		t.Errorf("delete twice: %v", err)
	}
}

func TestSearch(t *testing.T) {
	uc, ctx := newUseCases(t)
	if _, err := uc.DiscoverDevice.Execute(ctx, devices.DiscoverDeviceCommand{
		Address: addrA, Name: "WH-1000XM4", Kind: "headphones",
	}); err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if _, err := uc.DiscoverDevice.Execute(ctx, devices.DiscoverDeviceCommand{
		Address: addrB, Name: "MX Master 3S", Kind: "mouse",
	}); err != nil {
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
			got, err := uc.SearchDevices.Execute(ctx, devices.SearchDevicesCommand{Query: tc.query})
			if err != nil {
				t.Fatalf("Search: %v", err)
			}
			if len(got.Devices) != tc.want {
				t.Errorf("Search(%q) returned %d, want %d", tc.query, len(got.Devices), tc.want)
			}
		})
	}
}

func TestAdapter(t *testing.T) {
	t.Run("turning off disconnects what is connected", func(t *testing.T) {
		uc, ctx := discovered(t)
		if _, err := uc.PairDevice.Execute(ctx, devices.PairDeviceCommand{Address: addrA}); err != nil {
			t.Fatalf("Pair: %v", err)
		}
		if _, err := uc.ConnectDevice.Execute(ctx, devices.ConnectDeviceCommand{Address: addrA}); err != nil {
			t.Fatalf("Connect: %v", err)
		}

		res, err := uc.DisableAdapter.Execute(ctx, devices.DisableAdapterCommand{})
		if err != nil {
			t.Fatalf("DisableAdapter: %v", err)
		}
		if res.Adapter.Enabled() {
			t.Error("the adapter is still on")
		}

		all := listDevices(t, ctx, uc)
		if all[0].State() != bluetooth.StatePaired {
			t.Errorf("state = %v, want paired after turning off", all[0].State())
		}
	})

	t.Run("turning on does not reconnect", func(t *testing.T) {
		uc, ctx := discovered(t)
		if _, err := uc.PairDevice.Execute(ctx, devices.PairDeviceCommand{Address: addrA}); err != nil {
			t.Fatalf("Pair: %v", err)
		}
		if _, err := uc.ConnectDevice.Execute(ctx, devices.ConnectDeviceCommand{Address: addrA}); err != nil {
			t.Fatalf("Connect: %v", err)
		}
		if _, err := uc.DisableAdapter.Execute(ctx, devices.DisableAdapterCommand{}); err != nil {
			t.Fatalf("DisableAdapter: %v", err)
		}

		if _, err := uc.EnableAdapter.Execute(ctx, devices.EnableAdapterCommand{}); err != nil {
			t.Fatalf("EnableAdapter: %v", err)
		}

		all := listDevices(t, ctx, uc)
		if all[0].State() != bluetooth.StatePaired {
			t.Errorf("state = %v: turning on must not reconnect", all[0].State())
		}
	})

	t.Run("is idempotent", func(t *testing.T) {
		uc, ctx := newUseCases(t)

		for i := 0; i < 2; i++ {
			res, err := uc.EnableAdapter.Execute(ctx, devices.EnableAdapterCommand{})
			if err != nil || !res.Adapter.Enabled() {
				t.Fatalf("EnableAdapter #%d = %v, %v", i, res.Adapter.Enabled(), err)
			}
		}
		for i := 0; i < 2; i++ {
			res, err := uc.DisableAdapter.Execute(ctx, devices.DisableAdapterCommand{})
			if err != nil || res.Adapter.Enabled() {
				t.Fatalf("DisableAdapter #%d = %v, %v", i, res.Adapter.Enabled(), err)
			}
		}
	})
}

// With the radio off you cannot pair or connect, but managing already-known
// devices keeps working.
func TestAdapterOffBlocksConnections(t *testing.T) {
	blocked := map[string]func(devices.UseCases, context.Context) error{
		"pair": func(uc devices.UseCases, ctx context.Context) error {
			_, err := uc.PairDevice.Execute(ctx, devices.PairDeviceCommand{Address: addrA})
			return err
		},
		"connect": func(uc devices.UseCases, ctx context.Context) error {
			_, err := uc.ConnectDevice.Execute(ctx, devices.ConnectDeviceCommand{Address: addrA})
			return err
		},
	}

	for name, op := range blocked {
		t.Run(name+" blocked", func(t *testing.T) {
			uc, ctx := newUseCasesWithAdapter(t, false)
			if _, err := uc.DiscoverDevice.Execute(ctx, devices.DiscoverDeviceCommand{
				Address: addrA, Name: "X", Kind: "mouse",
			}); err != nil {
				t.Fatalf("Discover: %v", err)
			}

			err := op(uc, ctx)
			if !errors.Is(err, bluetooth.ErrAdapterDisabled) {
				t.Fatalf("error = %v, want ErrAdapterDisabled", err)
			}
			if !errs.IsConflict(err) {
				t.Error("must arrive classified as a conflict")
			}

			all := listDevices(t, ctx, uc)
			if all[0].State() != bluetooth.StateDiscovered {
				t.Errorf("the state changed to %v", all[0].State())
			}
		})
	}

	t.Run("management is still allowed", func(t *testing.T) {
		uc, ctx := newUseCasesWithAdapter(t, false)
		if _, err := uc.DiscoverDevice.Execute(ctx, devices.DiscoverDeviceCommand{
			Address: addrA, Name: "X", Kind: "mouse",
		}); err != nil {
			t.Fatalf("Discover: %v", err)
		}

		if _, err := uc.RenameDevice.Execute(ctx, devices.RenameDeviceCommand{Address: addrA, Name: "Renamed"}); err != nil {
			t.Errorf("Rename with the radio off: %v", err)
		}
		if _, err := uc.RemoveDevice.Execute(ctx, devices.RemoveDeviceCommand{Address: addrA}); err != nil {
			t.Errorf("Remove with the radio off: %v", err)
		}
	})
}

func TestScan(t *testing.T) {
	t.Run("records the unknown ones", func(t *testing.T) {
		uc, ctx := newUseCases(t)

		res, err := uc.ScanDevices.Execute(ctx, devices.ScanDevicesCommand{})
		if err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if len(res.Devices) == 0 {
			t.Fatal("the first scan found nothing")
		}

		all := listDevices(t, ctx, uc)
		if len(all) != len(res.Devices) {
			t.Errorf("saved %d of %d found", len(all), len(res.Devices))
		}
		for _, d := range all {
			if d.State() != bluetooth.StateDiscovered {
				t.Errorf("%s arrived in state %v", d.Name(), d.State())
			}
		}
	})

	t.Run("is idempotent: no duplicates, no overwriting the known", func(t *testing.T) {
		uc, ctx := newUseCases(t)

		first, err := uc.ScanDevices.Execute(ctx, devices.ScanDevicesCommand{})
		if err != nil {
			t.Fatalf("Scan: %v", err)
		}

		// Pair one and rename another before scanning again.
		target := first.Devices[0].Address().String()
		if _, err := uc.PairDevice.Execute(ctx, devices.PairDeviceCommand{Address: target}); err != nil {
			t.Fatalf("Pair: %v", err)
		}
		if _, err := uc.RenameDevice.Execute(ctx, devices.RenameDeviceCommand{Address: target, Name: "Mine"}); err != nil {
			t.Fatalf("Rename: %v", err)
		}

		second, err := uc.ScanDevices.Execute(ctx, devices.ScanDevicesCommand{})
		if err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if len(second.Devices) != 0 {
			t.Errorf("the second scan added %d devices", len(second.Devices))
		}

		all := listDevices(t, ctx, uc)
		if len(all) != len(first.Devices) {
			t.Errorf("duplicated: %d, want %d", len(all), len(first.Devices))
		}

		found, err := uc.SearchDevices.Execute(ctx, devices.SearchDevicesCommand{Query: "Mine"})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(found.Devices) != 1 || found.Devices[0].State() != bluetooth.StatePaired {
			t.Error("the scan overwrote the name or state of the known device")
		}
	})

	t.Run("does not scan with the radio off", func(t *testing.T) {
		uc, ctx := newUseCasesWithAdapter(t, false)

		if _, err := uc.ScanDevices.Execute(ctx, devices.ScanDevicesCommand{}); !errors.Is(err, bluetooth.ErrAdapterDisabled) {
			t.Fatalf("error = %v, want ErrAdapterDisabled", err)
		}
		if all := listDevices(t, ctx, uc); len(all) != 0 {
			t.Errorf("recorded %d devices with the radio off", len(all))
		}
	})

	t.Run("respects a cancelled context", func(t *testing.T) {
		uc, _ := newUseCases(t)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if _, err := uc.ScanDevices.Execute(ctx, devices.ScanDevicesCommand{}); !errors.Is(err, context.Canceled) {
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

	uc := devices.NewUseCases(devices.Deps{
		Repo:     repo,
		Adapters: memory.NewAdapterRepository(true),
		Scanner:  &registeringScanner{repo: repo, devices: nearby},
	})

	added, err := uc.ScanDevices.Execute(ctx, devices.ScanDevicesCommand{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(added.Devices) != len(nearby) {
		t.Fatalf("reported %d new devices, want %d", len(added.Devices), len(nearby))
	}

	// The second pass no longer finds anything new.
	again, err := uc.ScanDevices.Execute(ctx, devices.ScanDevicesCommand{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(again.Devices) != 0 {
		t.Errorf("the second scan reported %d new devices", len(again.Devices))
	}
}
