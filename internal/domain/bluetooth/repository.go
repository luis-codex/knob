package bluetooth

import "context"

// Repository es el puerto de persistencia: lo declara el dominio y lo
// implementa infraestructura.
type Repository interface {
	// Save inserta o actualiza el dispositivo.
	Save(ctx context.Context, d Device) error
	// Delete elimina por dirección. Devuelve ErrNotFound si no existe.
	Delete(ctx context.Context, address Address) error
	// FindByAddress devuelve ErrNotFound si no existe.
	FindByAddress(ctx context.Context, address Address) (Device, error)
	// List devuelve los dispositivos en orden de descubrimiento.
	List(ctx context.Context) ([]Device, error)
}

// AdapterRepository guarda el estado de la radio. Es un puerto aparte porque
// el adaptador es otro agregado: no comparte ciclo de vida con los
// dispositivos.
type AdapterRepository interface {
	// Get devuelve el adaptador. Nunca falla por no existir: si no hay nada
	// guardado devuelve uno apagado.
	Get(ctx context.Context) (Adapter, error)
	Save(ctx context.Context, a Adapter) error
}
