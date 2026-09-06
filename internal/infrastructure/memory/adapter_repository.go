package memory

import (
	"context"
	"sync"

	"settings-cli/internal/domain/bluetooth"
)

// AdapterRepository implementa bluetooth.AdapterRepository. Guarda un único
// valor, así que no necesita búsqueda.
type AdapterRepository struct {
	mu      sync.RWMutex
	adapter bluetooth.Adapter
}

var _ bluetooth.AdapterRepository = (*AdapterRepository)(nil)

// NewAdapterRepository arranca con el adaptador en el estado dado.
func NewAdapterRepository(enabled bool) *AdapterRepository {
	return &AdapterRepository{adapter: bluetooth.NewAdapter(enabled)}
}

func (r *AdapterRepository) Get(ctx context.Context) (bluetooth.Adapter, error) {
	if err := ctx.Err(); err != nil {
		return bluetooth.Adapter{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.adapter, nil
}

func (r *AdapterRepository) Save(ctx context.Context, a bluetooth.Adapter) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapter = a
	return nil
}
