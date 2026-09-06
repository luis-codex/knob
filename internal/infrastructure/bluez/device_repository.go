package bluez

import (
	"bufio"
	"context"
	"strings"

	"settings-cli/internal/domain/bluetooth"
	"settings-cli/internal/domain/errs"
)

// Repository expone los dispositivos que conoce BlueZ.
type Repository struct{}

var _ bluetooth.Repository = (*Repository)(nil)

func NewRepository() *Repository { return &Repository{} }

// List enumera los dispositivos conocidos y consulta el detalle de cada uno.
// Son N+1 invocaciones, aceptable para una lista de dispositivos Bluetooth.
func (r *Repository) List(ctx context.Context) ([]bluetooth.Device, error) {
	out, err := run(ctx, "devices")
	if err != nil {
		return nil, err
	}

	var devices []bluetooth.Device
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		address, ok := parseDeviceLine(scanner.Text())
		if !ok {
			continue
		}

		device, err := r.FindByAddress(ctx, address)
		if err != nil {
			// Un dispositivo que desaparece entre las dos llamadas no debe
			// tumbar el listado entero.
			continue
		}
		devices = append(devices, device)
	}
	return devices, nil
}

// parseDeviceLine lee "Device AA:BB:CC:DD:EE:FF Nombre".
func parseDeviceLine(line string) (bluetooth.Address, bool) {
	parts := strings.Fields(strings.TrimSpace(line))
	if len(parts) < 2 || parts[0] != "Device" {
		return bluetooth.Address{}, false
	}

	address, err := bluetooth.NewAddress(parts[1])
	if err != nil {
		return bluetooth.Address{}, false
	}
	return address, true
}

func (r *Repository) FindByAddress(ctx context.Context, address bluetooth.Address) (bluetooth.Device, error) {
	out, err := run(ctx, "info", address.String())
	if err != nil {
		return bluetooth.Device{}, err
	}
	return deviceFromInfo(address, out)
}

// deviceFromInfo traduce la salida de `bluetoothctl info` al dominio. Está
// separado de la invocación para poder probarlo sin hardware.
func deviceFromInfo(address bluetooth.Address, out string) (bluetooth.Device, error) {
	if strings.Contains(out, "not available") {
		return bluetooth.Device{}, bluetooth.ErrNotFound
	}

	fields := parseFields(out)

	// Alias es el nombre que ve el usuario y puede cambiar; Name es el que
	// anuncia el dispositivo. Se prefiere Alias, con Name de reserva y la
	// dirección como último recurso: un dispositivo sin nombre no debe
	// tumbar el listado.
	name, err := bluetooth.NewName(firstNonEmpty(fields["Alias"], fields["Name"], address.String()))
	if err != nil {
		return bluetooth.Device{}, err
	}

	return bluetooth.Restore(
		address,
		name,
		parseKind(fields["Icon"]),
		parseState(fields),
		parseBattery(fields["Battery Percentage"]),
	)
}

func parseState(fields map[string]string) bluetooth.State {
	switch {
	case yes(fields, "Connected"):
		return bluetooth.StateConnected
	case yes(fields, "Paired"):
		return bluetooth.StatePaired
	default:
		return bluetooth.StateDiscovered
	}
}

// Save reconcilia: compara el estado deseado con el real y emite las órdenes
// que hagan falta.
//
// Es la diferencia de fondo con un almacén. BlueZ no guarda lo que le pases;
// hay que *hacer* que la realidad coincida.
func (r *Repository) Save(ctx context.Context, d bluetooth.Device) error {
	current, err := r.FindByAddress(ctx, d.Address())
	if err != nil {
		return err
	}
	if current.State() == d.State() {
		return r.saveAlias(ctx, d, current)
	}

	address := d.Address().String()

	switch d.State() {
	case bluetooth.StateConnected:
		if !current.State().IsPaired() {
			if _, err := run(ctx, "pair", address); err != nil {
				return err
			}
		}
		_, err = run(ctx, "connect", address)

	case bluetooth.StatePaired:
		if current.State() == bluetooth.StateConnected {
			_, err = run(ctx, "disconnect", address)
			break
		}
		_, err = run(ctx, "pair", address)

	case bluetooth.StateDiscovered:
		// Olvidar: BlueZ desconecta por su cuenta al quitar el emparejamiento.
		_, err = run(ctx, "remove", address)
	}

	if err != nil {
		return err
	}
	return r.saveAlias(ctx, d, current)
}

// saveAlias no está soportado por bluetoothctl, que no expone la propiedad
// Alias. Cambiar el nombre exigiría hablar con D-Bus directamente.
func (r *Repository) saveAlias(_ context.Context, d, current bluetooth.Device) error {
	if d.Name().String() == current.Name().String() {
		return nil
	}
	return errs.Conflict("renombrar dispositivos no está soportado con BlueZ")
}

func (r *Repository) Delete(ctx context.Context, address bluetooth.Address) error {
	out, err := run(ctx, "remove", address.String())
	if err != nil {
		return err
	}
	if strings.Contains(out, "not available") {
		return bluetooth.ErrNotFound
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
