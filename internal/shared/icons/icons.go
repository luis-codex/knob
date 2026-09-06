// Package icons reúne los símbolos de la interfaz en un solo sitio.
//
// Son el juego por defecto: el tema los recoge en styles.Icons, así que
// cambiarlos para un terminal concreto no obliga a tocar este paquete.
//
// Todos deben ocupar una celda salvo los que se documenten como dobles: la
// rejilla de las filas cuenta con ello.
package icons

// Marcadores de estado en una lista.
const (
	// Cursor señala la fila bajo el cursor.
	Cursor = "›"
	// Active señala el elemento en uso: el dispositivo predeterminado.
	Active = "●"
	// Paused señala un flujo de audio detenido.
	Paused = "⏸"
	// Playing señala un flujo que suena. Parpadea, así que la pantalla lo
	// enciende y lo apaga.
	Playing = "●"
)

// Trazos del layout y de las barras.
const (
	BarOn       = "▰"
	BarOff      = "·"
	DividerH    = "─"
	DividerV    = "│"
	ScrollThumb = "┃"
	ScrollTrack = "│"
	Ellipsis    = "…"
)

// Teclas. Se escriben con el símbolo cuando existe y con su nombre cuando no:
// "esc" se reconoce mejor que cualquier glifo.
const (
	// Enter ocupa una celda.
	Enter = "↵"
	// UpDown y LeftRight ocupan dos.
	UpDown    = "↑↓"
	LeftRight = "←→"
	Tab       = "⇥"
	Escape    = "esc"
	Space     = "espacio"
	Delete    = "supr"
)

// Separator va entre una pista de teclado y la siguiente.
const Separator = "·"
