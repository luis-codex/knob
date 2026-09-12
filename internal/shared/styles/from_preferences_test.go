package styles_test

import (
	"testing"

	"charm.land/lipgloss/v2"

	"knob/internal/domain/preferences"
	"knob/internal/shared/styles"
)

func TestFromPreferencesUnsetStaysNil(t *testing.T) {
	custom := styles.FromPreferences(preferences.Theme{})
	if custom.Dark.Accent != nil || custom.Light.Accent != nil {
		t.Error("an unset domain color must map to nil, so Merge uses the default")
	}
}

func TestFromPreferencesConvertsSetColors(t *testing.T) {
	accent, err := preferences.NewColor("#516BEB")
	if err != nil {
		t.Fatalf("NewColor: %v", err)
	}

	theme := preferences.Theme{Dark: preferences.Palette{Accent: accent}}
	custom := styles.FromPreferences(theme)

	if custom.Dark.Accent != lipgloss.Color("#516BEB") {
		t.Errorf("Dark.Accent = %v", custom.Dark.Accent)
	}
	if custom.Light.Accent != nil {
		t.Error("setting Dark must not touch Light")
	}
}
