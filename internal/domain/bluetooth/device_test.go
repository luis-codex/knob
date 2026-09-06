package bluetooth_test

import (
	"errors"
	"strings"
	"testing"

	"settings-cli/internal/domain/bluetooth"
	"settings-cli/internal/domain/errs"
)

func mustDevice(t *testing.T) bluetooth.Device {
	t.Helper()
	addr, err := bluetooth.NewAddress("AA:BB:CC:DD:EE:FF")
	if err != nil {
		t.Fatalf("NewAddress: %v", err)
	}
	name, err := bluetooth.NewName("WH-1000XM4")
	if err != nil {
		t.Fatalf("NewName: %v", err)
	}
	device, err := bluetooth.Discover(addr, name, bluetooth.KindHeadphones)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	return device
}

// paired y connected devuelven un dispositivo ya en ese estado.
func paired(t *testing.T) bluetooth.Device {
	t.Helper()
	d, err := mustDevice(t).Pair()
	if err != nil {
		t.Fatalf("Pair: %v", err)
	}
	return d
}

func connected(t *testing.T) bluetooth.Device {
	t.Helper()
	d, err := paired(t).Connect()
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	return d
}

func TestNewAddress(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		err   error
	}{
		{name: "normaliza a mayúsculas", input: "aa:bb:cc:dd:ee:ff", want: "AA:BB:CC:DD:EE:FF"},
		{name: "acepta guiones", input: "AA-BB-CC-DD-EE-FF", want: "AA:BB:CC:DD:EE:FF"},
		{name: "recorta espacios", input: "  AA:BB:CC:DD:EE:FF  ", want: "AA:BB:CC:DD:EE:FF"},
		{name: "vacía", input: "", err: bluetooth.ErrInvalidAddress},
		{name: "pocos octetos", input: "AA:BB:CC:DD:EE", err: bluetooth.ErrInvalidAddress},
		{name: "demasiados octetos", input: "AA:BB:CC:DD:EE:FF:00", err: bluetooth.ErrInvalidAddress},
		{name: "octeto corto", input: "AA:BB:CC:DD:EE:F", err: bluetooth.ErrInvalidAddress},
		{name: "carácter no hexadecimal", input: "AA:BB:CC:DD:EE:GG", err: bluetooth.ErrInvalidAddress},
		{name: "sin separadores", input: "AABBCCDDEEFF", err: bluetooth.ErrInvalidAddress},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := bluetooth.NewAddress(tc.input)

			if tc.err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("error = %v, se esperaba %v", err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if got.String() != tc.want {
				t.Errorf("String() = %q, se esperaba %q", got, tc.want)
			}
		})
	}
}

// La misma dirección escrita de dos formas debe ser el mismo Address.
func TestAddressEqualsTrasNormalizar(t *testing.T) {
	a, _ := bluetooth.NewAddress("aa-bb-cc-dd-ee-ff")
	b, _ := bluetooth.NewAddress("AA:BB:CC:DD:EE:FF")

	if !a.Equals(b) {
		t.Errorf("%q y %q deberían ser iguales", a, b)
	}
}

func TestNewName(t *testing.T) {
	if _, err := bluetooth.NewName("  "); !errors.Is(err, bluetooth.ErrEmptyName) {
		t.Errorf("nombre vacío: %v", err)
	}
	if _, err := bluetooth.NewName(strings.Repeat("a", bluetooth.MaxNameLength+1)); !errors.Is(err, bluetooth.ErrNameTooLong) {
		t.Errorf("nombre largo: %v", err)
	}
	got, err := bluetooth.NewName("  Teclado K380 ")
	if err != nil || got.String() != "Teclado K380" {
		t.Errorf("NewName = %q, %v", got, err)
	}
}

func TestBattery(t *testing.T) {
	t.Run("desconocida es el valor cero", func(t *testing.T) {
		var zero bluetooth.Battery
		if zero.Known() {
			t.Error("el valor cero debe ser desconocido, no 0 %")
		}
		if bluetooth.UnknownBattery().Known() {
			t.Error("UnknownBattery debe ser desconocida")
		}
	})

	t.Run("rango válido", func(t *testing.T) {
		for _, level := range []int{0, 50, 100} {
			b, err := bluetooth.NewBattery(level)
			if err != nil {
				t.Fatalf("NewBattery(%d): %v", level, err)
			}
			if !b.Known() || b.Level() != level {
				t.Errorf("NewBattery(%d) = %d/%v", level, b.Level(), b.Known())
			}
		}
	})

	t.Run("fuera de rango", func(t *testing.T) {
		for _, level := range []int{-1, 101} {
			if _, err := bluetooth.NewBattery(level); !errors.Is(err, bluetooth.ErrInvalidBattery) {
				t.Errorf("NewBattery(%d) = %v", level, err)
			}
		}
	})
}

func TestDiscoverNaceSinEmparejar(t *testing.T) {
	d := mustDevice(t)

	if d.State() != bluetooth.StateDiscovered {
		t.Errorf("estado = %v, se esperaba discovered", d.State())
	}
	if d.Battery().Known() {
		t.Error("un dispositivo recién visto no informa batería")
	}
}

// El recorrido completo del ciclo de vida.
func TestCicloDeVida(t *testing.T) {
	d := mustDevice(t)

	d, err := d.Pair()
	if err != nil || d.State() != bluetooth.StatePaired {
		t.Fatalf("Pair = %v, %v", d.State(), err)
	}

	d, err = d.Connect()
	if err != nil || d.State() != bluetooth.StateConnected {
		t.Fatalf("Connect = %v, %v", d.State(), err)
	}

	d, err = d.Disconnect()
	if err != nil || d.State() != bluetooth.StatePaired {
		t.Fatalf("Disconnect = %v, %v", d.State(), err)
	}

	d, err = d.Unpair()
	if err != nil || d.State() != bluetooth.StateDiscovered {
		t.Fatalf("Unpair = %v, %v", d.State(), err)
	}
}

// Las transiciones imposibles deben ser conflicto, no dato inválido.
func TestTransicionesInvalidas(t *testing.T) {
	tests := []struct {
		name string
		do   func(*testing.T) (bluetooth.Device, error)
		err  error
	}{
		{"conectar sin emparejar", func(t *testing.T) (bluetooth.Device, error) { return mustDevice(t).Connect() }, bluetooth.ErrNotPaired},
		{"desconectar sin conectar", func(t *testing.T) (bluetooth.Device, error) { return paired(t).Disconnect() }, bluetooth.ErrNotConnected},
		{"emparejar dos veces", func(t *testing.T) (bluetooth.Device, error) { return paired(t).Pair() }, bluetooth.ErrAlreadyPaired},
		{"emparejar estando conectado", func(t *testing.T) (bluetooth.Device, error) { return connected(t).Pair() }, bluetooth.ErrAlreadyPaired},
		{"conectar dos veces", func(t *testing.T) (bluetooth.Device, error) { return connected(t).Connect() }, bluetooth.ErrAlreadyConnected},
		{"desemparejar sin emparejar", func(t *testing.T) (bluetooth.Device, error) { return mustDevice(t).Unpair() }, bluetooth.ErrNotPaired},
		{"batería sin conexión", func(t *testing.T) (bluetooth.Device, error) {
			b, _ := bluetooth.NewBattery(50)
			return paired(t).ReportBattery(b)
		}, bluetooth.ErrNotConnected},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.do(t)

			if !errors.Is(err, tc.err) {
				t.Fatalf("error = %v, se esperaba %v", err, tc.err)
			}
			if !errs.IsConflict(err) {
				t.Error("una transición imposible debe ser conflicto, no dato inválido")
			}
			if !got.IsZero() {
				t.Error("una transición fallida no debe devolver dispositivo")
			}
		})
	}
}

// Desemparejar desconecta de paso y olvida la batería.
func TestUnpairDesdeConectado(t *testing.T) {
	b, _ := bluetooth.NewBattery(80)
	d, err := connected(t).ReportBattery(b)
	if err != nil {
		t.Fatalf("ReportBattery: %v", err)
	}

	d, err = d.Unpair()
	if err != nil {
		t.Fatalf("Unpair: %v", err)
	}
	if d.State() != bluetooth.StateDiscovered {
		t.Errorf("estado = %v", d.State())
	}
	if d.Battery().Known() {
		t.Error("al olvidar el dispositivo la batería deja de conocerse")
	}
}

// Las transiciones devuelven copias: el original no cambia.
func TestTransicionesNoMutanElOriginal(t *testing.T) {
	original := mustDevice(t)

	if _, err := original.Pair(); err != nil {
		t.Fatalf("Pair: %v", err)
	}
	if original.State() != bluetooth.StateDiscovered {
		t.Errorf("el original mutó a %v", original.State())
	}
}

func TestParseKind(t *testing.T) {
	for _, kind := range []bluetooth.Kind{
		bluetooth.KindUnknown, bluetooth.KindHeadphones, bluetooth.KindSpeaker,
		bluetooth.KindMouse, bluetooth.KindKeyboard, bluetooth.KindPhone,
	} {
		got, err := bluetooth.ParseKind(kind.String())
		if err != nil {
			t.Fatalf("ParseKind(%q): %v", kind, err)
		}
		if got != kind {
			t.Errorf("ida y vuelta de %v dio %v", kind, got)
		}
	}

	if _, err := bluetooth.ParseKind("tostadora"); !errors.Is(err, bluetooth.ErrUnknownKind) {
		t.Errorf("tipo desconocido: %v", err)
	}
}

func TestRestoreConservaEstado(t *testing.T) {
	addr, _ := bluetooth.NewAddress("11:22:33:44:55:66")
	name, _ := bluetooth.NewName("MX Master")
	battery, _ := bluetooth.NewBattery(45)

	d, err := bluetooth.Restore(addr, name, bluetooth.KindMouse, bluetooth.StateConnected, battery)
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}

	if d.State() != bluetooth.StateConnected || d.Battery().Level() != 45 {
		t.Errorf("Restore devolvió %v / %d%%", d.State(), d.Battery().Level())
	}

	if _, err := bluetooth.Restore(bluetooth.Address{}, name, bluetooth.KindMouse, bluetooth.StatePaired, battery); !errors.Is(err, bluetooth.ErrInvalidAddress) {
		t.Errorf("sin dirección: %v", err)
	}
}
