package audio

import "strings"

// StreamID identifies a stream. It is the index the server assigns: it changes
// between runs and is no use for remembering anything from one session to the
// next.
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

// Stream is an application's audio stream: what is playing and whose it is.
//
// Unlike Device it is ephemeral -- it is born and dies with the playback --
// and cannot be the default: nobody picks "the default stream".
type Stream struct {
	id     StreamID
	app    Name
	title  string
	volume Volume
	muted  bool
	paused bool
}

// RestoreStream rebuilds a stream exactly as the server reports it.
//
// title is opaque system text (the track, the file) and may come empty or
// very long: it is not validated, whoever renders it trims it.
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

// Paused reports that the application has playback stopped.
func (s Stream) Paused() bool { return s.paused }

func (s Stream) IsZero() bool { return s.id.IsZero() }

func (s Stream) SetVolume(v Volume) Stream {
	s.volume = v
	return s
}

// AdjustVolume adds delta, clamping at the ends.
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
