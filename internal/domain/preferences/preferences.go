// Package preferences holds the user-configurable knobs knob persists --
// audio behaviour, theme and interface -- their value objects and the
// storage port. It depends on no other layer: in particular it must never
// import charm.land/lipgloss or image/color. Rendering those values belongs
// to internal/shared/styles, not here.
package preferences

// Settings is the aggregate: every preference knob persists, grouped by
// concern. It has one instance per user -- there is no collection, no ID.
type Settings struct {
	Audio     Audio
	Theme     Theme
	Interface Interface
}

// Default is what knob uses before any preference is ever saved.
func Default() Settings {
	return Settings{
		Audio:     defaultAudio(),
		Theme:     defaultTheme(),
		Interface: defaultInterface(),
	}
}
