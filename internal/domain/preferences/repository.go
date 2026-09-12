package preferences

import "context"

// Repository is the port to the preferences store: the domain declares it,
// infrastructure implements it. Swapping the store -- a TOML file today,
// something else later -- is a new adapter here, nothing more.
type Repository interface {
	// Load returns the stored preferences. A value the store holds but the
	// domain rejects is replaced by its default and reported in
	// Loaded.Rejected; the error is reserved for not being able to read the
	// store at all -- an unreadable directory, or a file that is not valid
	// TOML.
	Load(ctx context.Context) (Loaded, error)
	// Save persists s, replacing whatever was stored before.
	Save(ctx context.Context, s Settings) error
}

// Rejected is a stored value the domain refused, and why.
type Rejected struct {
	Key    string
	Reason string
}

// Loaded is the result of Load: the usable settings, plus what was dropped
// along the way.
type Loaded struct {
	Settings Settings
	Rejected []Rejected
}
