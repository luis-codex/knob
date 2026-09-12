package tomlstore_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"knob/internal/domain/preferences"
	"knob/internal/infrastructure/tomlstore"
)

// setDir points the store at a fresh, temporary config directory.
func setDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	return dir
}

// writeRaw drops content at the store's path, creating its parent.
func writeRaw(t *testing.T, content string) string {
	t.Helper()

	path, err := tomlstore.Path()
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

func TestPathRespectsXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-test")

	path, err := tomlstore.Path()
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	if want := "/tmp/xdg-test/knob/config.toml"; path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
}

func TestLoadNoFileWritesDefaults(t *testing.T) {
	setDir(t)

	loaded, err := tomlstore.NewRepository().Load(context.Background())
	if err != nil {
		t.Fatalf("having no config must not produce an error: %v", err)
	}
	if len(loaded.Rejected) != 0 {
		t.Fatalf("unexpected rejections: %v", loaded.Rejected)
	}
	if loaded.Settings != preferences.Default() {
		t.Errorf("got %+v, want the defaults", loaded.Settings)
	}

	path, _ := tomlstore.Path()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("the store must write the defaults on first run: %v", err)
	}
}

func TestSaveThenLoadRoundTrip(t *testing.T) {
	setDir(t)
	repo := tomlstore.NewRepository()
	ctx := context.Background()

	step, err := preferences.NewVolumeStep(20)
	if err != nil {
		t.Fatalf("NewVolumeStep: %v", err)
	}
	mode, err := preferences.NewMode(preferences.ModeDark)
	if err != nil {
		t.Fatalf("NewMode: %v", err)
	}
	accent, err := preferences.NewColor("#516BEB")
	if err != nil {
		t.Fatalf("NewColor: %v", err)
	}

	want := preferences.Default()
	want.Audio.VolumeStep = step
	want.Theme.Mode = mode
	want.Theme.Dark.Accent = accent
	want.Interface.SidebarHidden = true
	want.Interface.Animations = false

	if err := repo.Save(ctx, want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := repo.Load(ctx)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Rejected) != 0 {
		t.Fatalf("unexpected rejections: %v", loaded.Rejected)
	}
	if loaded.Settings != want {
		t.Errorf("got %+v, want %+v", loaded.Settings, want)
	}
}

func TestLoadAbsentKeyKeepsItsDefault(t *testing.T) {
	setDir(t)
	writeRaw(t, `
[audio]
volume_step = 15
`)

	loaded, err := tomlstore.NewRepository().Load(context.Background())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Rejected) != 0 {
		t.Fatalf("unexpected rejections: %v", loaded.Rejected)
	}
	if loaded.Settings.Audio.VolumeStep.Value() != 15 {
		t.Errorf("VolumeStep = %d, want 15", loaded.Settings.Audio.VolumeStep.Value())
	}
	if !loaded.Settings.Interface.Animations {
		t.Error("an absent boolean key must keep its default (true), not fall back to the zero value")
	}
	if loaded.Settings.Interface.SidebarHidden {
		t.Error("an absent boolean key must keep its default (false)")
	}
	if !loaded.Settings.Theme.Mode.IsAuto() {
		t.Errorf("mode = %q, want the default auto", loaded.Settings.Theme.Mode.String())
	}
}

func TestLoadBadValuesKeepDefaultsAndAreReported(t *testing.T) {
	setDir(t)
	writeRaw(t, `
[audio]
volume_step = 999

[theme]
mode = "system"

[theme.dark]
accent = "blue"
text = "#EAEAE8"
`)

	loaded, err := tomlstore.NewRepository().Load(context.Background())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	wantKeys := map[string]bool{
		"audio.volume_step": false,
		"theme.mode":        false,
		"theme.dark.accent": false,
	}
	if len(loaded.Rejected) != len(wantKeys) {
		t.Fatalf("want %d rejections, got %d: %+v", len(wantKeys), len(loaded.Rejected), loaded.Rejected)
	}
	for _, r := range loaded.Rejected {
		if _, ok := wantKeys[r.Key]; !ok {
			t.Errorf("unexpected rejected key %q", r.Key)
		}
		wantKeys[r.Key] = true
	}
	for key, seen := range wantKeys {
		if !seen {
			t.Errorf("missing rejection for %q", key)
		}
	}

	if loaded.Settings.Audio.VolumeStep.Value() != preferences.DefaultVolumeStep {
		t.Errorf("VolumeStep = %d, want the default kept", loaded.Settings.Audio.VolumeStep.Value())
	}
	if !loaded.Settings.Theme.Mode.IsAuto() {
		t.Error("an invalid mode must fall back to auto")
	}
	if !loaded.Settings.Theme.Dark.Accent.IsZero() {
		t.Error("an invalid color must stay unset, not half-applied")
	}
	if loaded.Settings.Theme.Dark.Text.String() != "#EAEAE8" {
		t.Error("a valid color must not be lost because a sibling was invalid")
	}
}

func TestLoadMalformedFileReturnsErrorAndDefaults(t *testing.T) {
	setDir(t)
	writeRaw(t, "[audio")

	loaded, err := tomlstore.NewRepository().Load(context.Background())
	if err == nil {
		t.Fatal("malformed TOML must be reported as an error, not silently swallowed")
	}
	if loaded.Settings != preferences.Default() {
		t.Errorf("got %+v, want the defaults so startup is not blocked", loaded.Settings)
	}
}

// This is the data-loss guard: a syntax error must never be clobbered by a
// save, or the user's in-progress fix is destroyed.
func TestSaveRefusesOverAFileThatDoesNotParse(t *testing.T) {
	setDir(t)
	path := writeRaw(t, "[audio")

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	err = tomlstore.NewRepository().Save(context.Background(), preferences.Default())
	if err == nil {
		t.Fatal("Save must refuse to write over a file it could not parse")
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(after) != string(before) {
		t.Error("the broken file was overwritten; the user's data is gone")
	}
}
