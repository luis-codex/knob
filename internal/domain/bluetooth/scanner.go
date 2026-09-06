package bluetooth

import "context"

// Scanner descubre dispositivos cercanos.
//
// Es un puerto de salida como Repository, pero hacia el hardware y no hacia
// almacenamiento: devuelve lo que hay alrededor, sin saber qué se conoce ya.
type Scanner interface {
	// Scan devuelve los dispositivos visibles, todos en StateDiscovered.
	Scan(ctx context.Context) ([]Device, error)
}
