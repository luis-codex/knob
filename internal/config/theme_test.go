package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// writeTheme drops a theme.toml in a temporary config directory.
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

func TestNoFileIsNotAnError(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	custom, errs := LoadTheme()
	if len(errs) != 0 {
		t.Fatalf("having no theme must not produce errors: %v", errs)
	}
	if custom.Dark.Accent != nil || custom.Light.Accent != nil {
		t.Error("with no file there must be no custom colors")
	}
}

func TestPartialOverride(t *testing.T) {
	writeTheme(t, `
# Only the accent; the rest stays at the default.
[dark]
accent = "#516BEB"
`)

	custom, errs := LoadTheme()
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if custom.Dark.Accent != lipgloss.Color("#516BEB") {
		t.Errorf("accent = %v", custom.Dark.Accent)
	}
	if custom.Dark.Text != nil {
		t.Error("an undeclared color must stay nil so Merge uses the default")
	}
	if custom.Light.Accent != nil {
		t.Error("declaring [dark] must not touch [light]")
	}
}

func TestVariantsAreIndependent(t *testing.T) {
	writeTheme(t, `
[light]
accent = "#3B4FC4"

[dark]
accent = "#516BEB"
`)

	custom, errs := LoadTheme()
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if custom.For(true) != custom.Dark || custom.For(false) != custom.Light {
		t.Error("For does not pick the right variant")
	}
	if custom.Light.Accent == custom.Dark.Accent {
		t.Error("the two variants came out identical")
	}
}

// An invalid color must not bring down the rest of the theme.
func TestInvalidColorDoesNotDropTheOthers(t *testing.T) {
	writeTheme(t, `
[dark]
accent = "blue"
text = "#EAEAE8"
danger = "#GG0000"
`)

	custom, errs := LoadTheme()
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(errs), errs)
	}
	for _, err := range errs {
		if !strings.Contains(err.Error(), "[dark]") {
			t.Errorf("the error does not say which section it comes from: %v", err)
		}
	}

	if custom.Dark.Text != lipgloss.Color("#EAEAE8") {
		t.Error("a valid color was lost because of an invalid one")
	}
	if custom.Dark.Accent != nil || custom.Dark.Danger != nil {
		t.Error("invalid colors must stay nil, not half-applied")
	}
}

func TestBrokenTomlDoesNotBringDownTheApp(t *testing.T) {
	writeTheme(t, "[dark\naccent = ")

	custom, errs := LoadTheme()
	if len(errs) == 0 {
		t.Fatal("broken TOML must be reported")
	}
	if custom.Dark.Accent != nil {
		t.Error("with the file broken there must be no colors")
	}
}

func TestAcceptsThreeAndSixDigits(t *testing.T) {
	writeTheme(t, `
[dark]
accent = "#fff"
text = "#EAEAE8"
`)

	custom, errs := LoadTheme()
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if custom.Dark.Accent != lipgloss.Color("#fff") {
		t.Errorf("accent = %v", custom.Dark.Accent)
	}
}

// The path must respect XDG_CONFIG_HOME.
func TestThemePathRespectsXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-test")

	path, err := ThemePath()
	if err != nil {
		t.Fatalf("ThemePath: %v", err)
	}
	if want := filepath.Join("/tmp/xdg-test", appDir, themeFile); path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
}
