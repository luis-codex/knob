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

// newService monta el servicio con los repositorios reales: los de memoria son
// la implementación de producción, así que no hace falta un mock. La radio
// arranca encendida salvo que se diga lo contrario.
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
	t.Run("registra sin emparejar", func(t *testing.T) {
		svc, ctx := discovered(t)

		all, err := svc.List(ctx)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(all) != 1 || all[0].State() != bluetooth.StateDiscovered {
			t.Fatalf("List devolvió %d dispositivos en estado %v", len(all), all[0].State())
		}
	})

	t.Run("la misma dirección dos veces es conflicto", func(t *testing.T) {
		svc, ctx := discovered(t)

		// En minúsculas y con guiones: debe reconocerse como la misma.
		_, err := svc.Discover(ctx, "aa-bb-cc-dd-ee-ff", "Otro nombre", "mouse")
		if !errors.Is(err, bluetooth.ErrAlreadyKnown) {
			t.Fatalf("error = %v, se esperaba ErrAlreadyKnown", err)
		}
		if !errs.IsConflict(err) {
			t.Error("debe llegar clasificado como conflicto")
		}

		all, _ := svc.List(ctx)
		if len(all) != 1 {
			t.Errorf("un alta duplicada no debe añadir nada: %d dispositivos", len(all))
		}
		if all[0].Name().String() != "WH-1000XM4" {
			t.Errorf("el alta duplicada pisó el nombre: %q", all[0].Name())
		}
	})

	t.Run("rechaza datos inválidos sin persistir", func(t *testing.T) {
		tests := []struct {
			name            string
			addr, dev, kind string
			err             error
		}{
			{"dirección inválida", "no-es-una-mac", "X", "mouse", bluetooth.ErrInvalidAddress},
			{"nombre vacío", addrA, "   ", "mouse", bluetooth.ErrEmptyName},
			{"tipo desconocido", addrA, "X", "tostadora", bluetooth.ErrUnknownKind},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				svc, ctx := newService(t)

				if _, err := svc.Discover(ctx, tc.addr, tc.dev, tc.kind); !errors.Is(err, tc.err) {
					t.Fatalf("error = %v, se esperaba %v", err, tc.err)
				}
				if all, _ := svc.List(ctx); len(all) != 0 {
					t.Errorf("no debe persistir nada: %d dispositivos", len(all))
				}
			})
		}
	})
}

// El recorrido completo, comprobando que cada paso queda guardado.
func TestCicloDeVidaPersiste(t *testing.T) {
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
				t.Errorf("tras %s el estado guardado es %v, se esperaba %v", step.name, all[0].State(), step.want)
			}
		})
	}
}

// Una transición imposible no debe modificar lo guardado.
func TestTransicionInvalidaNoPersiste(t *testing.T) {
	svc, ctx := discovered(t)

	_, err := svc.Connect(ctx, addrA)
	if !errors.Is(err, bluetooth.ErrNotPaired) {
		t.Fatalf("error = %v, se esperaba ErrNotPaired", err)
	}
	if !errs.IsConflict(err) {
		t.Error("debe llegar clasificado como conflicto")
	}

	all, _ := svc.List(ctx)
	if all[0].State() != bluetooth.StateDiscovered {
		t.Errorf("el estado cambió a %v", all[0].State())
	}
}

func TestDispositivoInexistente(t *testing.T) {
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
				t.Fatalf("error = %v, se esperaba ErrNotFound", err)
			}
		})
	}
}

func TestRename(t *testing.T) {
	svc, ctx := discovered(t)

	if _, err := svc.Rename(ctx, addrA, "  Auriculares del salón "); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	all, _ := svc.List(ctx)
	if all[0].Name().String() != "Auriculares del salón" {
		t.Errorf("nombre = %q", all[0].Name())
	}

	if _, err := svc.Rename(ctx, addrA, ""); !errors.Is(err, bluetooth.ErrEmptyName) {
		t.Fatalf("nombre vacío: %v", err)
	}
	all, _ = svc.List(ctx)
	if all[0].Name().String() != "Auriculares del salón" {
		t.Errorf("un rename inválido cambió el nombre: %q", all[0].Name())
	}
}

func TestReportBattery(t *testing.T) {
	svc, ctx := discovered(t)

	// Sin conectar, el dispositivo no informa de nada.
	if _, err := svc.ReportBattery(ctx, addrA, 80); !errors.Is(err, bluetooth.ErrNotConnected) {
		t.Fatalf("sin conectar: %v", err)
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
		t.Errorf("batería = %d/%v", all[0].Battery().Level(), all[0].Battery().Known())
	}

	if _, err := svc.ReportBattery(ctx, addrA, 101); !errors.Is(err, bluetooth.ErrInvalidBattery) {
		t.Errorf("nivel fuera de rango: %v", err)
	}
}

func TestRemove(t *testing.T) {
	svc, ctx := discovered(t)

	if err := svc.Remove(ctx, addrA); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if all, _ := svc.List(ctx); len(all) != 0 {
		t.Errorf("quedan %d dispositivos", len(all))
	}
	if err := svc.Remove(ctx, addrA); !errors.Is(err, bluetooth.ErrNotFound) {
		t.Errorf("borrar dos veces: %v", err)
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
		{"vacía devuelve todo", "", 2},
		{"solo espacios devuelve todo", "  ", 2},
		{"subcadena", "master", 1},
		{"ignora mayúsculas", "wh-1000", 1},
		{"sin coincidencias", "zzz", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := svc.Search(ctx, tc.query)
			if err != nil {
				t.Fatalf("Search: %v", err)
			}
			if len(got) != tc.want {
				t.Errorf("Search(%q) devolvió %d, se esperaba %d", tc.query, len(got), tc.want)
			}
		})
	}
}

func TestAdapter(t *testing.T) {
	t.Run("apagar desconecta lo conectado", func(t *testing.T) {
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
			t.Error("el adaptador sigue encendido")
		}

		all, _ := svc.List(ctx)
		if all[0].State() != bluetooth.StatePaired {
			t.Errorf("estado = %v, se esperaba paired tras apagar", all[0].State())
		}
	})

	t.Run("encender no reconecta", func(t *testing.T) {
		svc, ctx := discovered(t)
		svc.Pair(ctx, addrA)
		svc.Connect(ctx, addrA)
		svc.DisableAdapter(ctx)

		if _, err := svc.EnableAdapter(ctx); err != nil {
			t.Fatalf("EnableAdapter: %v", err)
		}

		all, _ := svc.List(ctx)
		if all[0].State() != bluetooth.StatePaired {
			t.Errorf("estado = %v: encender no debe reconectar", all[0].State())
		}
	})

	t.Run("es idempotente", func(t *testing.T) {
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

// Con la radio apagada no se puede emparejar ni conectar, pero la gestión de
// dispositivos ya conocidos sigue funcionando.
func TestAdapterApagadoBloqueaConexiones(t *testing.T) {
	blocked := map[string]func(*devices.Service, context.Context) error{
		"pair":    func(s *devices.Service, ctx context.Context) error { _, err := s.Pair(ctx, addrA); return err },
		"connect": func(s *devices.Service, ctx context.Context) error { _, err := s.Connect(ctx, addrA); return err },
	}

	for name, op := range blocked {
		t.Run(name+" bloqueado", func(t *testing.T) {
			svc, ctx := newServiceWithAdapter(t, false)
			if _, err := svc.Discover(ctx, addrA, "X", "mouse"); err != nil {
				t.Fatalf("Discover: %v", err)
			}

			err := op(svc, ctx)
			if !errors.Is(err, bluetooth.ErrAdapterDisabled) {
				t.Fatalf("error = %v, se esperaba ErrAdapterDisabled", err)
			}
			if !errs.IsConflict(err) {
				t.Error("debe llegar clasificado como conflicto")
			}

			all, _ := svc.List(ctx)
			if all[0].State() != bluetooth.StateDiscovered {
				t.Errorf("el estado cambió a %v", all[0].State())
			}
		})
	}

	t.Run("la gestión sigue permitida", func(t *testing.T) {
		svc, ctx := newServiceWithAdapter(t, false)
		if _, err := svc.Discover(ctx, addrA, "X", "mouse"); err != nil {
			t.Fatalf("Discover: %v", err)
		}

		if _, err := svc.Rename(ctx, addrA, "Renombrado"); err != nil {
			t.Errorf("Rename con la radio apagada: %v", err)
		}
		if err := svc.Remove(ctx, addrA); err != nil {
			t.Errorf("Remove con la radio apagada: %v", err)
		}
	})
}

func TestScan(t *testing.T) {
	t.Run("registra los desconocidos", func(t *testing.T) {
		svc, ctx := newService(t)

		added, err := svc.Scan(ctx)
		if err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if len(added) == 0 {
			t.Fatal("la primera búsqueda no encontró nada")
		}

		all, _ := svc.List(ctx)
		if len(all) != len(added) {
			t.Errorf("guardó %d de %d encontrados", len(all), len(added))
		}
		for _, d := range all {
			if d.State() != bluetooth.StateDiscovered {
				t.Errorf("%s llegó en estado %v", d.Name(), d.State())
			}
		}
	})

	t.Run("es idempotente: no duplica ni pisa lo conocido", func(t *testing.T) {
		svc, ctx := newService(t)

		first, err := svc.Scan(ctx)
		if err != nil {
			t.Fatalf("Scan: %v", err)
		}

		// Se empareja uno y se renombra otro antes de volver a buscar.
		target := first[0].Address().String()
		if _, err := svc.Pair(ctx, target); err != nil {
			t.Fatalf("Pair: %v", err)
		}
		if _, err := svc.Rename(ctx, target, "Mío"); err != nil {
			t.Fatalf("Rename: %v", err)
		}

		second, err := svc.Scan(ctx)
		if err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if len(second) != 0 {
			t.Errorf("la segunda búsqueda añadió %d dispositivos", len(second))
		}

		all, _ := svc.List(ctx)
		if len(all) != len(first) {
			t.Errorf("se duplicaron: %d, se esperaban %d", len(all), len(first))
		}

		found, err := svc.Search(ctx, "Mío")
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(found) != 1 || found[0].State() != bluetooth.StatePaired {
			t.Error("la búsqueda pisó el nombre o el estado del dispositivo conocido")
		}
	})

	t.Run("con la radio apagada no busca", func(t *testing.T) {
		svc, ctx := newServiceWithAdapter(t, false)

		if _, err := svc.Scan(ctx); !errors.Is(err, bluetooth.ErrAdapterDisabled) {
			t.Fatalf("error = %v, se esperaba ErrAdapterDisabled", err)
		}
		if all, _ := svc.List(ctx); len(all) != 0 {
			t.Errorf("registró %d dispositivos con la radio apagada", len(all))
		}
	})

	t.Run("respeta el contexto cancelado", func(t *testing.T) {
		svc, _ := newService(t)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if _, err := svc.Scan(ctx); !errors.Is(err, context.Canceled) {
			t.Errorf("error = %v, se esperaba context.Canceled", err)
		}
	})
}

// registeringScanner imita a BlueZ: al escanear, el propio repositorio pasa a
// conocer lo encontrado. Es el caso que rompía el recuento de novedades.
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

func TestScanCuentaNovedadesConBackendQueSeAutorregistra(t *testing.T) {
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
		t.Fatalf("informó de %d novedades, se esperaban %d", len(added), len(nearby))
	}

	// La segunda pasada ya no encuentra nada nuevo.
	again, err := svc.Scan(ctx)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(again) != 0 {
		t.Errorf("la segunda búsqueda informó de %d novedades", len(again))
	}
}
