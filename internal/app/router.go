package app

import (
	"context"

	"knob/internal/application/prefs"
	"knob/internal/application/sound"
	"knob/internal/domain/preferences"
	"knob/internal/pages"
	"knob/internal/shared/components"
	"knob/internal/shared/layouts"
)

// Navigation IDs: the key that ties the sidebar and the router together.
const (
	navWiFi        = "network.wifi"
	navSound       = "sound.devices"
	navPreferences = "preferences.settings"
)

// navGroups is the sidebar menu.
//
// "Network" (navWiFi) is not listed while the screen is half-done. The page
// stays mounted in newRouter, so to show it again it is enough to put its
// NavGroup back here with the label "Wifi & Network".
func navGroups() []components.NavGroup {
	return []components.NavGroup{
		{
			Label: "Sound",
			Items: []components.NavItem{
				{ID: navSound, Label: "Audio & mic"},
			},
		},
		{
			Label: "Preferences",
			Items: []components.NavItem{
				{ID: navPreferences, Label: "Settings"},
			},
		},
	}
}

// newRouter builds the pages once, so they keep their state when navigating
// away and back.
func newRouter(ctx context.Context, soundUC sound.UseCases, prefsUC prefs.UseCases, initial preferences.Settings) map[string]layouts.Section {
	return map[string]layouts.Section{
		navSound:       pages.NewAudio(ctx, "Audio & mic", soundUC, initial.Audio.VolumeStep.Value(), initial.Interface.Animations),
		navPreferences: pages.NewPreferences(ctx, "Settings", prefsUC, initial),
		navWiFi:        pages.NewWiFi("Wifi & Network"),
	}
}
