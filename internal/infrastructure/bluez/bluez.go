// Package bluez implementa los puertos de Bluetooth contra el demonio BlueZ
// del sistema, invocando bluetoothctl.
//
// Se usa la CLI y no D-Bus directamente para no arrastrar dependencias: a
// cambio hay que parsear texto, que es lo que hace parseFields.
package bluez

import (
	"bufio"
	"context"
	"os/exec"
	"strconv"
	"strings"

	"settings-cli/internal/domain/bluetooth"
	"settings-cli/internal/domain/errs"
)

// binary es el ejecutable que se invoca. Variable para poder sustituirlo en
// pruebas.
var binary = "bluetoothctl"

// run ejecuta bluetoothctl y devuelve su salida.
//
// bluetoothctl escribe los fallos en stdout y devuelve 0 en varios casos, así
// que el código de salida no basta: quien llama debe mirar el contenido.
func run(ctx context.Context, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, binary, args...).CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return string(out), errs.Wrap(errs.KindConflict, "no se pudo hablar con BlueZ", err)
	}
	return string(out), nil
}

// parseFields extrae las líneas "Clave: valor" de la salida. Las claves se
// repiten (UUID), así que se queda con la primera, que es la que interesa.
func parseFields(out string) map[string]string {
	fields := make(map[string]string)

	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		key, value, ok := strings.Cut(strings.TrimSpace(scanner.Text()), ": ")
		if !ok {
			continue
		}
		if _, seen := fields[key]; !seen {
			fields[key] = strings.TrimSpace(value)
		}
	}
	return fields
}

func yes(fields map[string]string, key string) bool {
	return fields[key] == "yes"
}

// parseBattery lee "Battery Percentage: 0x52 (82)": el valor útil es el
// decimal entre paréntesis.
func parseBattery(raw string) bluetooth.Battery {
	open := strings.Index(raw, "(")
	close := strings.Index(raw, ")")
	if open < 0 || close < open {
		return bluetooth.UnknownBattery()
	}

	level, err := strconv.Atoi(raw[open+1 : close])
	if err != nil {
		return bluetooth.UnknownBattery()
	}

	battery, err := bluetooth.NewBattery(level)
	if err != nil {
		return bluetooth.UnknownBattery()
	}
	return battery
}

// icons traduce el icono de BlueZ al tipo del dominio.
var icons = map[string]bluetooth.Kind{
	"audio-headset":    bluetooth.KindHeadphones,
	"audio-headphones": bluetooth.KindHeadphones,
	"audio-card":       bluetooth.KindSpeaker,
	"input-mouse":      bluetooth.KindMouse,
	"input-keyboard":   bluetooth.KindKeyboard,
	"input-tablet":     bluetooth.KindKeyboard,
	"phone":            bluetooth.KindPhone,
}

func parseKind(icon string) bluetooth.Kind {
	if kind, ok := icons[icon]; ok {
		return kind
	}
	return bluetooth.KindUnknown
}
