// Package pulse implementa el puerto de audio contra el servidor de sonido,
// invocando `pactl -f json`. Sirve tanto para PulseAudio como para PipeWire,
// que expone el mismo protocolo.
//
// Se usa la salida JSON y no la de texto: trae los booleanos ya tipados, los
// nombres de los predeterminados directos y evita parsear a mano un formato
// que cambia entre versiones.
package pulse

import (
	"context"
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"

	"settings-cli/internal/domain/audio"
	"settings-cli/internal/domain/errs"
)

// binary es el ejecutable que se invoca. Variable para poder sustituirlo en
// pruebas.
var binary = "pactl"

func run(ctx context.Context, args ...string) ([]byte, error) {
	out, err := exec.CommandContext(ctx, binary, append([]string{"-f", "json"}, args...)...).Output()
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errs.Wrap(errs.KindConflict, "no se pudo hablar con el servidor de sonido", err)
	}
	return out, nil
}

// kindOf traduce la dirección al vocabulario de pactl.
func kindOf(direction audio.Direction) string {
	if direction == audio.Input {
		return "source"
	}
	return "sink"
}

// jsonChannel es el volumen de un canal. Solo interesa el porcentaje: el valor
// crudo y los decibelios son detalle del servidor.
type jsonChannel struct {
	ValuePercent string `json:"value_percent"`
}

type jsonDevice struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Mute        bool                   `json:"mute"`
	Volume      map[string]jsonChannel `json:"volume"`
	// MonitorSource, en una fuente, es la salida de la que es copia. Vacío
	// significa que es una entrada real.
	MonitorSource string `json:"monitor_source"`
}

type jsonInfo struct {
	DefaultSink   string `json:"default_sink_name"`
	DefaultSource string `json:"default_source_name"`
}

// parsePercent lee "45%".
func parsePercent(s string) (int, bool) {
	digits, ok := strings.CutSuffix(strings.TrimSpace(s), "%")
	if !ok {
		return 0, false
	}

	value, err := strconv.Atoi(digits)
	if err != nil {
		return 0, false
	}
	return value, true
}

// volumeOf resume el volumen por canales en un solo nivel.
//
// Se toma el máximo y no "el primero": el JSON los trae en un mapa, y el
// recorrido de un mapa en Go es aleatorio, así que quedarse con uno
// cualquiera daría un nivel distinto en cada lectura de un dispositivo
// desbalanceado. El máximo es además lo que enseñan los mezcladores.
func volumeOf(channels map[string]jsonChannel) (audio.Volume, bool) {
	level, found := 0, false

	for _, channel := range channels {
		value, ok := parsePercent(channel.ValuePercent)
		if !ok {
			continue
		}
		if !found || value > level {
			level, found = value, true
		}
	}

	if !found {
		return audio.Volume{}, false
	}
	return audio.ClampVolume(level), true
}

// toDevice traduce un dispositivo de pactl al dominio. Devuelve false si el
// bloque no describe un dispositivo utilizable.
func toDevice(d jsonDevice, direction audio.Direction, defaultName string) (audio.Device, bool) {
	// Los monitores son la copia de una salida, no una entrada real: sacarlos
	// aquí evita que un micrófono inexistente aparezca en la lista.
	if direction == audio.Input && d.MonitorSource != "" {
		return audio.Device{}, false
	}

	id, err := audio.NewID(d.Name)
	if err != nil {
		return audio.Device{}, false
	}

	name, err := audio.NewName(firstNonEmpty(d.Description, d.Name))
	if err != nil {
		return audio.Device{}, false
	}

	volume, ok := volumeOf(d.Volume)
	if !ok {
		return audio.Device{}, false
	}

	device, err := audio.Restore(id, name, direction, volume, d.Mute, d.Name == defaultName)
	if err != nil {
		return audio.Device{}, false
	}
	return device, true
}

func decodeDevices(raw []byte) ([]jsonDevice, error) {
	var devices []jsonDevice
	if err := json.Unmarshal(raw, &devices); err != nil {
		return nil, errs.Wrap(errs.KindConflict, "respuesta ilegible del servidor de sonido", err)
	}
	return devices, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
