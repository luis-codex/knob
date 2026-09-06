package app

import (
	"context"

	"settings-cli/internal/application/devices"
	"settings-cli/internal/application/sound"
	"settings-cli/internal/pages"
	"settings-cli/internal/shared/components"
	"settings-cli/internal/shared/layouts"
)

// IDs de navegación: la clave que une sidebar y router.
const (
	navWiFi      = "network.wifi"
	navBluetooth = "network.bluetooth"
	navSound     = "sound.devices"
)

// navGroups es el menú del sidebar.
//
// "Network & Bluetooth" (navWiFi y navBluetooth) no se lista mientras las dos
// pantallas estén a medias. Las páginas siguen montadas en newRouter, así que
// para volver a enseñarlas basta con reponer aquí su NavGroup con esos dos IDs
// y las etiquetas "Wifi & Network" y "Bluetooth".
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

// newRouter construye las páginas una vez, de modo que conservan su estado
// al navegar fuera y volver.
func newRouter(ctx context.Context, deviceSvc *devices.Service, soundSvc *sound.Service) map[string]layouts.Section {
	return map[string]layouts.Section{
		navSound:     pages.NewAudio(ctx, "Audio & mic", soundSvc),
		navWiFi:      pages.NewWiFi("Wifi & Network"),
		navBluetooth: pages.NewBluetooth("Bluetooth", deviceSvc),
	}
}
