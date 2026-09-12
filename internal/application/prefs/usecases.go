package prefs

// UseCases bundles every preferences use case so the composition root wires
// them in one place and a page holds a single field.
type UseCases struct {
	Load Load
	Save Save
}

// NewUseCases builds every use case from the shared ports.
func NewUseCases(deps Deps) UseCases {
	return UseCases{
		Load: NewLoad(deps),
		Save: NewSave(deps),
	}
}
