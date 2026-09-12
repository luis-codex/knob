package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeSettings drops a config.toml in a temporary config directory.
func writeSettings(t *testing.T, content string) {
	t.Helper()

	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	path := filepath.Join(dir, appDir)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(path, settingsFile), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func TestLoadSettingsNoFileIsDefault(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	got, errs := LoadSettings()
	if len(errs) != 0 {
		t.Fatalf("having no config must not produce errors: %v", errs)
	}
	if got != DefaultSettings() {
		t.Errorf("got %+v, want DefaultSettings() %+v", got, DefaultSettings())
	}
}

func TestLoadSettingsFullFile(t *testing.T) {
	writeSettings(t, `
[audio]
volume_step = 10

# an unknown section must be ignored, not rejected
[theme]
mode = "dark"
`)

	got, errs := LoadSettings()
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if got.VolumeStep != 10 {
		t.Errorf("VolumeStep = %d, want 10", got.VolumeStep)
	}
}

func TestLoadSettingsAbsentKeyKeepsItsDefault(t *testing.T) {
	writeSettings(t, `
[audio]
volume_step = 15
`)

	got, errs := LoadSettings()
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if got.VolumeStep != 15 {
		t.Errorf("VolumeStep = %d, want 15", got.VolumeStep)
	}
}

func TestLoadSettingsBadValueKeepsThatFieldDefaultAndReportsIt(t *testing.T) {
	writeSettings(t, `
[audio]
volume_step = 999
`)

	got, errs := LoadSettings()
	if len(errs) != 1 {
		t.Fatalf("want exactly one error, got %v", errs)
	}
	if !strings.Contains(errs[0].Error(), "volume_step") {
		t.Errorf("error %q does not name the offending key", errs[0])
	}
	if got.VolumeStep != DefaultSettings().VolumeStep {
		t.Errorf("VolumeStep = %d, want the default %d kept", got.VolumeStep, DefaultSettings().VolumeStep)
	}
}

func TestLoadSettingsMalformedFileFallsBackToDefaults(t *testing.T) {
	writeSettings(t, `[audio`)

	got, errs := LoadSettings()
	if len(errs) == 0 {
		t.Fatal("a malformed file must be reported")
	}
	if got != DefaultSettings() {
		t.Errorf("got %+v, want DefaultSettings() so startup is not blocked", got)
	}
}
