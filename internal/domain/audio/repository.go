package audio

import "context"

// Repository es el puerto hacia el servidor de sonido: lo declara el dominio y
// lo implementa infraestructura.
//
// No hay Delete ni alta: los dispositivos de sonido aparecen y desaparecen con
// el hardware, no por decisión de la aplicación.
type Repository interface {
	// List devuelve los dispositivos de una dirección, en el orden que
	// informe el sistema.
	List(ctx context.Context, direction Direction) ([]Device, error)
	// FindByID devuelve ErrNotFound si no existe.
	FindByID(ctx context.Context, id ID) (Device, error)
	// Save aplica volumen y silencio al dispositivo real.
	Save(ctx context.Context, d Device) error
	// SetDefault marca el dispositivo como predeterminado para su dirección.
	SetDefault(ctx context.Context, id ID) error
}

// StreamRepository es el puerto hacia los flujos de las aplicaciones.
//
// Va aparte de Repository porque son colecciones distintas: los dispositivos
// persisten y los flujos aparecen y desaparecen con cada reproducción.
type StreamRepository interface {
	// List devuelve los flujos activos.
	List(ctx context.Context) ([]Stream, error)
	// FindByID devuelve ErrStreamNotFound si el flujo ya terminó.
	FindByID(ctx context.Context, id StreamID) (Stream, error)
	// Save aplica volumen y silencio al flujo real.
	Save(ctx context.Context, s Stream) error
}

// Watcher avisa de cambios hechos fuera de la aplicación: teclas de volumen,
// un mezclador gráfico, conectar unos auriculares.
//
// Es un puerto de entrada: empuja en vez de responder. Sin él la interfaz
// solo se entera de lo que cambia ella misma.
type Watcher interface {
	// Changes entrega un aviso por cada cambio. El canal se cierra cuando el
	// contexto termina.
	//
	// Los avisos no llevan detalle ni se acumulan: varios cambios seguidos
	// pueden llegar como uno solo. Quien lo reciba debe releer el estado, no
	// deducirlo del número de avisos.
	Changes(ctx context.Context) (<-chan struct{}, error)
}
