package audio

import "strings"

// StreamID identifica un flujo. Es el índice que asigna el servidor: cambia
// entre ejecuciones y no sirve para recordar nada de una sesión a otra.
type StreamID struct {
	value int
	set   bool
}

func NewStreamID(index int) (StreamID, error) {
	if index < 0 {
		return StreamID{}, ErrInvalidStreamID
	}
	return StreamID{value: index, set: true}, nil
}

func (i StreamID) Index() int { return i.value }

func (i StreamID) IsZero() bool { return !i.set }

func (i StreamID) Equals(other StreamID) bool { return i.set == other.set && i.value == other.value }

// Stream es un flujo de audio de una aplicación: lo que suena y de quién.
//
// A diferencia de Device es efímero —nace y muere con la reproducción— y no
// puede ser predeterminado: nadie elige "el flujo por defecto".
type Stream struct {
	id     StreamID
	app    Name
	title  string
	volume Volume
	muted  bool
	paused bool
}

// RestoreStream reconstruye un flujo tal y como lo informa el servidor.
//
// title es texto opaco del sistema (la pista, el fichero) y puede venir vacío
// o larguísimo: no se valida, lo recorta quien lo pinte.
func RestoreStream(id StreamID, app Name, title string, volume Volume, muted, paused bool) (Stream, error) {
	switch {
	case id.IsZero():
		return Stream{}, ErrInvalidStreamID
	case app.IsZero():
		return Stream{}, ErrEmptyName
	}

	return Stream{
		id:     id,
		app:    app,
		title:  strings.TrimSpace(title),
		volume: volume,
		muted:  muted,
		paused: paused,
	}, nil
}

func (s Stream) ID() StreamID { return s.id }

func (s Stream) App() Name { return s.app }

func (s Stream) Title() string { return s.title }

func (s Stream) Volume() Volume { return s.volume }

func (s Stream) Muted() bool { return s.muted }

// Paused indica que la aplicación tiene la reproducción detenida.
func (s Stream) Paused() bool { return s.paused }

func (s Stream) IsZero() bool { return s.id.IsZero() }

func (s Stream) SetVolume(v Volume) Stream {
	s.volume = v
	return s
}

// AdjustVolume suma delta acotando en los extremos.
func (s Stream) AdjustVolume(delta int) Stream {
	return s.SetVolume(ClampVolume(s.volume.Level() + delta))
}

func (s Stream) SetMuted(muted bool) Stream {
	s.muted = muted
	return s
}

func (s Stream) ToggleMuted() Stream {
	return s.SetMuted(!s.muted)
}
