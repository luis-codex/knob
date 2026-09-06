// Package config lee la configuración del usuario y la traduce al vocabulario
// de la interfaz.
//
// No vive bajo infrastructure porque no implementa ningún puerto del dominio:
// no traduce hacia dentro, sino hacia la presentación. Es apoyo del
// composition root, que es su único cliente.
//
// El paquete styles se queda como tokens puros y recibe los colores ya
// resueltos: leer el disco es cosa de aquí.
package config

import (
	"errors"
	"fmt"
	"image/color"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"charm.land/lipgloss/v2"
	"github.com/BurntSushi/toml"

	"settings-cli/internal/shared/styles"
)

const (
	// appDir es la carpeta bajo el directorio de configuración del usuario.
	appDir = "settings-cli"
	// themeFile es el fichero de tema dentro de esa carpeta.
	themeFile = "theme.toml"
)

// hexPattern acepta #RGB y #RRGGBB, que es lo que entiende lipgloss.
var hexPattern = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// palette son los colores tal y como vienen del fichero, sin validar.
type palette struct {
	Text         string `toml:"text"`
	Muted        string `toml:"muted"`
	Faint        string `toml:"faint"`
	Line         string `toml:"line"`
	LineStrong   string `toml:"line_strong"`
	Accent       string `toml:"accent"`
	AccentStrong string `toml:"accent_strong"`
	OnAccent     string `toml:"on_accent"`
	Bg           string `toml:"bg"`
	BgSelected   string `toml:"bg_selected"`
	Success      string `toml:"success"`
	Warning      string `toml:"warning"`
	Danger       string `toml:"danger"`
}

type file struct {
	Light palette `toml:"light"`
	Dark  palette `toml:"dark"`
}

// ThemePath es dónde se busca el fichero. Respeta XDG_CONFIG_HOME.
func ThemePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appDir, themeFile), nil
}

// LoadTheme lee el tema del usuario.
//
// Devuelve siempre una paleta utilizable: los errores acompañan al resultado en
// vez de sustituirlo. Un tema mal escrito no puede dejar a nadie sin abrir sus
// ajustes, así que lo que falle se queda en el valor por defecto y se avisa.
func LoadTheme() (styles.Custom, []error) {
	path, err := ThemePath()
	if err != nil {
		return styles.Custom{}, []error{fmt.Errorf("no se pudo localizar la configuración: %w", err)}
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		// No tener tema es lo normal, no un fallo.
		if errors.Is(err, fs.ErrNotExist) {
			return styles.Custom{}, nil
		}
		return styles.Custom{}, []error{fmt.Errorf("no se pudo leer %s: %w", path, err)}
	}

	var parsed file
	if err := toml.Unmarshal(raw, &parsed); err != nil {
		return styles.Custom{}, []error{fmt.Errorf("%s no es TOML válido: %w", path, err)}
	}

	light, lightErrs := toPalette("light", parsed.Light)
	dark, darkErrs := toPalette("dark", parsed.Dark)

	return styles.Custom{Light: light, Dark: dark}, append(lightErrs, darkErrs...)
}

// toPalette convierte los hex a colores. Un color inválido no invalida los
// demás: se descarta ese, se avisa y se sigue con el resto.
func toPalette(section string, p palette) (styles.Palette, []error) {
	var errs []error

	parse := func(key, value string) color.Color {
		if value == "" {
			return nil // ausente: se queda el de por defecto
		}
		if !hexPattern.MatchString(value) {
			errs = append(errs, fmt.Errorf("[%s] %s: %q no es un color hexadecimal", section, key, value))
			return nil
		}
		return lipgloss.Color(value)
	}

	// Tabla en vez de trece condiciones: añadir un color al tema es añadir una
	// línea aquí y otra en la struct del fichero.
	var out styles.Palette
	fields := []struct {
		key   string
		value string
		dst   *color.Color
	}{
		{"text", p.Text, &out.Text},
		{"muted", p.Muted, &out.Muted},
		{"faint", p.Faint, &out.Faint},
		{"line", p.Line, &out.Line},
		{"line_strong", p.LineStrong, &out.LineStrong},
		{"accent", p.Accent, &out.Accent},
		{"accent_strong", p.AccentStrong, &out.AccentStrong},
		{"on_accent", p.OnAccent, &out.OnAccent},
		{"bg", p.Bg, &out.Bg},
		{"bg_selected", p.BgSelected, &out.BgSelected},
		{"success", p.Success, &out.Success},
		{"warning", p.Warning, &out.Warning},
		{"danger", p.Danger, &out.Danger},
	}

	for _, f := range fields {
		if c := parse(f.key, f.value); c != nil {
			*f.dst = c
		}
	}

	return out, errs
}

// exampleTheme es la plantilla que se escribe con WriteExampleTheme. Lleva
// todos los colores comentados con su valor por defecto: así se ve qué se
// puede tocar sin tener que leer el código.
const exampleTheme = `# Tema de settings-cli.
#
# Descomenta solo lo que quieras cambiar: lo que falte se queda en el valor
# por defecto. Los colores son hexadecimales, #RGB o #RRGGBB.

[dark]
# text          = "#EAEAE8"   # texto principal
# muted         = "#A8A8A5"   # texto secundario
# faint         = "#80807E"   # etiquetas y metadatos
# line          = "#28282B"   # separadores
# line_strong   = "#3E3E42"   # borde del modal
accent          = "#516BEB"   # teclas, selección, botón primario
# accent_strong = "#7D91F2"   # acento enfatizado
# on_accent     = "#0D0A0C"   # texto sobre el acento
# bg            = "#0D0A0C"   # fondo del modal
# bg_selected   = "#111634"   # fondo de la fila seleccionada
# success       = "#28DC82"
# warning       = "#FBBF24"
# danger        = "#FF5C5C"

[light]
# text          = "#1A1A1A"
# muted         = "#6B6B68"
# faint         = "#94948F"
# line          = "#E2E2E0"
# line_strong   = "#C9C9C6"
# accent        = "#3B4FC4"
# accent_strong = "#516BEB"
# on_accent     = "#FFFFFF"
# bg            = "#FAFAF8"
# bg_selected   = "#E7EBFD"
# success       = "#12864F"
# warning       = "#B45309"
# danger        = "#C0392B"
`

// WriteExampleTheme deja la plantilla en su sitio y devuelve la ruta.
//
// No pisa un tema existente: sobrescribir lo que el usuario ya haya ajustado
// sería peor que no hacer nada.
func WriteExampleTheme() (string, error) {
	path, err := ThemePath()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(path); err == nil {
		return path, fs.ErrExist
	} else if !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(exampleTheme), 0o644); err != nil {
		return "", err
	}
	return path, nil
}
