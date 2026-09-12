// Package prefs holds the preferences use cases: Load and Save, each with a
// single Execute method, sharing the injected port through Deps.
//
// It is named prefs rather than preferences so it does not clash with the
// domain package.
package prefs

import "knob/internal/domain/preferences"

// Deps are the ports every use case in this package needs. The composition
// root fills it once and NewUseCases hands it to each use case.
type Deps struct {
	Repo preferences.Repository
}
