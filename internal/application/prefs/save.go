package prefs

import (
	"context"

	"knob/internal/domain/preferences"
)

// Save persists the given preferences, replacing whatever was stored before.
type Save struct{ deps Deps }

func NewSave(deps Deps) Save { return Save{deps: deps} }

type SaveCommand struct{ Settings preferences.Settings }

type SaveResponse struct{ Settings preferences.Settings }

func (u Save) Execute(ctx context.Context, cmd SaveCommand) (SaveResponse, error) {
	if err := u.deps.Repo.Save(ctx, cmd.Settings); err != nil {
		return SaveResponse{}, err
	}
	return SaveResponse(cmd), nil
}
