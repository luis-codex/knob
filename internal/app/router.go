package app

import (
	"context"

	"knob/internal/application/devices"
	"knob/internal/application/sound"
	"knob/internal/pages"
	"knob/internal/shared/components"
	"knob/internal/shared/layouts"
)

// Navigation IDs: the key that ties the sidebar and the router together.
const (
	navWiFi      = "network.wifi"
	navBluetooth = "network.bluetooth"
	navSound     = "sound.devices"
)

// navGroups is the sidebar menu.
//
// "Network & Bluetooth" (navWiFi and navBluetooth) is not listed while both
// screens are half-done. The pages stay mounted in newRouter, so to show them
// again it is enough to put their NavGroup back here with those two IDs and the
// labels "Wifi & Network" and "Bluetooth".
func navGroups() []components.NavGroup {
	return []components.NavGroup{
		{
			Label: "Sound",
			Items: []components.NavItem{
				{ID: navSound, Label: "Audio & mic"},
			},
		},
	}
}

// newRouter builds the pages once, so they keep their state when navigating
// away and back.
func newRouter(ctx context.Context, deviceUC devices.UseCases, soundUC sound.UseCases, volumeStep int) map[string]layouts.Section {
	return map[string]layouts.Section{
		navSound:     pages.NewAudio(ctx, "Audio & mic", soundUC, volumeStep),
		navWiFi:      pages.NewWiFi("Wifi & Network"),
		navBluetooth: pages.NewBluetooth("Bluetooth", deviceUC),
	}
}
