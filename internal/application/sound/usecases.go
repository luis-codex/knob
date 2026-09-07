package sound

// UseCases bundles every sound use case so the composition root wires them in
// one place and a page holds a single field.
type UseCases struct {
	ListOutputs        ListOutputs
	ListInputs         ListInputs
	ListStreams        ListStreams
	AdjustVolume       AdjustVolume
	ToggleMuted        ToggleMuted
	MakeDefault        MakeDefault
	AdjustStreamVolume AdjustStreamVolume
	ToggleStreamMuted  ToggleStreamMuted
	WatchChanges       WatchChanges
}

// NewUseCases builds every use case from the shared ports.
func NewUseCases(deps Deps) UseCases {
	return UseCases{
		ListOutputs:        NewListOutputs(deps),
		ListInputs:         NewListInputs(deps),
		ListStreams:        NewListStreams(deps),
		AdjustVolume:       NewAdjustVolume(deps),
		ToggleMuted:        NewToggleMuted(deps),
		MakeDefault:        NewMakeDefault(deps),
		AdjustStreamVolume: NewAdjustStreamVolume(deps),
		ToggleStreamMuted:  NewToggleStreamMuted(deps),
		WatchChanges:       NewWatchChanges(deps),
	}
}
