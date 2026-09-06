package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// writeTheme deja un theme.toml en un directorio de configuración temporal.
func writeTheme(t *testing.T, content string) {
	t.Helper()

	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	path := filepath.Join(dir, appDir)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(path, themeFile), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func TestSinFicheroNoEsError(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	custom, errs := LoadTheme()
	if len(errs) != 0 {
		t.Fatalf("no tener tema no debe dar errores: %v", errs)
	}
	if custom.Dark.Accent != nil || custom.Light.Accent != nil {
		t.Error("sin fichero no debe haber colores personalizados")
	}
}

func TestSobrescrituraParcial(t *testing.T) {
	writeTheme(t, `
# Solo el acento; el resto se queda por defecto.
[dark]
accent = "#516BEB"
`)

	custom, errs := LoadTheme()
	if len(errs) != 0 {
		t.Fatalf("errores inesperados: %v", errs)
	}

	if custom.Dark.Accent != lipgloss.Color("#516BEB") {
		t.Errorf("accent = %v", custom.Dark.Accent)
	}
	if custom.Dark.Text != nil {
		t.Error("un color no declarado debe quedar nulo para que Merge use el de por defecto")
	}
	if custom.Light.Accent != nil {
		t.Error("declarar [dark] no debe tocar [light]")
	}
}

func TestVariantesIndependientes(t *testing.T) {
	writeTheme(t, `
[light]
accent = "#3B4FC4"

[dark]
accent = "#516BEB"
`)

	custom, errs := LoadTheme()
	if len(errs) != 0 {
		t.Fatalf("errores inesperados: %v", errs)
	}
	if custom.For(true) != custom.Dark || custom.For(false) != custom.Light {
		t.Error("For no elige la variante correcta")
	}
	if custom.Light.Accent == custom.Dark.Accent {
		t.Error("las dos variantes salieron iguales")
	}
}

// Un color inválido no debe tumbar el resto del tema.
func TestColorInvalidoNoDescartaLosDemas(t *testing.T) {
	writeTheme(t, `
[dark]
accent = "azul"
text = "#EAEAE8"
danger = "#GG0000"
`)

	custom, errs := LoadTheme()
	if len(errs) != 2 {
		t.Fatalf("se esperaban 2 errores, hubo %d: %v", len(errs), errs)
	}
	for _, err := range errs {
		if !strings.Contains(err.Error(), "[dark]") {
			t.Errorf("el error no dice de qué sección viene: %v", err)
		}
	}

	if custom.Dark.Text != lipgloss.Color("#EAEAE8") {
		t.Error("un color válido se perdió por culpa de otro inválido")
	}
	if custom.Dark.Accent != nil || custom.Dark.Danger != nil {
		t.Error("los colores inválidos deben quedar nulos, no a medias")
	}
}

func TestTomlRotoNoTumbaLaApp(t *testing.T) {
	writeTheme(t, "[dark\naccent = ")

	custom, errs := LoadTheme()
	if len(errs) == 0 {
		t.Fatal("un TOML roto debe avisar")
	}
	if custom.Dark.Accent != nil {
		t.Error("con el fichero roto no debe haber colores")
	}
}

func TestAceptaTresYSeisDigitos(t *testing.T) {
	writeTheme(t, `
[dark]
accent = "#fff"
text = "#EAEAE8"
`)

	custom, errs := LoadTheme()
	if len(errs) != 0 {
		t.Fatalf("errores inesperados: %v", errs)
	}
	if custom.Dark.Accent != lipgloss.Color("#fff") {
		t.Errorf("accent = %v", custom.Dark.Accent)
	}
}

// La ruta debe respetar XDG_CONFIG_HOME.
func TestThemePathRespetaXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-de-prueba")

	path, err := ThemePath()
	if err != nil {
		t.Fatalf("ThemePath: %v", err)
	}
	if want := filepath.Join("/tmp/xdg-de-prueba", appDir, themeFile); path != want {
		t.Errorf("ruta = %q, se esperaba %q", path, want)
	}
}
