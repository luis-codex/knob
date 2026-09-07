package sound

import "context"

// WatchChanges reports changes made outside the application: volume keys, a
// graphical mixer, plugging in a pair of headphones.
type WatchChanges struct{ deps Deps }

func NewWatchChanges(deps Deps) WatchChanges { return WatchChanges{deps: deps} }

type WatchChangesCommand struct{}

type WatchChangesResponse struct{ Changes <-chan struct{} }

func (u WatchChanges) Execute(ctx context.Context, _ WatchChangesCommand) (WatchChangesResponse, error) {
	changes, err := u.deps.Watcher.Changes(ctx)
	if err != nil {
		return WatchChangesResponse{}, err
	}
	return WatchChangesResponse{Changes: changes}, nil
}
