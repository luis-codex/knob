// Package devices contiene los casos de uso de dispositivos Bluetooth.
//
// Es la frontera del dominio: recibe primitivas de la interfaz, las convierte
// en objetos de valor y orquesta el repositorio.
//
// Se llama devices (plural) y no bluetooth para no chocar con el paquete de
// dominio.
package devices

import (
	"context"
	"errors"
	"strings"

	"settings-cli/internal/domain/bluetooth"
)

// transition es una operación del agregado que devuelve el dispositivo
// resultante. Coincide con la firma de Pair, Connect y compañía, que se pasan
// como expresiones de método.
type transition func(bluetooth.Device) (bluetooth.Device, error)

// Service agrupa los casos de uso sobre dispositivos y adaptador.
//
// La regla de que no se puede emparejar ni conectar con la radio apagada vive
// aquí y no en el dominio: cruza dos agregados, y ninguno de los dos puede
// conocer al otro sin acoplarlos.
type Service struct {
	repo     bluetooth.Repository
	adapters bluetooth.AdapterRepository
	scanner  bluetooth.Scanner
}

// NewService recibe los puertos, no implementaciones concretas.
func NewService(repo bluetooth.Repository, adapters bluetooth.AdapterRepository, scanner bluetooth.Scanner) *Service {
	return &Service{repo: repo, adapters: adapters, scanner: scanner}
}

// Scan busca dispositivos cercanos y registra los que no se conocían.
// Devuelve solo los nuevos: los ya conocidos se dejan intactos para no pisar
// su estado ni su nombre.
func (s *Service) Scan(ctx context.Context) ([]bluetooth.Device, error) {
	if err := s.requireAdapter(ctx); err != nil {
		return nil, err
	}

	// Lo ya conocido se mira ANTES de buscar. Con un backend real el propio
	// escaneo hace que el repositorio pase a conocer lo que encuentra, así
	// que preguntarle después daría siempre "ya conocido" y nunca habría
	// novedades.
	known, err := s.knownAddresses(ctx)
	if err != nil {
		return nil, err
	}

	found, err := s.scanner.Scan(ctx)
	if err != nil {
		return nil, err
	}

	var added []bluetooth.Device
	for _, device := range found {
		if _, ok := known[device.Address().String()]; ok {
			continue
		}

		if err := s.repo.Save(ctx, device); err != nil {
			return nil, err
		}
		added = append(added, device)
	}
	return added, nil
}

func (s *Service) knownAddresses(ctx context.Context) (map[string]struct{}, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	known := make(map[string]struct{}, len(all))
	for _, d := range all {
		known[d.Address().String()] = struct{}{}
	}
	return known, nil
}

// isKnown distingue "no existe" de "el repositorio falló".
func (s *Service) isKnown(ctx context.Context, address bluetooth.Address) (bool, error) {
	switch _, err := s.repo.FindByAddress(ctx, address); {
	case err == nil:
		return true, nil
	case errors.Is(err, bluetooth.ErrNotFound):
		return false, nil
	default:
		return false, err
	}
}

// Adapter devuelve el estado de la radio.
func (s *Service) Adapter(ctx context.Context) (bluetooth.Adapter, error) {
	return s.adapters.Get(ctx)
}

// EnableAdapter enciende la radio. No reconecta nada: reconectar es una
// decisión del usuario, no un efecto de encender.
func (s *Service) EnableAdapter(ctx context.Context) (bluetooth.Adapter, error) {
	adapter, err := s.adapters.Get(ctx)
	if err != nil {
		return bluetooth.Adapter{}, err
	}

	next := adapter.Enable()
	if err := s.adapters.Save(ctx, next); err != nil {
		return bluetooth.Adapter{}, err
	}
	return next, nil
}

// DisableAdapter apaga la radio y corta las conexiones activas.
//
// Los dispositivos se desconectan antes de guardar el adaptador: si algo
// fallara a mitad, el estado guardado seguiría diciendo "encendido", que es
// más fiel que lo contrario.
func (s *Service) DisableAdapter(ctx context.Context) (bluetooth.Adapter, error) {
	adapter, err := s.adapters.Get(ctx)
	if err != nil {
		return bluetooth.Adapter{}, err
	}

	if err := s.disconnectAll(ctx); err != nil {
		return bluetooth.Adapter{}, err
	}

	next := adapter.Disable()
	if err := s.adapters.Save(ctx, next); err != nil {
		return bluetooth.Adapter{}, err
	}
	return next, nil
}

func (s *Service) disconnectAll(ctx context.Context) error {
	all, err := s.repo.List(ctx)
	if err != nil {
		return err
	}

	for _, d := range all {
		if d.State() != bluetooth.StateConnected {
			continue
		}

		next, err := d.Disconnect()
		if err != nil {
			return err
		}
		if err := s.repo.Save(ctx, next); err != nil {
			return err
		}
	}
	return nil
}

// requireAdapter falla si la radio está apagada.
func (s *Service) requireAdapter(ctx context.Context) error {
	adapter, err := s.adapters.Get(ctx)
	if err != nil {
		return err
	}
	if !adapter.Enabled() {
		return bluetooth.ErrAdapterDisabled
	}
	return nil
}

// apply carga el dispositivo, le aplica la transición y guarda el resultado.
// Todas las operaciones de estado siguen este mismo camino.
func (s *Service) apply(ctx context.Context, address string, change transition) (bluetooth.Device, error) {
	addr, err := bluetooth.NewAddress(address)
	if err != nil {
		return bluetooth.Device{}, err
	}

	current, err := s.repo.FindByAddress(ctx, addr)
	if err != nil {
		return bluetooth.Device{}, err
	}

	next, err := change(current)
	if err != nil {
		return bluetooth.Device{}, err
	}

	if err := s.repo.Save(ctx, next); err != nil {
		return bluetooth.Device{}, err
	}
	return next, nil
}

// Discover registra un dispositivo nuevo, sin emparejar.
func (s *Service) Discover(ctx context.Context, address, name, kind string) (bluetooth.Device, error) {
	addr, err := bluetooth.NewAddress(address)
	if err != nil {
		return bluetooth.Device{}, err
	}

	deviceName, err := bluetooth.NewName(name)
	if err != nil {
		return bluetooth.Device{}, err
	}

	deviceKind, err := bluetooth.ParseKind(kind)
	if err != nil {
		return bluetooth.Device{}, err
	}

	// Registrar dos veces la misma dirección es un conflicto, no un alta: el
	// alta silenciosa perdería el estado del dispositivo ya conocido.
	known, err := s.isKnown(ctx, addr)
	if err != nil {
		return bluetooth.Device{}, err
	}
	if known {
		return bluetooth.Device{}, bluetooth.ErrAlreadyKnown
	}

	device, err := bluetooth.Discover(addr, deviceName, deviceKind)
	if err != nil {
		return bluetooth.Device{}, err
	}

	if err := s.repo.Save(ctx, device); err != nil {
		return bluetooth.Device{}, err
	}
	return device, nil
}

// Pair y Connect exigen la radio encendida. Las demás operaciones son de
// gestión y funcionan igual con el Bluetooth apagado, como en cualquier panel
// de ajustes.
func (s *Service) Pair(ctx context.Context, address string) (bluetooth.Device, error) {
	if err := s.requireAdapter(ctx); err != nil {
		return bluetooth.Device{}, err
	}
	return s.apply(ctx, address, bluetooth.Device.Pair)
}

func (s *Service) Unpair(ctx context.Context, address string) (bluetooth.Device, error) {
	return s.apply(ctx, address, bluetooth.Device.Unpair)
}

func (s *Service) Connect(ctx context.Context, address string) (bluetooth.Device, error) {
	if err := s.requireAdapter(ctx); err != nil {
		return bluetooth.Device{}, err
	}
	return s.apply(ctx, address, bluetooth.Device.Connect)
}

func (s *Service) Disconnect(ctx context.Context, address string) (bluetooth.Device, error) {
	return s.apply(ctx, address, bluetooth.Device.Disconnect)
}

// Rename cambia el nombre visible del dispositivo.
func (s *Service) Rename(ctx context.Context, address, name string) (bluetooth.Device, error) {
	newName, err := bluetooth.NewName(name)
	if err != nil {
		return bluetooth.Device{}, err
	}

	return s.apply(ctx, address, func(d bluetooth.Device) (bluetooth.Device, error) {
		return d.Rename(newName)
	})
}

// ReportBattery registra el nivel de carga informado por el dispositivo.
func (s *Service) ReportBattery(ctx context.Context, address string, level int) (bluetooth.Device, error) {
	battery, err := bluetooth.NewBattery(level)
	if err != nil {
		return bluetooth.Device{}, err
	}

	return s.apply(ctx, address, func(d bluetooth.Device) (bluetooth.Device, error) {
		return d.ReportBattery(battery)
	})
}

// Remove olvida el dispositivo por completo. A diferencia de Unpair, deja de
// estar en la lista.
func (s *Service) Remove(ctx context.Context, address string) error {
	addr, err := bluetooth.NewAddress(address)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, addr)
}

// List devuelve todos los dispositivos conocidos.
func (s *Service) List(ctx context.Context) ([]bluetooth.Device, error) {
	return s.repo.List(ctx)
}

// Search filtra por nombre, sin distinguir mayúsculas. Con query vacía
// equivale a List.
func (s *Service) Search(ctx context.Context, query string) ([]bluetooth.Device, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	needle := strings.ToLower(strings.TrimSpace(query))
	if needle == "" {
		return all, nil
	}

	out := make([]bluetooth.Device, 0, len(all))
	for _, d := range all {
		if strings.Contains(strings.ToLower(d.Name().String()), needle) {
			out = append(out, d)
		}
	}
	return out, nil
}
