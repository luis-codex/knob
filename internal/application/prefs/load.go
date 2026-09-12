package prefs

import (
	"context"

	"knob/internal/domain/preferences"
)

// Load returns the stored preferences.
type Load struct{ deps Deps }

func NewLoad(deps Deps) Load { return Load{deps: deps} }

type LoadCommand struct{}

type LoadResponse struct {
	Settings preferences.Settings
	Rejected []preferences.Rejected
}

func (u Load) Execute(ctx context.Context, _ LoadCommand) (LoadResponse, error) {
	loaded, err := u.deps.Repo.Load(ctx)
	if err != nil {
		return LoadResponse{}, err
	}
	return LoadResponse{Settings: loaded.Settings, Rejected: loaded.Rejected}, nil
}
