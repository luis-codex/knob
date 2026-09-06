package audio

import "context"

// Repository is the port to the sound server: the domain declares it and
// infrastructure implements it.
//
// There is no Delete or create: sound devices appear and disappear with the
// hardware, not by decision of the application.
type Repository interface {
	// List returns the devices for a direction, in whatever order the system
	// reports.
	List(ctx context.Context, direction Direction) ([]Device, error)
	// FindByID returns ErrNotFound if it does not exist.
	FindByID(ctx context.Context, id ID) (Device, error)
	// Save applies volume and mute to the real device.
	Save(ctx context.Context, d Device) error
	// SetDefault marks the device as the default for its direction.
	SetDefault(ctx context.Context, id ID) error
}

// StreamRepository is the port to application streams.
//
// It is separate from Repository because they are different collections:
// devices persist, streams appear and disappear with each playback.
type StreamRepository interface {
	// List returns the active streams.
	List(ctx context.Context) ([]Stream, error)
	// FindByID returns ErrStreamNotFound if the stream has ended.
	FindByID(ctx context.Context, id StreamID) (Stream, error)
	// Save applies volume and mute to the real stream.
	Save(ctx context.Context, s Stream) error
}

// Watcher reports changes made outside the application: volume keys, a
// graphical mixer, plugging in a pair of headphones.
//
// It is an input port: it pushes instead of responding. Without it the
// interface only learns about what it changes itself.
type Watcher interface {
	// Changes delivers one notification per change. The channel closes when
	// the context ends.
	//
	// Notifications carry no detail and do not accumulate: several changes in
	// a row may arrive as one. The receiver must re-read the state, not infer
	// it from the number of notifications.
	Changes(ctx context.Context) (<-chan struct{}, error)
}
